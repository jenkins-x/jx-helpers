package requirements

import (
	"github.com/jenkins-x/jx-helpers/v3/pkg/gitclient/giturl"
	"github.com/jenkins-x/jx-helpers/v3/pkg/stringhelpers"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"os"

	"github.com/jenkins-x/jx-helpers/v3/pkg/gitclient/credentialhelper"

	jxcore "github.com/jenkins-x/jx-api/v4/pkg/apis/core/v4beta1"
	"github.com/jenkins-x/jx-api/v4/pkg/client/clientset/versioned"
	"github.com/jenkins-x/jx-helpers/v3/pkg/gitclient"
	"github.com/jenkins-x/jx-helpers/v3/pkg/kube/jxenv"
	"github.com/pkg/errors"
)

// GetClusterRequirementsConfig returns the cluster requirements from the cluster git repo
func GetClusterRequirementsConfig(g gitclient.Interface, jxClient versioned.Interface) (*jxcore.RequirementsConfig, error) {
	env, err := jxenv.GetDevEnvironment(jxClient, jxcore.DefaultNamespace)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get dev environment")
	}
	if env == nil {
		return nil, errors.New("failed to find a dev environment source url as there is no 'dev' Environment resource")
	}
	if env.Spec.Source.URL == "" {
		return nil, errors.New("failed to find a dev environment source url on development environment resource")
	}
	return GetRequirementsFromGit(g, env.Spec.Source.URL)
}

// GetRequirementsFromGit gets the requirements from the given git URL
func GetRequirementsFromGit(g gitclient.Interface, gitURL string) (*jxcore.RequirementsConfig, error) {
	req, _, err := GetRequirementsAndGit(g, gitURL)
	return req, err
}

// GetRequirementsAndGit gets the requirements from the given git URL and the git contents
func GetRequirementsAndGit(g gitclient.Interface, gitURL string) (*jxcore.RequirementsConfig, string, error) {
	// if we have a kubernetes secret with git auth mounted to the filesystem when running in cluster
	// we need to turn it into a git credentials file see https://git-scm.com/docs/git-credential-store
	secretMountPath := os.Getenv(credentialhelper.GIT_SECRET_MOUNT_PATH)
	if secretMountPath != "" {
		err := credentialhelper.WriteGitCredentialFromSecretMount()
		if err != nil {
			return nil, "", errors.Wrapf(err, "failed to write git credentials file for secret %s ", secretMountPath)
		}
	} else {
		log.Logger().Warnf("no $GIT_SECRET_MOUNT_PATH environment variable set")
	}
	// clone cluster repo to a temp dir and load the requirements
	dir, err := gitclient.CloneToDir(g, gitURL, "")
	if err != nil {
		return nil, dir, errors.Wrapf(err, "failed to clone cluster git repo %s", gitURL)
	}

	requirements, _, err := jxcore.LoadRequirementsConfig(dir, false)
	if err != nil {
		return nil, dir, errors.Wrapf(err, "failed to find requirements from git clone in directory %s", dir)
	}

	if &requirements.Spec == nil {
		return nil, dir, errors.Wrapf(err, "failed to load requirements in directory %s", dir)
	}
	return &requirements.Spec, dir, nil
}

// EnvironmentGitURL looks up the environment configuration based on environment name and returns the git URL
// or an empty string if the environment does not have an owner or repository configured
func EnvironmentGitURL(c *jxcore.RequirementsConfig, name string) string {
	for _, env := range c.Environments {
		if env.Key == name {
			if env.GitURL != "" {
				return env.GitURL
			}
			repo := env.Repository
			if repo == "" {
				return ""
			}
			gitServer := stringhelpers.FirstNotEmptyString(env.GitServer, c.Cluster.GitServer, giturl.GitHubURL)
			gitKind := stringhelpers.FirstNotEmptyString(env.GitKind, c.Cluster.GitKind)
			owner := stringhelpers.FirstNotEmptyString(env.Owner, c.Cluster.EnvironmentGitOwner)
			if owner == "" || gitServer == "" {
				return ""
			}
			if gitKind == "bitbucketserver" {
				gitServer = stringhelpers.UrlJoin(gitServer, "scm")
			}
			return stringhelpers.UrlJoin(gitServer, owner, repo) + ".git"
		}
	}
	return ""
}
