package main

import (
	"fmt"
	vault "github.com/hashicorp/vault/api"
	log "github.com/sirupsen/logrus"
	"strings"
)

type VaultWrapper struct {
	client *vault.Logical
	prefix string
}

func NewVaultWrapper(client *vault.Logical, prefix string) *VaultWrapper {
	vaultWrapper := VaultWrapper{
		client: client,
		prefix: prefix,
	}

	return &vaultWrapper
}

func (vr *VaultWrapper) GetSecretKeys(path string) ([]string, error) {
	lookupKey := fmt.Sprintf("%s%s", vr.prefix, path)
	resp, err := vr.client.Read(lookupKey)
	if err != nil {
		return []string{}, fmt.Errorf("Error reading secret from vault: %s", err.Error())
	}

	if resp == nil {
		return []string{}, fmt.Errorf("Error no secrets found in '%s'", path)
	}
	return vr.keys(resp.Data), nil
}

func (vr *VaultWrapper) GetList(path string) ([]string, error) {
	lookupKey := fmt.Sprintf("%s%s", vr.prefix, path)
	resp, err := vr.client.List(lookupKey)

	if err != nil {
		return []string{}, fmt.Errorf("Error reading paths from vault: %s", err.Error())
	}

	if resp == nil {
		return []string{}, fmt.Errorf("No paths found in '%s'", path)
	}
	resp_keys := resp.Data["keys"]

	_, ok := resp_keys.([]interface{})
	toReturn := make([]string, 0, len(resp_keys.([]interface{})))
	if ok {
		for _, v := range resp_keys.([]interface{}) {
			toReturn = append(toReturn, v.(string))
		}
	} else {
		log.Fatal("resp_keys is not slice")
	}
	return toReturn, nil
}

func (vr *VaultWrapper) keys(m map[string]interface{}) []string { // Helper method
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func (vr *VaultWrapper) GetSecret(path string, secretName string) (string, error) {
	lookupKey := fmt.Sprintf("%s%s", vr.prefix, path)
	resp, err := vr.client.Read(lookupKey)

	if err != nil {
		return "", fmt.Errorf("Error reading secret from vault: %s", err.Error())
	}

	if resp == nil {
		return "", fmt.Errorf("Error no secrets found in '%s'", path)
	}

	secretInterface, ok := resp.Data[secretName]
	if !ok {
		secretKeys := vr.keys(resp.Data)
		if len(secretKeys) == 0 {
			return "", fmt.Errorf("Error no secrets found in '%s'", path)
		}
		return "", fmt.Errorf(
			"Error secret %s!%s not found, available secrets in %s: %s",
			path,
			secretName,
			path,
			strings.Join(secretKeys, ","),
		)
	}

	return secretInterface.(string), nil
}

func (vr *VaultWrapper) GetAllPaths() ([]string, error) {
	rootList, _ := vr.GetList("/")
	pathList := make([]string, 0)
	if len(rootList) == 0 {
		return []string{"/"}, nil
	} else {
		childList := make([]string, 0)
		temp_list := rootList
		childrenExist := true
		currentPath := "/"

		for childrenExist == true {

			for _, value := range temp_list {
				currentValueWithPath := ""
				if value[0:1] == "/" {
					currentValueWithPath = value
				} else {
					currentValueWithPath = fmt.Sprintf("%s%s", currentPath, value)
				}
				if !strings.HasSuffix(currentValueWithPath, "/") {
					slashValueExists := false
					for _, slashValue := range temp_list {
						if fmt.Sprintf("%s/", value) == slashValue {
							slashValueExists = true
						}
					}
					if slashValueExists {
						continue
					}
				}
				pathList = append(pathList, currentValueWithPath)
				currentChildren, _ := vr.GetList(currentValueWithPath)
				if len(currentChildren) > 0 {
					for _, childValue := range currentChildren {
						childList = append(childList, fmt.Sprintf("%s%s", currentValueWithPath, childValue))
					}
				}
			}

			if len(childList) == 0 {
				childrenExist = false
			} else {
				temp_list = childList
				childList = []string{}
			}
		}
	}
	return pathList, nil
}
