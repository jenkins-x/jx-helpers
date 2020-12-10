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
