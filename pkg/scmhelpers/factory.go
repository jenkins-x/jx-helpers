package scmhelpers

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/jx-helpers/v3/pkg/files"
	"github.com/jenkins-x/jx-helpers/v3/pkg/gitclient/giturl"
	"github.com/jenkins-x/jx-helpers/v3/pkg/gitclient/loadcreds"
	"github.com/jenkins-x/jx-helpers/v3/pkg/input"
	"github.com/jenkins-x/jx-helpers/v3/pkg/options"
	"github.com/jenkins-x/jx-helpers/v3/pkg/stringhelpers"
	"github.com/jenkins-x/jx-helpers/v3/pkg/termcolor"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	info = termcolor.ColorInfo
)

// Factory helper for discovering the git source URL and token
type Factory struct {
	GitKind      string
	GitServerURL string
	GitUsername  string
	GitToken     string
	ScmClient    *scm.Client

	// Input specifies the input to use if we are not using batch mode so that we can lazily populate
	// the git user and token if it cannot be discovered automatically
	Input input.Interface

	// GitCredentialFile allows faking for easier testing
	GitCredentialFile string
}

// AddFlags adds CLI arguments to configure the parameters
func (o *Factory) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.GitServerURL, "git-server", "", "", "the git server URL to create the scm client")
	cmd.Flags().StringVarP(&o.GitKind, "git-kind", "", "", "the kind of git server to connect to")
	cmd.Flags().StringVarP(&o.GitToken, "git-token", "", "", "the git token used to operate on the git repository. If not specified it's loaded from the git credentials file")
	cmd.Flags().StringVarP(&o.GitUsername, "git-username", "", "", "the git username used to operate on the git repository. If not specified it's loaded from the git credentials file")
}

// Create creates an ScmClient
func (o *Factory) Create() (*scm.Client, error) {
	if o.GitServerURL == "" {
		return nil, options.MissingOption("git-server")
	}
	if o.GitToken == "" {
		err := o.FindGitToken()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to find git token")
		}
	}

	scmClient, gitToken, err := NewScmClient(o.GitKind, o.GitServerURL, o.GitToken)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create ScmClient for server %s", o.GitServerURL)
	}
	o.ScmClient = scmClient
	if gitToken != "" {
		o.GitToken = gitToken
	}
	return o.ScmClient, nil
}

// FindGitToken ensures the GitToken is populated
func (o *Factory) FindGitToken() error {
	var err error
	fileName := o.GitCredentialFile
	if fileName == "" {
		fileName, err = loadcreds.GitCredentialsFile()
		if err != nil {
			return errors.Wrapf(err, "failed to find git credentials file")
		}
	}
	if fileName == "" {
		return errors.Wrapf(err, "could not deduce the git credentials file")
	}
	creds, exists, err := loadcreds.LoadGitCredentialsFile(fileName)
	if err != nil {
		return errors.Wrapf(err, "failed to load git credentials")
	}
	serverCreds := loadcreds.GetServerCredentials(creds, o.GitServerURL)
	if o.GitUsername == "" {
		o.GitUsername = serverCreds.Username
	}
	if o.GitToken == "" {
		o.GitToken = serverCreds.Password
	}
	if o.GitToken == "" {
		o.GitToken = serverCreds.Token
	}
	if o.GitUsername == "" {
		o.GitUsername = os.Getenv("GIT_USERNAME")
	}
	if o.GitToken == "" {
		o.GitToken = os.Getenv("GIT_TOKEN")
	}

	// lets try default missing values
	if o.Input != nil && (o.GitUsername == "" || o.GitToken == "") {
		if o.GitUsername == "" {
			message := fmt.Sprintf("please enter the git username to use for server %s:", o.GitServerURL)

			o.GitUsername, err = o.Input.PickValue(message, "", true, "we need a git username to use")
			if err != nil {
				return errors.Wrapf(err, "failed to enter the git user name")
			}
		}
		if o.GitToken == "" {
			tokenURL := giturl.ProviderAccessTokenURL(o.GitKind, o.GitServerURL, o.GitUsername)

			log.Logger().Infof("\nto create a git token click this URL %s", info(tokenURL))
			log.Logger().Infof("you can then copy the token into the following input...\n\n")

			message := fmt.Sprintf("please enter the git API token to use for server %s:", o.GitServerURL)

			o.GitToken, err = o.Input.PickValue(message, "", true, "we need a git token to use")
			if err != nil {
				return errors.Wrapf(err, "failed to enter the git user name")
			}
		}

		if o.GitUsername != "" && o.GitToken != "" {
			// lets append the git credential file...
			text := ""
			if exists {
				data, err := ioutil.ReadFile(fileName)
				if err != nil {
					return errors.Wrapf(err, "failed to load file %s", fileName)
				}
				text = string(data)
			}
			if text != "" && !strings.HasSuffix(text, "\n") {
				text = text + "\n"
			}

			authURL, err := o.CreateAuthenticatedURL(o.GitServerURL)
			if err != nil {
				return errors.Wrapf(err, "failed to create authenticated URL for %s", o.GitServerURL)
			}
			text += authURL + "\n"

			dir := filepath.Dir(fileName)
			err = os.MkdirAll(dir, files.DefaultDirWritePermissions)
			if err != nil {
				return errors.Wrapf(err, "failed to create git credentials dir: %s", dir)
			}
			err = ioutil.WriteFile(fileName, []byte(text), files.DefaultFileWritePermissions)
			if err != nil {
				return errors.Wrapf(err, "failed to save file %s", fileName)
			}

			log.Logger().Infof("saved git credentials to file %s", info(fileName))
		}
	}
	if o.GitToken == "" {
		return errors.Errorf("could not find git token for git server %s", o.GitServerURL)
	}
	return nil
}

// GetUsername gets the current user name from the ScmClient if its not explicitly configured
func (o *Factory) GetUsername() (string, error) {
	if o.GitUsername == "" {
		if o.ScmClient == nil {
			return "", errors.Errorf("no ScmClient created yet. Did you call Create()")
		}
		ctx := context.Background()
		user, _, err := o.ScmClient.Users.Find(ctx)
		if err != nil {
			return "", errors.Wrapf(err, "failed to lookup current user")
		}
		o.GitUsername = user.Login
	}
	return o.GitUsername, nil
}

// CreateAuthenticatedURL creates the Git repository URL with the username and password encoded for HTTPS based URLs
func (o *Factory) CreateAuthenticatedURL(cloneURL string) (string, error) {
	if o.GitToken == "" {
		return "", options.MissingOption("git-token")
	}
	userName, err := o.GetUsername()
	if err != nil {
		return "", errors.Wrapf(err, "failed to find GitUsername")
	}
	answer, err := stringhelpers.URLSetUserPassword(cloneURL, userName, o.GitToken)
	if err != nil {
		return "", errors.Wrapf(err, "failed to create authenticated git URL")
	}
	return answer, nil
}
