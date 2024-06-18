package gitclient

import (
	"fmt"
	"os"
	"os/user"

	"github.com/jenkins-x/jx-helpers/v3/pkg/termcolor"

	"github.com/jenkins-x/jx-api/v4/pkg/util"
	"github.com/jenkins-x/jx-helpers/v3/pkg/homedir"
	"github.com/jenkins-x/jx-helpers/v3/pkg/kube"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
)

var info = termcolor.ColorInfo

// EnsureUserAndEmailSetup returns the user name and email for the gitter
// lazily setting them if they are blank either from the given values or if they are empty
// using environment variables `GIT_AUTHOR_NAME` and `GIT_AUTHOR_EMAIL` or using default values
func EnsureUserAndEmailSetup(gitter Interface, dir string, gitUserName string, gitUserEmail string) (string, string, error) {
	userName, _ := gitter.Command(dir, "config", "--get", "user.name")
	userEmail, _ := gitter.Command(dir, "config", "--get", "user.email")
	// Seems like changing it to alpine image, results in username being set to root
	// ToDo(@ankitm123): Look at why this happened in alpine go image and not the normal go image
	if userName == "" || userName == "root" {
		userName = gitUserName
		if userName == "" {
			userName = os.Getenv("GIT_AUTHOR_NAME")
			if userName == "" {
				user, err := user.Current()
				if err == nil && user != nil {
					userName = user.Username
				}
				if userName == "" {
					userName = DefaultGitUserName
				}
			}
		}
		_, err := gitter.Command(dir, "config", "--global", "--add", "user.name", userName)
		if err != nil {
			return userName, userEmail, fmt.Errorf("Failed to set the git username to %s: %w", userName, err)
		}
	}
	if userEmail == "" {
		userEmail = gitUserEmail
		if userEmail == "" {
			userEmail = os.Getenv("GIT_AUTHOR_EMAIL")
			if userEmail == "" {
				userEmail = DefaultGitUserEmail
			}
		}
		_, err := gitter.Command(dir, "config", "--global", "--add", "user.email", userEmail)
		if err != nil {
			return userName, userEmail, fmt.Errorf("Failed to set the git email to %s: %w", userEmail, err)
		}
	}
	return userName, userEmail, nil
}

// SetUserAndEmail sets the user and email if they have not been set
// Uses environment variables `GIT_AUTHOR_NAME` and `GIT_AUTHOR_EMAIL`
func SetUserAndEmail(gitter Interface, dir string, gitUserName string, gitUserEmail string, assumeInCluster bool) (string, string, error) {
	userName := ""
	userEmail := ""
	if assumeInCluster || kube.IsInCluster() {
		userName = gitUserName
		userEmail = gitUserEmail
	} else {
		// lets load the current values and if they are specified lets not modify them as they are probably correct
		userName, _ = gitter.Command(dir, "config", "--global", "--get", "user.name")
		userEmail, _ = gitter.Command(dir, "config", "--global", "--get", "user.email")

		if userName != "" && userEmail != "" {
			log.Logger().Infof("have git user name %s and email %s setup already so not going to modify them", userName, userEmail)
			return userName, userEmail, nil
		}
	}
	if userName == "" {
		userName = os.Getenv("GIT_USER_NAME")
		if userName == "" {
			userName = os.Getenv("GIT_AUTHOR_NAME")
		}
		if userName == "" {
			user, err := user.Current()
			if err == nil && user != nil {
				userName = user.Username
			}
			if userName == "" {
				userName = DefaultGitUserName
			}
		}
	}
	_, err := gitter.Command(dir, "config", "--global", "--add", "user.name", userName)
	if err != nil {
		return userName, userEmail, fmt.Errorf("Failed to set the git username to %s: %w", userName, err)
	}
	if userEmail == "" {
		userName = os.Getenv("GIT_USER_EMAIL")
		if userEmail == "" {
			userEmail = os.Getenv("GIT_AUTHOR_EMAIL")
		}
		if userEmail == "" {
			userEmail = DefaultGitUserEmail
		}
	}
	_, err = gitter.Command(dir, "config", "--global", "--add", "user.email", userEmail)
	if err != nil {
		return userName, userEmail, fmt.Errorf("Failed to set the git email to %s: %w", userEmail, err)
	}
	log.Logger().Infof("setup git user %s email %s", info(userName), info(userEmail))
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
		return fmt.Errorf("failed to make sure the home directory %s was created: %w", dir, err)
	}

	_, err = gitter.Command(dir, "config", "--global", "credential.helper", "store")
	if err != nil {
		return fmt.Errorf("failed to setup git: %w", err)
	}
	if os.Getenv("XDG_CONFIG_HOME") == "" {
		log.Logger().Warnf("Note that the environment variable $XDG_CONFIG_HOME is not defined so we may not be able to push to git!")
	}
	return nil
}
