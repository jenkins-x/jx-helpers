// +build unit

package cmdrunner_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/jenkins-x/jx-helpers/v3/pkg/cmdrunner"
	"github.com/jenkins-x/jx-helpers/v3/pkg/testhelpers"
	"github.com/stretchr/testify/assert"
)

const (
	scriptsDir   = "/test_data/scripts"
	errorMessage = "run should exit without failure"
	failure      = "FAILURE!"
)

func TestRunPass(t *testing.T) {
	t.Parallel()

	tmpFileName := "test_run_pass.txt"

	startPath, err := filepath.Abs("")
	if err != nil {
		panic(err)
	}
	exPath := filepath.Join(startPath, scriptsDir)
	tempfile, err := os.Create(filepath.Join(exPath, tmpFileName))
	tempfile.Close() // Close the file so that it can be edited by the script on windows
	defer os.Remove(tempfile.Name())

	cmd := cmdrunner.Command{
		Name:    filepath.Join(exPath, getFailIteratorScript()),
		Dir:     exPath,
		Args:    []string{tmpFileName, "3"},
		Timeout: 15 * time.Second,
	}

	res, err := cmd.Run()

	assert.NoError(t, err, errorMessage)
	assert.Equal(t, "PASS", res)
	assert.Equal(t, 2, len(cmd.Errors))
	assert.Equal(t, 3, cmd.Attempts())
	assert.Equal(t, true, cmd.DidError())
	assert.Equal(t, false, cmd.DidFail())
	assert.NotEqual(t, nil, cmd.Error())
}

func TestRunPassFirstTime(t *testing.T) {
	t.Parallel()

	tmpFileName := "test_run_pass_first_time.txt"

	startPath, err := filepath.Abs("")
	if err != nil {
		panic(err)
	}
	exPath := startPath + scriptsDir
	tempfile, err := os.Create(filepath.Join(exPath, tmpFileName))
	tempfile.Close()
	defer os.Remove(tempfile.Name())

	cmd := cmdrunner.Command{
		Name: filepath.Join(exPath, getFailIteratorScript()),
		Dir:  exPath,
		Args: []string{tmpFileName, "1"},
	}

	res, err := cmd.Run()

	assert.NoError(t, err, errorMessage)
	assert.Equal(t, "PASS", res)
	assert.Equal(t, 0, len(cmd.Errors))
	assert.Equal(t, 1, cmd.Attempts())
	assert.Equal(t, false, cmd.DidError())
	assert.Equal(t, false, cmd.DidFail())
	assert.Equal(t, nil, cmd.Error())

}

func TestRunFailWithTimeout(t *testing.T) {
	t.Parallel()

	tmpFileName := "test_run_fail_with_timeout.txt"

	startPath, err := filepath.Abs("")
	if err != nil {
		panic(err)
	}
	exPath := startPath + scriptsDir
	tempfile, err := os.Create(filepath.Join(exPath, tmpFileName))
	tempfile.Close()
	defer os.Remove(filepath.Join(exPath, tmpFileName))

	cmd := cmdrunner.Command{
		Name:    filepath.Join(exPath, getFailIteratorScript()),
		Dir:     exPath,
		Args:    []string{tmpFileName, "100"},
		Timeout: 1 * time.Second,
	}

	res, err := cmd.Run()

	assert.Error(t, err, errorMessage)
	assert.Equal(t, "", res)
	assert.Equal(t, true, cmd.DidError())
	assert.Equal(t, true, cmd.DidFail())
}

func TestRunThreadSafety(t *testing.T) {
	testhelpers.SkipForWindows(t, "Pre-existing test. Windows doesn't have a decent sleep builtin to run no-interactively")
	t.Parallel()
	startPath, err := filepath.Abs("")
	if err != nil {
		panic(err)
	}
	exPath := filepath.Join(startPath, scriptsDir)
	cmd := cmdrunner.Command{
		Name:    filepath.Join(exPath, "sleep.sh"),
		Dir:     exPath,
		Args:    []string{"0.2"},
		Timeout: 10000000 * time.Nanosecond,
	}

	res, err := cmd.Run()

	assert.NoError(t, err, errorMessage)
	assert.Equal(t, "0.2", res)
	assert.Equal(t, false, cmd.DidError())
	assert.Equal(t, false, cmd.DidFail())
	assert.Equal(t, 1, cmd.Attempts())
}

func TestRunWithoutRetry(t *testing.T) {
	t.Parallel()

	tmpFileName := "test_run_without_retry.txt"

	startPath, err := filepath.Abs("")
	if err != nil {
		panic(err)
	}
	tempfile, err := os.Create(filepath.Join(startPath, scriptsDir, tmpFileName))
	tempfile.Close()
	defer os.Remove(tempfile.Name())

	dir := filepath.Join(startPath, scriptsDir)
	cmd := cmdrunner.Command{
		Name:    filepath.Join(dir, getFailIteratorScript()),
		Dir:     dir,
		Args:    []string{tmpFileName, "100"},
		Timeout: 3 * time.Second,
	}

	res, err := cmd.RunWithoutRetry()

	assert.Error(t, err, errorMessage)
	assert.Equal(t, failure, res)
	assert.Equal(t, true, cmd.DidError())
	assert.Equal(t, true, cmd.DidFail())
	assert.Equal(t, 1, len(cmd.Errors))
	assert.Equal(t, 1, cmd.Attempts())

}

func TestRunVerbose(t *testing.T) {
	t.Parallel()

	tmpFileName := "test_run_verbose.txt"

	startPath, err := filepath.Abs("")
	if err != nil {
		panic(err)
	}
	tempfile, err := os.Create(filepath.Join(startPath, scriptsDir, tmpFileName))
	tempfile.Close()
	defer os.Remove(tempfile.Name())

	dir := filepath.Join(startPath, scriptsDir)
	cmd := cmdrunner.Command{
		Name:    filepath.Join(dir, getFailIteratorScript()),
		Dir:     dir,
		Args:    []string{tmpFileName, "100"},
		Timeout: 3 * time.Second,
	}

	res, err := cmd.RunWithoutRetry()

	assert.Error(t, err, errorMessage)
	assert.Equal(t, failure, res)
	assert.Equal(t, true, cmd.DidError())
	assert.Equal(t, true, cmd.DidFail())
	assert.Equal(t, 1, len(cmd.Errors))
	assert.Equal(t, 1, cmd.Attempts())
}

func TestRunQuiet(t *testing.T) {
	t.Parallel()

	tmpFileName := "test_run_quiet.txt"

	startPath, err := filepath.Abs("")
	if err != nil {
		panic(err)
	}
	tempfile, err := os.Create(filepath.Join(startPath, scriptsDir, tmpFileName))
	tempfile.Close()
	defer os.Remove(tempfile.Name())

	dir := filepath.Join(startPath, scriptsDir)
	cmd := cmdrunner.Command{
		Name:    filepath.Join(dir, getFailIteratorScript()),
		Dir:     dir,
		Args:    []string{tmpFileName, "100"},
		Timeout: 3 * time.Second,
	}

	res, err := cmd.RunWithoutRetry()

	assert.Error(t, err, errorMessage)
	assert.Equal(t, failure, res)
	assert.Equal(t, true, cmd.DidError())
	assert.Equal(t, true, cmd.DidFail())
	assert.Equal(t, 1, len(cmd.Errors))
	assert.Equal(t, 1, cmd.Attempts())
}

func getFailIteratorScript() string {
	ex := "fail_iterator.sh"
	if runtime.GOOS == "windows" {
		ex = "fail_iterator.bat"
	}
	return ex
}
