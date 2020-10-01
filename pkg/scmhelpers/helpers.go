package scmhelpers

import (
	"context"
	"os"
	"strings"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/go-scm/scm/factory"
	"github.com/jenkins-x/jx-api/pkg/client/clientset/versioned"
	"github.com/jenkins-x/jx-helpers/v3/pkg/gitclient/giturl"
	"github.com/jenkins-x/jx-helpers/v3/pkg/gitclient/loadcreds"
	"github.com/jenkins-x/jx-helpers/v3/pkg/kube/jxclient"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// DiscoverGitKind discovers the git kind for the given git server from the SourceRepository resources.
// If no jxClient is provided it is lazily created
func DiscoverGitKind(jxClient versioned.Interface, namespace, gitServerURL string) (string, error) {
	gitServerURL = strings.TrimSuffix(gitServerURL, "/")
	if gitServerURL == "" {
		log.Logger().Warnf("cannot discover git kind as no git server URL")
		return "", nil
	}

	gitKind := giturl.SaasGitKind(gitServerURL)
	if gitKind != "" {
		return gitKind, nil
	}

	// lets try detect the git kind from the SourceRepository
	var err error
	jxClient, namespace, err = jxclient.LazyCreateJXClientAndNamespace(jxClient, namespace)
	if err != nil {
		return gitKind, errors.Wrapf(err, "failed to create jx client")
	}

	resources, err := jxClient.JenkinsV1().SourceRepositories(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil && apierrors.IsNotFound(err) {
		return gitKind, errors.Wrapf(err, "failed to list SourceRepository resources in namespace %s", namespace)
	}
	for _, sr := range resources.Items {
		ss := &sr.Spec
		if strings.TrimSuffix(ss.Provider, "/") == gitServerURL {
			if ss.ProviderKind != "" {
				gitKind = ss.ProviderKind
				return gitKind, nil
			}
			log.Logger().Warnf("no gitKind for SourceRepository %s", sr.Name)
		}
	}
	log.Logger().Warnf("no gitKind could be found for provider %s", gitServerURL)
	return gitKind, nil
}
