package main

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/http"
	"github.com/hashicorp/vault/vault"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net"
	"testing"
)

var prefix = "/secret"
var vaultTestSecrets = []string{"test1", "test2", "subdir/test", "subdir"}

func SliceContains(input []string, check string) bool {
	for _, v := range input {
		if v == check {
			return true
		} else {
		}
	}
	return false
}

func TestVaultFunctions(t *testing.T) {
	ln, client := createTestVault(t)
	defer ln.Close()
	vault := NewVaultWrapper(client.Logical(), prefix)
	paths, err := vault.GetAllPaths()
	if err != nil {
		t.Errorf("GetAllPaths() returned error: %s", err.Error())
	}
	VaultGetAllPathsTest(t, vault, paths)
	VaultGetSecretKeysTest(t, vault, paths)
	VaultGetSecretTest(t, vault, paths)
}

func VaultGetAllPathsTest(t *testing.T, vault *VaultWrapper, paths []string) {
	log.Info("Testing GetAllPaths()...")
	if !SliceContains(paths, "/subdir/") {
		t.Errorf("GetAllPaths() returned incorrect data: /subdir/ not in %s", paths)
	}
	if !SliceContains(paths, "/test1") {
		t.Errorf("GetAllPaths() returned incorrect data: /test1 not in %s", paths)
	}
	if !SliceContains(paths, "/test2") {
		t.Errorf("GetAllPaths() returned incorrect data: /test2 not in %s", paths)
	}
	if !SliceContains(paths, "/subdir/test") {
		t.Errorf("GetAllPaths() returned incorrect data: /subdir/test not in %s", paths)
	}
}

func VaultGetSecretKeysTest(t *testing.T, vault *VaultWrapper, paths []string) {
	log.Info("Testing GetSecretKeys()...")
	for _, v := range paths {
		secretKeys, _ := vault.GetSecretKeys(v)
		if !SliceContains(secretKeys, "foo") {
			t.Errorf("GetSecretKeys() returned incorrect data: %s", secretKeys)
		}
		if !SliceContains(secretKeys, "hello") {
			t.Errorf("GetSecretKeys() returned incorrect data: %s", secretKeys)
		}
	}
}

func VaultGetSecretTest(t *testing.T, vault *VaultWrapper, paths []string) {
	log.Info("Testing GetSecret()...")

	for _, v := range paths {
		secretVal1, _ := vault.GetSecret(v, "hello")
		if secretVal1 != "world" {
			t.Errorf("GetSecret() returned incorrect data: %s", secretVal1)
		}
		secretVal2, _ := vault.GetSecret(v, "foo")
		if secretVal2 != "bar" {
			t.Errorf("GetSecret() returned incorrect data: %s", secretVal2)
		}
	}
}

func createTestVault(t *testing.T) (net.Listener, *api.Client) {
	t.Helper()

	// Create an in-memory, unsealed core (the "backend", if you will).
	core, keyShares, rootToken := vault.TestCoreUnsealed(t)
	_ = keyShares

	// Start an HTTP server for the core.
	ln, addr := http.TestServer(t, core)

	// Create a client that talks to the server, initially authenticating with
	// the root token.
	conf := api.DefaultConfig()
	conf.Address = addr

	client, err := api.NewClient(conf)
	if err != nil {
		t.Fatal(err)
	}
	client.SetToken(rootToken)

	// Setup required secrets, policies, etc.
	for _, v := range vaultTestSecrets {
		_, err = client.Logical().Write(fmt.Sprintf("%s/%s", prefix, v), map[string]interface{}{
			"hello": "world",
			"foo":   "bar",
		})
	}

	if err != nil {
		t.Fatal(err)
	}

	return ln, client
}

func TestWriteFile(t *testing.T) {
	var testData []interface{}
	testMap := make(map[string]string)
	testMap["foo"] = "bar"
	testMap["hello"] = "world"
	testData = append(testData,
		"foo",
		testMap,
		[]string{"foo", "bar"},
	)
	testFile := "testfile.txt"

	for _, v := range testData {
		WriteFile(v, testFile)
		data, _ := ioutil.ReadFile(testFile)
		jsonData, _ := json.Marshal(v)
		if string(data) != string(jsonData) {
			t.Errorf("WriteFile() wrote incorrect data, expected: %s, got: %s", string(jsonData), string(data))
		}
	}
}
