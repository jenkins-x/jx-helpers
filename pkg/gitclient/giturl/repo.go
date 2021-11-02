package giturl

import (
	"net/url"
	"strings"

	"github.com/jenkins-x/jx-helpers/v3/pkg/stringhelpers"
)

type GitRepository struct {
	ID               int64
	Name             string
	AllowMergeCommit bool
	HTMLURL          string
	CloneURL         string
	SSHURL           string
	Language         string
	Fork             bool
	Stars            int
	URL              string
	Scheme           string
	Host             string
	Organisation     string
	Project          string
	Private          bool
	HasIssues        bool
	OpenIssueCount   int
	HasWiki          bool
	HasProjects      bool
	Archived         bool
}

func (i *GitRepository) IsGitHub() bool {
	return GitHubHost == i.Host || strings.HasSuffix(i.URL, "https://github.com")
}

// PullRequestURL returns the URL of a pull request of the given name/number
func (i *GitRepository) PullRequestURL(prName string) string {
	return stringhelpers.UrlJoin("https://"+i.Host, i.Organisation, i.Name, "pull", prName)
}

// HttpURL returns the URL to browse this repository in a web browser
func (i *GitRepository) HTTPURL() string {
	host := i.Host
	if !strings.Contains(host, ":/") {
		host = "http://" + host
	}
	return stringhelpers.UrlJoin(host, i.Organisation, i.Name)
}

// HttpsURL returns the URL to browse this repository in a web browser
func (i *GitRepository) HTTPSURL() string {
	host := i.Host
	if !strings.Contains(host, ":/") {
		host = "https://" + host
	}
	return stringhelpers.UrlJoin(host, i.Organisation, i.Name)
}

// HostURL returns the URL to the host
func (i *GitRepository) HostURL() string {
	answer := i.Host
	if !strings.Contains(answer, ":/") {
		// lets find the scheme from the URL
		u := i.URL
		if u != "" {
			u2, err := url.Parse(u)
			if err != nil {
				// probably a git@ URL
				return "https://" + answer
			}
			s := u2.Scheme
			if s != "" {
				if !strings.HasSuffix(s, "://") {
					s += "://"
				}
				return s + answer
			}
		}
		return "https://" + answer
	}
	return answer
}

// URLWithoutUser returns the URL without any user/password
func (i *GitRepository) URLWithoutUser() string {
	u := i.URL
	if u != "" {
		if strings.HasPrefix(u, gitPrefix) {
			return u
		}
		u2, err := url.Parse(u)
		if err == nil {
			u2.User = nil
			return u2.String()
		}

	}
	host := i.Host
	if !strings.Contains(host, ":/") {
		host = "https://" + host
	}
	return host
}

func (i *GitRepository) HostURLWithoutUser() string {
	u := i.URL
	if u != "" {
		u2, err := url.Parse(u)
		if err == nil {
			u2.User = nil
			u2.Path = ""
			return u2.String()
		}

	}
	host := i.Host
	if !strings.Contains(host, ":/") {
		host = "https://" + host
	}
	return host
}

// PipelinePath returns the pipeline path for the master branch which can be used to query
// pipeline logs in `jx get build logs myPipelinePath`
func (i *GitRepository) PipelinePath() string {
	return i.Organisation + "/" + i.Name + "/master"
}

// SaasGitKind returns the kind for SaaS Git providers or "" if the URL could not be deduced
func SaasGitKind(gitServiceUrl string) string {
	gitServiceUrl = strings.TrimSuffix(gitServiceUrl, "/")
	switch gitServiceUrl {
	case "http://github.com":
		return KindGitHub
	case "https://github.com":
		return KindGitHub
	case "https://gitlab.com":
		return KindGitlab
	case "http://bitbucket.org":
		return KindBitBucketCloud
	case BitbucketCloudURL:
		return KindBitBucketCloud
	case "http://fake.git", FakeGitURL:
		return KindGitFake
	default:
		if strings.HasPrefix(gitServiceUrl, "https://github") {
			return KindGitHub
		}
		return ""
	}
}
