package scmhelpers

import (
	"context"
	"os"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/jx-helpers/pkg/gitclient/loadcreds"
	"github.com/jenkins-x/jx-helpers/pkg/options"
	"github.com/jenkins-x/jx-helpers/pkg/stringhelpers"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// Factory helper for discovering the git source URL and token
type Factory struct {
	GitKind      string
	GitServerURL string
	Owner        string
	GitUsername  string
	GitToken     string
	ScmClient    *scm.Client
}

// AddFlags adds CLI arguments to configure the parameters
func (o *Factory) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.GitServerURL, "git-server", "", "", "the git server URL to create the scm client")
	cmd.Flags().StringVarP(&o.GitKind, "git-kind", "", "", "the kind of git server to connect to")
	cmd.Flags().StringVarP(&o.GitToken, "git-token", "", "", "the git token used to operate on the git repository. If not specified it's loaded from the git credentials file")
	cmd.Flags().StringVarP(&o.GitUsername, "git-user", "", "", "the git username used to operate on the git repository")
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
	creds, err := loadcreds.LoadGitCredential()
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
