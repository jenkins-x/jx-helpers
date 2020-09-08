package scmhelpers

import (
	"os"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/go-scm/scm/factory"
	"github.com/jenkins-x/jx-helpers/pkg/gitclient/loadcreds"
	"github.com/pkg/errors"
)

// NewScmClient creates a new Scm client for the given git kind, server URL and token.
// If no token is supplied we default it
func NewScmClient(kind, gitServerURL, token string) (*scm.Client, string, error) {
	creds, err := loadcreds.LoadGitCredential()
	if err != nil {
		return nil, token, errors.Wrapf(err, "failed to load git credentials")
	}
	serverCreds := loadcreds.GetServerCredentials(creds, gitServerURL)
	if token == "" {
		token = serverCreds.Password
	}
	if token == "" {
		token = serverCreds.Token
	}
	if token == "" {
		token = os.Getenv("GIT_TOKEN")
	}
	if token == "" {
		return nil, token, errors.Wrapf(err, "failed to load git credentials")
	}
	if kind == "" || kind == "github" {
		kind = "github"
	}
	client, err := factory.NewClient(kind, gitServerURL, token)
	return client, token, err
}
