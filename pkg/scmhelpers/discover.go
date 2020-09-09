package scmhelpers

import (
	"net/url"
	"os"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/jx-helpers/pkg/gitclient/gitdiscovery"
	"github.com/jenkins-x/jx-helpers/pkg/gitclient/giturl"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// Options helper for discovering the git source URL and token
type Options struct {
	Dir                string
	FullRepositoryName string
	Repository         string
	Owner              string
	ScmClient          *scm.Client
	GitServerURL       string
	SourceURL          string
	GitKind            string
	GitToken           string
	GitURL             *giturl.GitRepository
}

// AddFlags adds CLI arguments to configure the parameters
func (o *Options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Dir, "dir", "", ".", "the directory to search for the .git to discover the git source URL")
	cmd.Flags().StringVarP(&o.FullRepositoryName, "repo", "r", "", "the full git repository name of the form 'owner/name'")
	cmd.Flags().StringVarP(&o.GitServerURL, "git-server", "", "", "the git server URL to create the git provider client. If not specified its defaulted from the current source URL")
	cmd.Flags().StringVarP(&o.GitKind, "git-kind", "", "", "the kind of git server to connect to")
	cmd.Flags().StringVarP(&o.GitToken, "git-token", "", "", "the git token used to operate on the git repository")
}

// Validate validates the inputs are valid and a ScmClient can be created
func (o *Options) Validate() error {
	var err error
	err = o.discoverRepositoryDetails()
	if err != nil {
		return errors.Wrapf(err, "failed to discover the repository details")
	}
	if o.GitServerURL == "" {
		return errors.Errorf("could not detect the git server URL. try supply --git-server")
	}
	if o.ScmClient == nil {
		if o.GitToken == "" && o.SourceURL != "" {
			// lets try get the git token from the source URL
			o.GitToken, err = GetPasswordFromSourceURL(o.SourceURL)
			if err != nil {
				return errors.Wrapf(err, "failed to detect git token from source URL")
			}
		}
		o.ScmClient, o.GitToken, err = NewScmClient(o.GitKind, o.GitServerURL, o.GitToken)
		if err != nil {
			return errors.Wrapf(err, "failed to create ScmClient: try supply --git-token")
		}
		if o.ScmClient == nil {
			return errors.Errorf("no ScmClient created for server %s", o.GitServerURL)
		}
	}
	return nil
}

// GetPasswordFromSourceURL returns password from the git URL
func GetPasswordFromSourceURL(sourceURL string) (string, error) {
	if sourceURL == "" {
		return "", nil
	}
	u, err := url.Parse(sourceURL)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse URL %s", sourceURL)
	}
	if u == nil || u.User == nil {
		return "", nil
	}
	answer, _ := u.User.Password()
	return answer, nil
}

func (o *Options) discoverRepositoryDetails() error {
	if o.SourceURL == "" {
		o.SourceURL = os.Getenv("SOURCE_URL")
	}
	var err error
	if o.SourceURL == "" {
		// lets try find the git URL from the current git clone
		o.SourceURL, err = gitdiscovery.FindGitURLFromDir(o.Dir)
		if err != nil {
			return errors.Wrapf(err, "failed to discover git URL in dir %s. you could try pass the git URL as an argument", o.Dir)
		}
	}
	if o.SourceURL != "" && o.GitURL == nil {
		o.GitURL, err = giturl.ParseGitURL(o.SourceURL)
		if err != nil {
			return errors.Wrapf(err, "failed to parse git URL %s", o.SourceURL)
		}
	}

	if o.GitURL != nil {
		if o.GitServerURL == "" {
			o.GitServerURL = o.GitURL.HostURL()
		}
		if o.Owner == "" {
			o.Owner = o.GitURL.Organisation
		}
		if o.Repository == "" {
			o.Repository = o.GitURL.Name
		}
	}
	if o.Owner == "" {
		o.Owner = os.Getenv("REPO_OWNER")
	}
	if o.Repository == "" {
		o.Repository = os.Getenv("REPO_NAME")
	}
	if o.FullRepositoryName == "" {
		o.FullRepositoryName = scm.Join(o.Owner, o.Repository)
	}
	return nil
}
