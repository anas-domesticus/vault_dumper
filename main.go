package main

import (
	"encoding/json"
	"flag"
	"fmt"
	vault "github.com/hashicorp/vault/api"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strings"
	"time"
)

// Env vars...
var vaultAddr = os.Getenv("VAULTURL") + ":" + os.Getenv("VAULTPORT")
var secretEngine = os.Getenv("SECRETENGINE")
var vaultToken = os.Getenv("VAULTTOKEN")

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

func GetEnvVars() {
	fatalError := false
	for _, v := range []string{
		"VAULTURL",
		"VAULTPORT",
		"SECRETENGINE",
		"VAULTTOKEN",
	} {
		_, varPresent := os.LookupEnv(v)
		if !(varPresent) {
			log.Error(fmt.Sprintf("Missing APIKEY environment variable: %s", v))
			fatalError = true
		}
	}
	if fatalError == true {
		os.Exit(1)
	}
}

func main() {
	initialiseLogger()
	GetEnvVars()
	flag.Parse()

	log.Info("Starting...")
	log.Info("Got vault URL as: " + vaultAddr)

	client, err := vault.NewClient(&vault.Config{Address: vaultAddr, HttpClient: httpClient})
	if err != nil {
		log.Fatal("Error occurred connecting to Vault: " + err.Error())
	}
	log.Info("Setting token...")
	client.SetToken(vaultToken)
	vault := NewVaultWrapper(client.Logical(), secretEngine)

	log.Info("Reading secrets from secret engine: " + secretEngine)

	secretPaths, err := vault.GetAllPaths()
	if err != nil {
		log.Fatal("Error in GetAllPaths(): " + err.Error())
	}
	log.Info("Found " + fmt.Sprintf("%d", len(secretPaths)) + " paths in secret engine...")

	outMap := make(map[string]map[string]string)

	for _, currSecret := range secretPaths {
		log.Info(fmt.Sprintf("Reading: %s...", currSecret))
		secretKeys, _ := vault.GetSecretKeys(currSecret)
		log.Info(fmt.Sprintf("Found %s keys in %s", fmt.Sprintf("%d", len(secretKeys)), currSecret))

		for _, keyName := range secretKeys {
			correctedSecret := currSecret
			if strings.HasSuffix(currSecret, "/") {
				correctedSecret = strings.TrimSuffix(currSecret, "/")
			}
			secretVal, _ := vault.GetSecret(correctedSecret, keyName)
			log.Info(fmt.Sprintf("Secret: %s:%s, value: %s", correctedSecret, keyName, "***"))
			serviceName := ""
			if strings.Contains(correctedSecret[1:len(correctedSecret)], "/") {
				i := strings.Index(correctedSecret[1:len(correctedSecret)], "/")
				serviceName = correctedSecret[1 : i+1]
			} else {
				serviceName = correctedSecret[1:len(correctedSecret)]
			}
			if _, ok := outMap[serviceName]; !ok {
				outMap[serviceName] = make(map[string]string)
			}
			outMap[serviceName][BuildJSONKey(secretEngine, correctedSecret, keyName)] = secretVal
		}
	}
	WriteFile(outMap, "output.json")
}

func BuildJSONKey(secretEngine, secret, keyName string) string {
	return fmt.Sprintf("%s%s:%s", strings.TrimPrefix(secretEngine, "/"), secret, keyName)
}

func WriteFile(data interface{}, fileName string) {
	f, err := os.Create(fileName)
	if err != nil {
		log.Error("Error writing file: ", err.Error())
	}
	defer f.Close()
	output, _ := json.Marshal(data)
	f.Write(output)
}
