package scmhelpers

import (
	"net/url"
	"os"
	"strings"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/jx-api/v4/pkg/client/clientset/versioned"
	"github.com/jenkins-x/jx-helpers/v3/pkg/cmdrunner"
	"github.com/jenkins-x/jx-helpers/v3/pkg/gitclient"
	"github.com/jenkins-x/jx-helpers/v3/pkg/gitclient/cli"
	"github.com/jenkins-x/jx-helpers/v3/pkg/gitclient/gitdiscovery"
	"github.com/jenkins-x/jx-helpers/v3/pkg/gitclient/giturl"
	"github.com/jenkins-x/jx-helpers/v3/pkg/stringhelpers"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// Options helper for discovering the git source URL and token
type Options struct {
	Dir                string
	FullRepositoryName string
	Owner              string
	Repository         string
	Branch             string
	ScmClient          *scm.Client
	GitServerURL       string
	SourceURL          string
	GitKind            string
	GitToken           string
	Namespace          string
	DiscoverFromGit    bool
	PreferUpstream     bool
	IgnoreMissingToken bool
	JXClient           versioned.Interface
	GitURL             *giturl.GitRepository
	GitClient          gitclient.Interface
	CommandRunner      cmdrunner.CommandRunner
}

// AddFlags adds CLI arguments to configure the parameters
func (o *Options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Dir, "dir", "", ".", "the directory to search for the .git to discover the git source URL")
	cmd.Flags().StringVarP(&o.GitServerURL, "git-server", "", "", "the git server URL to create the git provider client. If not specified its defaulted from the current source URL")
	if !o.DiscoverFromGit {
		cmd.Flags().StringVarP(&o.FullRepositoryName, "repo", "r", "", "the full git repository name of the form 'owner/name'")
		cmd.Flags().StringVarP(&o.SourceURL, "source-url", "", "", "the git source URL of the repository")
		cmd.Flags().StringVarP(&o.Branch, "branch", "", "", "specifies the branch if not inside a git clone")
	}
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
		o.ScmClient, o.GitToken, err = NewScmClient(o.GitKind, o.GitServerURL, o.GitToken, o.IgnoreMissingToken)
		if err != nil {
			return errors.Wrapf(err, "failed to create ScmClient: try supply --git-token")
		}
		if o.ScmClient == nil {
			if o.IgnoreMissingToken {
				return nil
			}
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
		// we may be a git: style URL so lets just ignore errors
		return "", nil
	}
	if u == nil || u.User == nil {
		return "", nil
	}
	answer, _ := u.User.Password()
	return answer, nil
}

func (o *Options) discoverRepositoryDetails() error {
	var err error
	if !o.DiscoverFromGit {
		if o.Owner == "" {
			o.Owner = os.Getenv("REPO_OWNER")
		}
		if o.Repository == "" {
			o.Repository = os.Getenv("REPO_NAME")
		}
		if o.FullRepositoryName == "" && o.Owner != "" && o.Repository != "" {
			o.FullRepositoryName = scm.Join(o.Owner, o.Repository)
		}
	}
	if o.SourceURL == "" {
		if o.DiscoverFromGit {
			// lets try find the git URL from the current git clone
			o.SourceURL, err = gitdiscovery.FindGitURLFromDir(o.Dir, o.PreferUpstream)
			if err != nil {
				o.SourceURL = os.Getenv("REPO_URL")
				if o.SourceURL == "" {
					o.SourceURL = os.Getenv("SOURCE_URL")
				}
				if o.SourceURL == "" && o.GitServerURL != "" && o.FullRepositoryName != "" {
					o.SourceURL = stringhelpers.UrlJoin(o.GitServerURL, o.FullRepositoryName)
				}
				if o.SourceURL == "" {
					return errors.Wrapf(err, "failed to discover git URL in dir %s. you could try pass the git URL as an argument", o.Dir)
				}
			}
		} else {
			o.SourceURL = os.Getenv("REPO_URL")
			if o.SourceURL == "" {
				o.SourceURL = os.Getenv("SOURCE_URL")
			}
			if o.SourceURL == "" && o.GitServerURL != "" && o.FullRepositoryName != "" {
				o.SourceURL = stringhelpers.UrlJoin(o.GitServerURL, o.FullRepositoryName)
			}
			if o.SourceURL == "" {
				return errors.Errorf("failed to discover git URL in dir %s. you could try pass the git URL as an argument", o.Dir)
			}
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
	if o.FullRepositoryName == "" && o.Owner != "" && o.Repository != "" {
		o.FullRepositoryName = scm.Join(o.Owner, o.Repository)
	}
	if o.GitKind == "" {
		o.GitKind, err = DiscoverGitKind(o.JXClient, o.Namespace, o.GitServerURL)
		if err != nil {
			return errors.Wrapf(err, "failed to discover git kind")
		}
	}
	if o.Branch == "" {
		o.Branch = os.Getenv("BRANCH_NAME")
		if o.Branch == "" {
			// lets see if we have a PR number
			pullNumber := os.Getenv("PULL_NUMBER")
			if pullNumber != "" {
				o.Branch = "PR-" + pullNumber
			} else {
				o.Branch = os.Getenv("PULL_BASE_REF")
			}
		}
	}
	return nil
}

// GetBranch returns the configured branch or discovers it from git
func (o *Options) GetBranch() (string, error) {
	if o.Branch != "" {
		return o.Branch, nil
	}
	if o.GitClient == nil {
		o.GitClient = cli.NewCLIClient("", o.CommandRunner)
	}
	branch, err := o.GitClient.Command(o.Dir, "rev-parse", "--abbrev-ref", "HEAD")
	branch = strings.TrimSpace(branch)
	if err != nil {
		return "", errors.Wrapf(err, "failed to find git branch in dir %s", o.Dir)
	}
	return branch, nil
}
