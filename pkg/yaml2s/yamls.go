package yaml2s

import (
	"fmt"
	"os"

	"github.com/jenkins-x/jx-helpers/v3/pkg/files"
	"gopkg.in/yaml.v2"
)

// LoadFile loads the given YAML file using the gopkg.in/yaml.v2 library
func LoadFile(fileName string, dest interface{}) error {
	exists, err := files.FileExists(fileName)
	if err != nil {
		return fmt.Errorf("failed to check if file exists  %s: %w", fileName, err)
	}
	if !exists {
		return nil
	}

	data, err := os.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", fileName, err)
	}

	err = yaml.Unmarshal(data, dest)
	if err != nil {
		return fmt.Errorf("failed to unmarshal file %s: %w", fileName, err)
	}
	return nil
}

// SaveFile saves the object using the gopkg.in/yaml.v2 library the given file name
func SaveFile(obj interface{}, fileName string) error {
	data, err := yaml.Marshal(obj)
	if err != nil {
		return fmt.Errorf("failed to marshal to YAML: %w", err)
	}
	err = os.WriteFile(fileName, data, files.DefaultFileWritePermissions)
	if err != nil {
		return fmt.Errorf("failed to save file %s: %w", fileName, err)
	}
	return nil
}
