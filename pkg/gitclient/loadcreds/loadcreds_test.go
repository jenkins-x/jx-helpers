// +build unit

package loadcreds_test

import (
	"path/filepath"
	"testing"

	"github.com/jenkins-x/jx-helpers/pkg/gitclient/loadcreds"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadGitCredentials(t *testing.T) {
	fileName := filepath.Join("test_data", "git", "credentials")
	config, err := loadcreds.LoadGitCredentialsFile(fileName)
	require.NoError(t, err, "should not have failed to load file %s", fileName)
	assert.NotNil(t, config, "should have returned not nil config for file %s", fileName)

	serverURL := "http://cheese.com"
	username := "user2"
	password := "pwd2"

	assertServerUserPassword(t, config, "https://github.com", "user1", "pwd1")
	assertServerUserPassword(t, config, serverURL, username, password)
}

func assertServerUserPassword(t *testing.T, configs []loadcreds.Credentials, serverURL string, username string, password string) loadcreds.Credentials {
	credentials := loadcreds.GetServerCredentials(configs, serverURL)
	require.NotEmpty(t, credentials.ServerURL, "no server found for URL %s", serverURL)

	assert.Equal(t, username, credentials.Username, "credentials.Username for URL %s", serverURL)
	assert.Equal(t, password, credentials.Password, "credentials.Password for URL %s", serverURL)
	//assert.Equal(t, password, credentials.Token, "credentials.ApiToken for URL %s", serverURL)

	t.Logf("found server %s username %s password %s", credentials.ServerURL, credentials.Username, credentials.Password)
	return credentials
}

func TestLoadGitCredentialsFileDoesNotExist(t *testing.T) {
	config, err := loadcreds.LoadGitCredentialsFile("test_data/does/not/exist")
	require.NoError(t, err, "should not have failed to load non existing git creds file")
	assert.Nil(t, config, "should have returned nil config")
}
