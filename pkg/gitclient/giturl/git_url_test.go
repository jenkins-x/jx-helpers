// +build unit

package giturl_test

import (
	"testing"

	"github.com/jenkins-x/jx-helpers/v3/pkg/gitclient/giturl"
	"github.com/stretchr/testify/assert"
)

type parseGitURLData struct {
	url          string
	host         string
	organisation string
	name         string
}

type parseGitOrganizationURLData struct {
	url          string
	host         string
	organisation string
}

func TestParseGitOrganizationURL(t *testing.T) {
	t.Parallel()
	testCases := []parseGitOrganizationURLData{
		{
			"git://host.xz/org/repo", "host.xz", "org",
		},
		{
			"git://host.xz/org", "host.xz", "org",
		},
		{
			"git://host.xz/org/repo.git", "host.xz", "org",
		},
		{
			"git://host.xz/org/repo.git/", "host.xz", "org",
		},
		{
			"git://github.com/jstrachan/npm-pipeline-test-project.git", "github.com", "jstrachan",
		},
		{
			"https://github.com/fabric8io/foo.git", "github.com", "fabric8io",
		},
		{
			"https://github.com/fabric8io/foo", "github.com", "fabric8io",
		},
		{
			"git@github.com:jstrachan/npm-pipeline-test-project.git", "github.com", "jstrachan",
		},
		{
			"git@github.com:bar/foo.git", "github.com", "bar",
		},
		{
			"git@github.com:bar/foo", "github.com", "bar",
		},
		{
			"bar/foo", "github.com", "bar",
		},
		{
			"bar", "github.com", "bar",
		},
		{
			"http://test-user@auth.example.com/scm/bar/foo.git", "auth.example.com", "bar",
		},
		{
			"https://bitbucketserver.com/projects/myproject/repos/foo/pull-requests/1", "bitbucketserver.com", "myproject",
		},
	}
	for _, data := range testCases {
		info, err := giturl.ParseGitOrganizationURL(data.url)
		assert.Nil(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, data.host, info.Host, "Host does not match for input %s", data.url)
		assert.Equal(t, data.organisation, info.Organisation, "Organisation does not match for input %s", data.url)
	}
}

func TestParseGitURL(t *testing.T) {
	t.Parallel()
	testCases := []parseGitURLData{
		{
			"git://host.xz/org/repo", "host.xz", "org", "repo",
		},
		{
			"git://host.xz/org/repo.git", "host.xz", "org", "repo",
		},
		{
			"git://host.xz/org/repo.git/", "host.xz", "org", "repo",
		},
		{
			"git://github.com/jstrachan/npm-pipeline-test-project.git", "github.com", "jstrachan", "npm-pipeline-test-project",
		},
		{
			"https://github.com/fabric8io/foo.git", "github.com", "fabric8io", "foo",
		},
		{
			"https://github.com/fabric8io/foo", "github.com", "fabric8io", "foo",
		},
		{
			"git@github.com:jstrachan/npm-pipeline-test-project.git", "github.com", "jstrachan", "npm-pipeline-test-project",
		},
		{
			"git@github.com:bar/foo.git", "github.com", "bar", "foo",
		},
		{
			"git@github.com:bar/foo", "github.com", "bar", "foo",
		},
		{
			"git@github.com:bar/overview", "github.com", "bar", "overview",
		},
		{
			"git@gitlab.com:bar/subgroup/foo", "gitlab.com", "bar", "foo",
		},
		{
			"https://gitlab.com/bar/subgroup/foo", "gitlab.com", "bar", "foo",
		},
		{
			"https://gitlab.com/bar/subgroup/overview", "gitlab.com", "bar", "overview",
		},
		{
			"bar/foo", "github.com", "bar", "foo",
		},
		{
			"http://test-user@auth.example.com/scm/bar/foo.git", "auth.example.com", "bar", "foo",
		},
		{
			"https://bitbucketserver.com/projects/myproject/repos/foo/pull-requests/1", "bitbucketserver.com", "myproject", "foo",
		},
		{
			"https://bitbucketserver.com/projects/myproject/repos/foo/pull-requests/1/overview", "bitbucketserver.com", "myproject", "foo",
		},
	}
	for _, data := range testCases {
		info, err := giturl.ParseGitURL(data.url)
		assert.Nil(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, data.host, info.Host, "Host does not match for input %s", data.url)
		assert.Equal(t, data.organisation, info.Organisation, "Organisation does not match for input %s", data.url)
		assert.Equal(t, data.name, info.Name, "Name does not match for input %s", data.url)
	}
}

func TestSaasKind(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		gitURL string
		kind   string
	}{
		"GitHub": {
			gitURL: "https://github.com/test",
			kind:   giturl.KindGitHub,
		},
		"GitHub Enterprise": {
			gitURL: "https://github.test.com",
			kind:   giturl.KindGitHub,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			k := giturl.SaasGitKind(tc.gitURL)
			assert.Equal(t, tc.kind, k)
		})
	}
}

func TestGitInfoHttpCloneURL(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		gitInfo  *giturl.GitRepository
		kind     string
		expected string
	}{
		{
			name: "github.com",
			gitInfo: &giturl.GitRepository{
				Name:         "some-repo",
				Host:         "github.com",
				Organisation: "some-org",
			},
			kind:     giturl.KindGitHub,
			expected: "https://github.com/some-org/some-repo.git",
		},
		{
			name: "github enterprise",
			gitInfo: &giturl.GitRepository{
				Name:         "some-repo",
				Host:         "somewhereelse.com",
				Organisation: "some-org",
			},
			kind:     giturl.KindGitHub,
			expected: "https://somewhereelse.com/some-org/some-repo.git",
		},
		{
			name: "gitlab",
			gitInfo: &giturl.GitRepository{
				Name:         "some-repo",
				Host:         "gitlab.com",
				Organisation: "some-org",
			},
			kind:     giturl.KindGitlab,
			expected: "https://gitlab.com/some-org/some-repo.git",
		},
		{
			name: "bitbucket server",
			gitInfo: &giturl.GitRepository{
				Name:         "some-repo",
				Host:         "bbs.something.com",
				Organisation: "some-org",
			},
			kind:     giturl.KindBitBucketServer,
			expected: "https://bbs.something.com/scm/some-org/some-repo.git",
		},
		{
			name: "no kind",
			gitInfo: &giturl.GitRepository{
				Name:         "some-repo",
				Host:         "whatever.com",
				Organisation: "some-org",
			},
			expected: "https://whatever.com/some-org/some-repo.git",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := giturl.HttpCloneURL(tc.gitInfo, tc.kind)
			assert.Equal(t, tc.expected, result)
		})
	}
}
