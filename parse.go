package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"golang.org/x/xerrors"
	"gopkg.in/yaml.v2"
)

type NoExistFileError struct {
	filepath string
}

func (e *NoExistFileError) Error() string {
	return fmt.Sprintf("%s does not exist\n", e.filepath)
}

func DeserializeYamlFile(filepath string, out interface{}) error {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return xerrors.Errorf("%w", &NoExistFileError{
			filepath: filepath,
		})
	}

	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return xerrors.Errorf("%w", err)
	}

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	return decoder.Decode(out)
}
