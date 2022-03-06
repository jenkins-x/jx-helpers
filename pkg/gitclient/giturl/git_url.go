package giturl

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/jenkins-x/jx-helpers/v3/pkg/stringhelpers"
)

const (
	GitHubHost = "github.com"
	GitHubURL  = "https://github.com"

	gitPrefix = "git@"
)

// ParseGitURL attempts to parse the given text as a URL or git URL-like string to determine
// the protocol, host, organisation and name
func ParseGitURL(text string) (*GitRepository, error) {
	answer := GitRepository{
		URL: text,
	}
	u, err := url.Parse(text)
	if err == nil && u != nil {
		answer.Host = u.Host
		// lets default to github
		if answer.Host == "" {
			answer.Host = GitHubHost
		}
		if answer.Scheme == "" {
			answer.Scheme = "https"
		}
		answer.Scheme = u.Scheme
		return parsePath(u.Path, &answer, true)
	}

	// handle git@ kinds of URIs
	if strings.HasPrefix(text, gitPrefix) {
		t := strings.TrimPrefix(text, gitPrefix)
		t = strings.TrimPrefix(t, "/")
		t = strings.TrimPrefix(t, "/")
		t = strings.TrimSuffix(t, "/")
		t = strings.TrimSuffix(t, ".git")

		arr := stringhelpers.RegexpSplit(t, ":|/")
		if len(arr) >= 3 {
			answer.Scheme = "git"
			answer.Host = arr[0]
			answer.Organisation = arr[1]
			// Dont do exact match on gitlab.com as there can be custom gitlab domains
			if strings.Contains(answer.Host, "gitlab") {
				answer.Name = strings.Join(arr[2:], "/")
			} else {
				answer.Name = arr[len(arr)-1]
			}
			return &answer, nil
		}
	}
	return nil, fmt.Errorf("could not parse Git URL %s", text)
}

// ParseGitOrganizationURL attempts to parse the given text as a URL or git URL-like string to determine
// the protocol, host, organisation
func ParseGitOrganizationURL(text string) (*GitRepository, error) {
	answer := GitRepository{
		URL: text,
	}
	u, err := url.Parse(text)
	if err == nil && u != nil {
		answer.Host = u.Host

		// lets default to github
		if answer.Host == "" {
			answer.Host = GitHubHost
		}
		if answer.Scheme == "" {
			answer.Scheme = "https"
		}
		answer.Scheme = u.Scheme
		return parsePath(u.Path, &answer, false)
	}
	// handle git@ kinds of URIs
	if strings.HasPrefix(text, gitPrefix) {
		t := strings.TrimPrefix(text, gitPrefix)
		t = strings.TrimPrefix(t, "/")
		t = strings.TrimPrefix(t, "/")
		t = strings.TrimSuffix(t, "/")
		t = strings.TrimSuffix(t, ".git")

		arr := stringhelpers.RegexpSplit(t, ":|/")
		if len(arr) >= 3 {
			answer.Scheme = "git"
			answer.Host = arr[0]
			answer.Organisation = arr[1]
			return &answer, nil
		}
	}
	return nil, fmt.Errorf("could not parse Git URL %s", text)
}

func parsePath(path string, info *GitRepository, requireRepo bool) (*GitRepository, error) {
	// This is necessary for Bitbucket Server in some cases.
	trimPath := strings.TrimPrefix(path, "/scm")

	// This is necessary for Bitbucket Server, EG: /projects/ORG/repos/NAME/pull-requests/1/overview
	reOverview := regexp.MustCompile("/pull-requests/[0-9]+/overview$")
	if reOverview.MatchString(trimPath) {
		trimPath = strings.TrimSuffix(trimPath, "/overview")
	}

	// This is necessary for Bitbucket Server in other cases
	trimPath = strings.Replace(trimPath, "/projects/", "/", 1)
	trimPath = strings.Replace(trimPath, "/repos/", "/", 1)
	re := regexp.MustCompile("/pull.*/[0-9]+$")
	trimPath = re.ReplaceAllString(trimPath, "")

	// Remove leading and trailing slashes so that splitting on "/" won't result
	// in empty strings at the beginning & end of the array.
	trimPath = strings.TrimPrefix(trimPath, "/")
	trimPath = strings.TrimSuffix(trimPath, "/")

	trimPath = strings.TrimSuffix(trimPath, ".git")
	arr := strings.Split(trimPath, "/")
	if len(arr) >= 2 {
		// We're assuming the beginning of the path is of the form /<org>/<repo> or /<org>/<subgroup>/.../<repo>
		info.Organisation = arr[0]
		info.Project = arr[0]
		if strings.Contains(info.Host, "gitlab") {
			info.Name = strings.Join(arr[1:], "/")
		} else {
			info.Name = arr[len(arr)-1]
		}

		return info, nil
	} else if len(arr) == 1 && !requireRepo {
		// We're assuming the beginning of the path is of the form /<org>/<repo>
		info.Organisation = arr[0]
		info.Project = arr[0]
		return info, nil
	}

	return info, fmt.Errorf("invalid path %s could not determine organisation and repository name", path)
}

// HttpCloneURL returns the HTTPS git URL this repository
func HttpCloneURL(repo *GitRepository, kind string) string {
	if kind == KindBitBucketServer {
		host := repo.Host
		if !strings.Contains(host, ":/") {
			host = "https://" + host
		}
		return stringhelpers.UrlJoin(host, "scm", repo.Organisation, repo.Name) + ".git"

	}
	return repo.HTTPSURL() + ".git"
}
