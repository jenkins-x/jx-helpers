package requirements_test

import (
	"testing"

	jxcore "github.com/jenkins-x/jx-api/v4/pkg/apis/core/v4beta1"
	"github.com/jenkins-x/jx-helpers/v3/pkg/requirements"
	"github.com/stretchr/testify/assert"
)

func TestEnvironmentGitURL(t *testing.T) {
	req := &jxcore.RequirementsConfig{
		Cluster: jxcore.ClusterConfig{
			DestinationConfig: jxcore.DestinationConfig{
				EnvironmentGitOwner: "envowner",
			},
		},
		Environments: []jxcore.EnvironmentConfig{
			{
				Key:        "dev",
				Repository: "dev",
			},
			{
				Key:        "repo",
				Repository: "mydevrepo",
			},
			{
				Key:        "owner-repo",
				Owner:      "myowner",
				Repository: "mydevrepo",
			},
			{
				Key:    "url",
				GitURL: "https://myserver/cheese.git",
			},
		},
	}
	gitlabReq := &jxcore.RequirementsConfig{
		Cluster: jxcore.ClusterConfig{
			DestinationConfig: jxcore.DestinationConfig{
				EnvironmentGitOwner: "envowner",
			},
			GitKind:   "gitlab",
			GitServer: "https://my.gitlab.com",
		},
		Environments: []jxcore.EnvironmentConfig{
			{
				Key:        "gitlab-staging",
				Owner:      "myowner",
				Repository: "myrepo",
			},
		},
	}

	bbsReq := &jxcore.RequirementsConfig{
		Cluster: jxcore.ClusterConfig{
			GitKind:   "bitbucketserver",
			GitServer: "https://my.bitbucket.server.com",
		},
		Environments: []jxcore.EnvironmentConfig{
			{
				Key:        "bbs-prod",
				Owner:      "cheese",
				Repository: "my-prod",
			},
		},
	}

	testCases := []struct {
		env      string
		expected string
		req      *jxcore.RequirementsConfig
	}{
		{
			env:      "dev",
			req:      req,
			expected: "https://github.com/envowner/dev.git",
		},
		{
			env:      "repo",
			req:      req,
			expected: "https://github.com/envowner/mydevrepo.git",
		},
		{
			env:      "owner-repo",
			req:      req,
			expected: "https://github.com/myowner/mydevrepo.git",
		},
		{
			env:      "url",
			req:      req,
			expected: "https://myserver/cheese.git",
		},
		{
			env:      "gitlab-staging",
			req:      gitlabReq,
			expected: "https://my.gitlab.com/myowner/myrepo.git",
		},
		{
			env:      "bbs-prod",
			req:      bbsReq,
			expected: "https://my.bitbucket.server.com/scm/cheese/my-prod.git",
		},
	}

	for _, tc := range testCases {
		actual := requirements.EnvironmentGitURL(tc.req, tc.env)
		assert.Equal(t, tc.expected, actual, "for environment %s", tc.env)
	}
}
