package credentialhelper

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jenkins-x/jx-helpers/v3/pkg/testhelpers"

	"github.com/jenkins-x/jx-helpers/v3/pkg/files"
	"github.com/stretchr/testify/assert"
)

func TestWriteGitCredentialToXDG(t *testing.T) {

	dir := t.TempDir()
	err := files.CopyDir(filepath.Join("test_data", "write_file_from_secret"), dir, true)
	assert.NoError(t, err)

	os.Setenv(XDG_CONFIG_HOME, filepath.Join(dir, "foo"))
	os.Setenv(GIT_SECRET_MOUNT_PATH, filepath.Join(dir, "bar"))

	err = WriteGitCredentialFromSecretMount()
	assert.NoError(t, err)

	testhelpers.AssertTextFileContentsEqual(t, filepath.Join(dir, "expected"), filepath.Join(dir, "foo", "git", "credentials"))

}

func TestWriteGitCredentialToUserHome(t *testing.T) {

	dir := t.TempDir()
	err := files.CopyDir(filepath.Join("test_data", "write_file_from_secret"), dir, true)
	assert.NoError(t, err)

	os.Unsetenv(XDG_CONFIG_HOME)
	os.Setenv("HOME", dir)
	os.Setenv(GIT_SECRET_MOUNT_PATH, filepath.Join(dir, "bar"))

	err = WriteGitCredentialFromSecretMount()
	assert.NoError(t, err)

	testhelpers.AssertTextFileContentsEqual(t, filepath.Join(dir, "expected"), filepath.Join(dir, ".git-credentials"))

}

func TestWriteGitCredentialError(t *testing.T) {

	os.Unsetenv(XDG_CONFIG_HOME)
	os.Unsetenv(GIT_SECRET_MOUNT_PATH)

	err := WriteGitCredentialFromSecretMount()
	assert.Error(t, err)
	assert.Equal(t, "no $GIT_SECRET_MOUNT_PATH environment variable set", err.Error())

}

func TestGitServerGithub(t *testing.T) {

	tests := []string{
		"",
		"github.com",
		"https://github.com",
	}

	for _, test := range tests {

		dir := t.TempDir()
		err := files.CopyDir(filepath.Join("test_data", "write_file_from_secret"), dir, true)
		assert.NoError(t, err)

		os.Setenv(XDG_CONFIG_HOME, filepath.Join(dir, "foo"))
		os.Setenv(GIT_SECRET_MOUNT_PATH, filepath.Join(dir, "bar"))
		os.Setenv(GIT_SECRET_SERVER, test)

		err = WriteGitCredentialFromSecretMount()
		assert.NoError(t, err)

		testhelpers.AssertTextFileContentsEqual(t, filepath.Join(dir, "expected"), filepath.Join(dir, "foo", "git", "credentials"))
	}
}

func TestGitServerAsNonGitHub(t *testing.T) {

	tests := []struct {
		testFolder string
		url        string
	}{

		{
			testFolder: "write_file_from_secret_http_url",
			url:        "http://example.com",
		},
		{
			testFolder: "write_file_from_secret_https_url",
			url:        "https://example.com",
		},
	}

	for _, test := range tests {

		dir := t.TempDir()
		err := files.CopyDir(filepath.Join("test_data", test.testFolder), dir, true)
		assert.NoError(t, err)

		os.Setenv(XDG_CONFIG_HOME, filepath.Join(dir, "foo"))
		os.Setenv(GIT_SECRET_MOUNT_PATH, filepath.Join(dir, "bar"))
		os.Setenv(GIT_SECRET_SERVER, test.url)

		err = WriteGitCredentialFromSecretMount()
		assert.NoError(t, err)

		testhelpers.AssertTextFileContentsEqual(t, filepath.Join(dir, "expected"), filepath.Join(dir, "foo", "git", "credentials"))
	}
}
