package gitclient

import (
	"os"
	"os/user"

	"github.com/jenkins-x/jx-api/pkg/util"
	"github.com/jenkins-x/jx-helpers/pkg/homedir"
	"github.com/jenkins-x/jx-logging/pkg/log"
	"github.com/pkg/errors"
)

// EnsureUserAndEmailSetup returns the user name and email for the gitter
// lazily setting them if they are blank either from the given values or if they are empty
// using environment variables `GIT_AUTHOR_NAME` and `GIT_AUTHOR_EMAIL` or using default values
func EnsureUserAndEmailSetup(gitter Interface, dir string, gitUserName string, gitUserEmail string) (string, string, error) {
	userName, _ := gitter.Command(dir, "config", "--get", "user.name")
	userEmail, _ := gitter.Command(dir, "config", "--get", "user.email")
	if userName == "" {
		userName = gitUserName
	}
	if userEmail == "" {
		userEmail = gitUserEmail
	}
	if userName == "" {
		userName = os.Getenv("GIT_AUTHOR_NAME")
		if userName == "" {
			user, err := user.Current()
			if err == nil && user != nil {
				userName = user.Username
			}
		}
		if userName == "" {
			userName = DefaultGitUserName
		}

		_, err := gitter.Command(dir, "config", "--global", "--add", "user.name", userName)
		if err != nil {
			return userName, userEmail, errors.Wrapf(err, "Failed to set the git username to %s", userName)
		}
	}
	if userEmail == "" {
		userEmail = os.Getenv("GIT_AUTHOR_EMAIL")
		if userEmail == "" {
			userEmail = DefaultGitUserEmail
		}
		_, err := gitter.Command(dir, "config", "--global", "--add", "user.email", userEmail)
		if err != nil {
			return userName, userEmail, errors.Wrapf(err, "Failed to set the git email to %s", userEmail)
		}
	}
	return userName, userEmail, nil
}

// SetCredentialHelper sets the credential store so that we detect the ~/git/credentials file for
// defaulting access tokens.
//
// If the dir parameter is blank we will use the home dir
func SetCredentialHelper(gitter Interface, dir string) error {
	if dir == "" {
		dir = homedir.HomeDir()
	}
	err := os.MkdirAll(dir, util.DefaultWritePermissions)
	if err != nil {
		return errors.Wrapf(err, "failed to make sure the home directory %s was created", dir)
	}

	_, err = gitter.Command(dir, "config", "--global", "credential.helper", "store")
	if err != nil {
		return errors.Wrapf(err, "failed to setup git")
	}
	if os.Getenv("XDG_CONFIG_HOME") == "" {
		log.Logger().Warnf("Note that the environment variable $XDG_CONFIG_HOME is not defined so we may not be able to push to git!")
	}
	return nil
}
