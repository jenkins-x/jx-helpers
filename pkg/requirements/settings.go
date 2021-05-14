package requirements

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	jxcore "github.com/jenkins-x/jx-api/v4/pkg/apis/core/v4beta1"
	"github.com/jenkins-x/jx-api/v4/pkg/util"
	"github.com/jenkins-x/jx-helpers/v3/pkg/files"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"
)

// LoadSettings loads the settings from the given directory if present or return nil
func LoadSettings(dir string, failOnValidationErrors bool) (*jxcore.Settings, error) {
	config := &jxcore.Settings{}
	path := filepath.Join(dir, ".jx", jxcore.SettingsFileName)

	exists, err := files.FileExists(path)
	if err != nil {
		return config, errors.Wrapf(err, "failed to check if file exists %s", path)
	}
	if !exists {
		return nil, nil
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return config, errors.Wrapf(err, "failed to read %s", path)
	}

	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML file %s due to %s", path, err)
	}

	validationErrors, err := util.ValidateYaml(config, data)
	if err != nil {
		return nil, fmt.Errorf("failed to validate YAML file %s due to %s", path, err)
	}

	if len(validationErrors) > 0 {
		log.Logger().Warnf("validation failures in YAML file %s: %s", path, strings.Join(validationErrors, ", "))
		if failOnValidationErrors {
			return nil, fmt.Errorf("validation failures in YAML file %s:\n%s", path, strings.Join(validationErrors, "\n"))
		}
	}
	return config, nil
}
