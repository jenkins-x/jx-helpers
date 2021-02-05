package scmhelpers

import (
	"strconv"
	"strings"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/jx-helpers/v3/pkg/gitclient/giturl"
	"github.com/pkg/errors"
)

// ParsePullRequestURL parses the PullRequest from the string
func ParsePullRequestURL(url string) (*scm.PullRequest, error) {
	u := strings.TrimSuffix(url, "/")
	idx := strings.LastIndex(u, "/")
	if idx < 0 {
		return nil, errors.Errorf("expected string like https://github.com/owner/repo/pulls/1234 but got %s", url)
	}
	text := u[idx+1:]
	if text == "" {
		return nil, errors.Errorf("no pull request number at the end of the string")
	}
	u = u[0:idx]
	n, err := strconv.Atoi(text)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse Pull Request URL %s number text: '%s'", url, text)
	}
	if n <= 0 {
		return nil, errors.Errorf("invalid PullRequest URL %s number %d", url, n)
	}

	idx = strings.LastIndex(u, "/")
	if idx < 0 {
		return nil, errors.Errorf("expected string like https://github.com/owner/repo/pulls/1234 but got %s", url)
	}
	u = u[0:idx]

	gitInfo, err := giturl.ParseGitURL(u)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse git URL %s", u)
	}

	owner := gitInfo.Organisation
	repo := gitInfo.Name
	fullName := scm.Join(owner, repo)

	return &scm.PullRequest{
		Number: n,
		Link:   url,
		Base: scm.PullRequestBranch{
			Repo: scm.Repository{
				Namespace: owner,
				Name:      repo,
				FullName:  fullName,
			},
		},
	}, nil
}
