package scmhelpers_test

import (
	"path/filepath"
	"testing"

	"github.com/jenkins-x/jx-helpers/v3/pkg/gitclient/giturl"
	"github.com/jenkins-x/jx-helpers/v3/pkg/input/fake"
	"github.com/jenkins-x/jx-helpers/v3/pkg/scmhelpers"
	"github.com/jenkins-x/jx-helpers/v3/pkg/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPromptUserForGitUsernameAndToken(t *testing.T) {
	tmpDir := t.TempDir()

	credentialFile := filepath.Join(tmpDir, "git/credentials")

	fakeInput := &fake.FakeInput{
		OrderedValues: []string{
			"my-username",
			"my-token",
		},
	}

	gitServerURL := giturl.GitHubURL

	f := &scmhelpers.Factory{
		GitServerURL:      gitServerURL,
		GitCredentialFile: credentialFile,
		Input:             fakeInput,
	}

	err := f.FindGitToken()
	require.NoError(t, err, "failed to find token")

	assert.FileExists(t, credentialFile, "should have created the git credentials file")
	testhelpers.AssertTextFilesEqual(t, filepath.Join("test_data", "expected-credentials-file"), credentialFile, "should have created the credentials file")
}
