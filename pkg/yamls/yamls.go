package yamls

import (
	"io/ioutil"

	"github.com/jenkins-x/jx-helpers/v3/pkg/files"
	"sigs.k8s.io/yaml"

	"github.com/pkg/errors"
)

// LoadFile loads the given YAML file
func LoadFile(fileName string, dest interface{}) error {
	exists, err := files.FileExists(fileName)
	if err != nil {
		return errors.Wrapf(err, "failed to check if file exists  %s", fileName)
	}
	if !exists {
		return nil
	}

	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return errors.Wrapf(err, "failed to read file %s", fileName)
	}

	err = yaml.Unmarshal(data, dest)
	if err != nil {
		return errors.Wrapf(err, "failed to unmarshal file %s", fileName)
	}
	return nil
}

// SaveFile saves the object as the given file name
func SaveFile(obj interface{}, fileName string) error {
	data, err := yaml.Marshal(obj)
	if err != nil {
		return errors.Wrap(err, "failed to marshal to YAML")
	}
	err = ioutil.WriteFile(fileName, data, files.DefaultFileWritePermissions)
	if err != nil {
		return errors.Wrapf(err, "failed to save file %s", fileName)
	}
	return nil
}
