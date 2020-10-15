package main

import (
	"encoding/json"
	"io/ioutil"
	"testing"
)

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
