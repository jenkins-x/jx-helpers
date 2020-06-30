package homedir

import (
	"os"
	"path/filepath"

	"github.com/jenkins-x/jx-helpers/pkg/files"
)

func HomeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	h := os.Getenv("USERPROFILE") // windows
	if h == "" {
		h = "."
	}
	return h
}

// ConfigDir passes in the env var for a home dir if defined or the default dir inside the home dir to use if not
func ConfigDir(envVar string, defaultDirName string) (string, error) {
	if envVar != "" {
		return envVar, nil
	}
	h := HomeDir()
	path := filepath.Join(h, defaultDirName)
	err := os.MkdirAll(path, files.DefaultDirWritePermissions)
	if err != nil {
		return "", err
	}
	return path, nil
}

// PluginBinDir returns the plugin directory
func PluginBinDir(envVar string, defaultDirName string) (string, error) {
	configDir, err := ConfigDir(envVar, defaultDirName)
	if err != nil {
		return "", err
	}
	path := filepath.Join(configDir, "plugins")
	err = os.MkdirAll(path, files.DefaultDirWritePermissions)
	if err != nil {
		return "", err
	}
	return path, nil
}
