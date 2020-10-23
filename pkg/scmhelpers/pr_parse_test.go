package scmhelpers_test

import (
	"testing"

	"github.com/jenkins-x/jx-helpers/v3/pkg/scmhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePullRequestURL(t *testing.T) {
	u := "https://github.com/myowner/myrepo/pull/1234"
	pr, err := scmhelpers.ParsePullRequestURL(u)
	require.NoError(t, err, "failed to parse %s", u)
	require.NotNil(t, pr, "should have returned a PullRequest for %s", u)

	assert.Equal(t, 1234, pr.Number, "pr.Number for %s", u)
	assert.Equal(t, u, pr.Link, "pr.Link for %s", u)
	repo := pr.Repository()
	assert.Equal(t, "myrepo", repo.Name, "pr.Repository().Name for %s", u)
	assert.Equal(t, "myowner", repo.Namespace, "pr.Repository().Namespace for %s", u)
	assert.Equal(t, "myowner/myrepo", repo.FullName, "pr.Repository().FullName for %s", u)

	t.Logf("parsed %s and got PR %s number: %d\n", u, repo.FullName, pr.Number)
}

func TestParsePullRequestWithInvalidURLsFail(t *testing.T) {
	badURLs := []string{"https://github.com/myowner/myrepo", "https://github.com/myowner/myrepo/pull/", "https://github.com/myowner/myrepo/pull//", "https://github.com/myowner/myrepo/pull/notNumber"}

	for _, u := range badURLs {
		_, err := scmhelpers.ParsePullRequestURL(u)
		require.Errorf(t, err, "should have failed to parse %s", u)
		t.Logf("on %s got expected error: %s", u, err.Error())
	}
}
