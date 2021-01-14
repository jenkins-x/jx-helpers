package requirements

import (
	"fmt"
	"os"
	"strings"

	v1 "github.com/jenkins-x/jx-api/v4/pkg/apis/jenkins.io/v1"
	"github.com/jenkins-x/jx-api/v4/pkg/util"
	"github.com/jenkins-x/jx-helpers/v3/pkg/gitclient/credentialhelper"

	"github.com/ghodss/yaml"
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

	// if we have a kubernetes secret with git auth mounted to the filesystem when running in cluster
	// we need to turn it into a git credentials file see https://git-scm.com/docs/git-credential-store
	secretMountPath := os.Getenv(credentialhelper.GIT_SECRET_MOUNT_PATH)
	if secretMountPath != "" {
		err = credentialhelper.WriteGitCredentialFromSecretMount()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to write git credentials file for secret %s ", secretMountPath)
		}
	}
	// clone cluster repo to a temp dir and load the requirements
	dir, err := gitclient.CloneToDir(g, env.Spec.Source.URL, "")
	if err != nil {
		return nil, errors.Wrapf(err, "failed to clone cluster git repo %s", env.Spec.Source.URL)
	}

	requirements, _, err := jxcore.LoadRequirementsConfig(dir, false)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find requirements from git clone in directory %s", dir)
	}

	if &requirements.Spec == nil {
		return nil, errors.Wrapf(err, "failed to load requirements in directory %s", dir)
	}
	return &requirements.Spec, nil
}

func GetTeamSettingsRequirementsConfig(settings *v1.TeamSettings) (*jxcore.RequirementsConfig, error) {
	if settings == nil {
		return nil, nil
	}

	// TeamSettings does not have a real value for BootRequirements, so this is probably not a boot cluster.
	if settings.BootRequirements == "" {
		return nil, nil
	}

	config := &jxcore.RequirementsConfig{}
	data := []byte(settings.BootRequirements)
	validationErrors, err := util.ValidateYaml(config, data)
	if err != nil {
		return config, fmt.Errorf("failed to validate requirements from team settings due to %s", err)
	}
	if len(validationErrors) > 0 {
		return config, fmt.Errorf("validation failures in requirements from team settings:\n%s", strings.Join(validationErrors, "\n"))
	}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return config, fmt.Errorf("failed to unmarshal requirements from team settings due to %s", err)
	}
	return config, nil
}
