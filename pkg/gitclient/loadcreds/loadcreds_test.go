//go:build unit
// +build unit

package loadcreds_test

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/jenkins-x/jx-helpers/v3/pkg/gitclient/loadcreds"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitCredentialsFile(t *testing.T) {
	f, err := loadcreds.GitCredentialsFile()
	require.NoError(t, err, "failed to run GitCredentialsFile")
	t.Logf("found git credentials file %s\n", f)
}

func TestLoadGitCredentials(t *testing.T) {
	fileName := filepath.Join("testdata", "git", "credentials")
	config, exists, err := loadcreds.LoadGitCredentialsFile(fileName)
	require.NoError(t, err, "should not have failed to load file %s", fileName)
	assert.NotNil(t, config, "should have returned not nil config for file %s", fileName)
	assert.True(t, exists, "the git credentials file should exist %s", fileName)

	serverURL := "http://cheese.com"
	username := "user2"
	password := "pwd2" //NOSONAR

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

func TestPasswordMask(t *testing.T) {
	buffer := new(bytes.Buffer)
	log.SetOutput(buffer)
	fileName := filepath.Join("testdata", "git", "credentials_missing_pass")
	_, _, _ = loadcreds.LoadGitCredentialsFile(fileName)
	// The test credential file has pass1234 as the password, but our logs should not contain this
	assert.NotContains(t, buffer.String(), "pass1234")
}
