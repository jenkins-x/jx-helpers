package testhelpers

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/go-errors/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertFileContains asserts that a given file exists and contains the given text
func AssertFileContains(t *testing.T, fileName string, containsText string) {
	if assert.FileExists(t, fileName) {
		data, err := os.ReadFile(fileName)
		assert.NoError(t, err, "Failed to read file %s", fileName)
		if err == nil {
			text := string(data)
			assert.True(t, strings.Index(text, containsText) >= 0, "The file %s does not contain text: %s", fileName, containsText)
		}
	}
}

// AssertFileDoesNotContain asserts that a given file exists and does not contain the given text
func AssertFileDoesNotContain(t *testing.T, fileName string, containsText string) {
	if assert.FileExists(t, fileName) {
		data, err := os.ReadFile(fileName)
		assert.NoError(t, err, "Failed to read file %s", fileName)
		if err == nil {
			text := string(data)
			idx := strings.Index(text, containsText)
			line := ""
			if idx > 0 {
				lines := strings.Split(text, "\n")
				for i, l := range lines {
					if strings.Index(l, containsText) >= 0 {
						line = "line " + strconv.Itoa(i+1) + " = " + l
						break
					}
				}
			}
			assert.True(t, idx < 0, "The file %s should not contain text: %s as %s", fileName, containsText, line)
		}
	}
}

// AssertFilesExist asserts that the list of file paths either exist or don't exist
func AssertFilesExist(t *testing.T, expected bool, paths ...string) {
	for _, path := range paths {
		if expected {
			assert.FileExists(t, path)
		} else {
			assert.NoFileExists(t, path)
		}
	}
}

// AssertDirsExist asserts that the list of directory paths either exist or don't exist
func AssertDirsExist(t *testing.T, expected bool, paths ...string) {
	for _, path := range paths {
		if expected {
			assert.DirExists(t, path)
		} else {
			assert.NoDirExists(t, path)
		}
	}
}

func AssertEqualFileText(t *testing.T, expectedFile string, actualFile string) error {
	expectedText, err := AssertLoadFileText(t, expectedFile)
	if err != nil {
		return err
	}
	actualText, err := AssertLoadFileText(t, actualFile)
	if err != nil {
		return err
	}
	assert.Equal(t, expectedText, actualText, "comparing text content of files %s and %s", expectedFile, actualFile)
	return nil
}

// AssertFilesEqualText asserts that all the given paths in the expected dir are equal to the files in the actual dir
func AssertFilesEqualText(t *testing.T, expectedDir string, actualDir string, paths ...string) {
	for _, f := range paths {
		generated := filepath.Join(actualDir, f)
		expected := filepath.Join(expectedDir, f)
		AssertEqualFileText(t, expected, generated)
		t.Logf("file %s matches expected text\n", f)
	}
}

// AssertLoadFileText asserts that the given file name can be loaded and returns the string contents
func AssertLoadFileText(t *testing.T, fileName string) (string, error) {
	if !assert.FileExists(t, fileName) {
		return "", fmt.Errorf("File %s does not exist", fileName)
	}
	data, err := os.ReadFile(fileName)
	assert.NoError(t, err, "Failed loading data for file: %s", fileName)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// AssertTextFileContentsEqual asserts that both the expected and actual files can be loaded as text
// and that their contents are identical
func AssertTextFileContentsEqual(t *testing.T, expectedFile string, actualFile string) {
	assert.NotEqual(t, expectedFile, actualFile, "should be given different file names")

	expected, err := AssertLoadFileText(t, expectedFile)
	require.NoError(t, err)
	actual, err := AssertLoadFileText(t, actualFile)
	require.NoError(t, err)

	assert.Equal(t, expected, actual, "contents of expected file %s and actual file %s", expectedFile, actualFile)
}

// AssertDirContentsEqual walks two directory structures and validates that the same files exist (by name) and that they have the same content
func AssertDirContentsEqual(t *testing.T, expectedDir string, actualDir string) {
	actualFiles := make(map[string]string, 0)
	expectedFiles := make(map[string]string, 0)
	err := filepath.Walk(actualDir, func(path string, info os.FileInfo, err error) error {
		relativePath, err := filepath.Rel(actualDir, path)
		if err != nil {
			return errors.New(err)
		}
		actualFiles[relativePath] = path
		return nil
	})
	assert.NoError(t, err)
	err = filepath.Walk(expectedDir, func(path string, info os.FileInfo, err error) error {
		relativePath, err := filepath.Rel(expectedDir, path)
		if err != nil {
			return errors.New(err)
		}
		expectedFiles[relativePath] = path
		return nil
	})
	assert.Len(t, actualFiles, len(expectedFiles))
	for relativePath, path := range expectedFiles {
		actualFile, ok := actualFiles[relativePath]
		assert.True(t, ok, "%s not present", relativePath)
		info, err := os.Stat(actualFile)
		assert.NoError(t, err)
		if !info.IsDir() {
			AssertTextFileContentsEqual(t, path, actualFile)
		}
	}
}
