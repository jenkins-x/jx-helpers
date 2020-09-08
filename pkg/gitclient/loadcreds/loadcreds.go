package loadcreds

import (
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/jenkins-x/jx-helpers/pkg/files"
	"github.com/jenkins-x/jx-helpers/pkg/homedir"
	"github.com/jenkins-x/jx-helpers/pkg/termcolor"
	"github.com/jenkins-x/jx-logging/pkg/log"
	"github.com/pkg/errors"
)

// Credentials the loaded credentials
type Credentials struct {
	ServerURL string
	Username  string
	Password  string
	Token     string
}

// GetServerCredentials returns the credentials for the given server URL or an empty struct
func GetServerCredentials(credentials []Credentials, serverURL string) Credentials {
	for i := range credentials {
		c := credentials[i]
		if c.ServerURL == serverURL {
			return c
		}
	}
	return Credentials{}
}

// GitCredentialsFile returns the location of the git credentials file
func GitCredentialsFile() (string, error) {
	homeDir := homedir.HomeDir()
	cfgHome := os.Getenv("XDG_CONFIG_HOME")
	if cfgHome == "" {
		cfgHome = homeDir
	}
	if cfgHome == "" {
		cfgHome = "."
	}
	paths := []string{
		filepath.Join(cfgHome, "git", "credentials"),
		filepath.Join(cfgHome, ".git-credentials"),
	}
	if homeDir != "" && homeDir != cfgHome {
		paths = append(paths,
			filepath.Join(homeDir, "git", "credentials"),
			filepath.Join(homeDir, ".git-credentials"),
		)
	}
	paths = append(paths,
		filepath.Join("git", "credentials"),
		filepath.Join(".git-credentials"),
	)
	for _, path := range paths {
		exists, err := files.FileExists(path)
		if err != nil {
			return path, errors.Wrapf(err, "failed to check if git credentials file exists %s", path)
		}
		if exists {
			return path, nil
		}
	}
	return "", nil
}

// LoadGitCredentialsFile loads the git credentials from the `git/credentials` file
// in `$XDG_CONFIG_HOME/git/credentials` or in the `~/git/credentials` directory
func LoadGitCredential() ([]Credentials, error) {
	fileName, err := GitCredentialsFile()
	if err != nil {
	  return nil, errors.Wrapf(err, "failed to find git credentials file")
	}
	if fileName == "" {
		return nil, nil
	}
	return LoadGitCredentialsFile(fileName)
}

// loadGitCredentialsAuthFile loads the git credentials file
func LoadGitCredentialsFile(fileName string) ([]Credentials, error) {
	log.Logger().Infof("loading git credentails file %s", termcolor.ColorInfo(fileName))
	
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load git credentials file %s", fileName)
	}

	var answer []Credentials
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		u, err := url.Parse(line)
		if err != nil {
			log.Logger().Warnf("ignoring invalid line in git credentials file: %s error: %s", fileName, err.Error())
			continue
		}

		user := u.User
		username := user.Username()
		password, _ := user.Password()
		if username == "" {
			log.Logger().Warnf("ignoring missing user name in git credentials file: %s URL: %s", fileName, line)
			continue
		}
		if password == "" {
			log.Logger().Warnf("ignoring missing password in git credentials file: %s URL: %s", fileName, line)
			continue
		}
		u.User = nil
		config := Credentials{}
		config.ServerURL = u.String()
		config.Username = username
		config.Password = password
		answer = append(answer, config)
	}
	return answer, nil
}
