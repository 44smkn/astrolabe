package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

func DeserializeYamlFile(filepath string, out interface{}) error {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return fmt.Errorf("%s does not exist\n", filepath)
	}

	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read file. %v\n", err)
	}

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	return decoder.Decode(out)
}
