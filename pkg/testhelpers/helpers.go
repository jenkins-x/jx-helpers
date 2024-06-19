package testhelpers

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jenkins-x/jx-helpers/v3/pkg/files"
	"sigs.k8s.io/yaml"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// IsDebugLog debug log?
func IsDebugLog() bool {
	return strings.ToLower(os.Getenv("JX_TEST_DEBUG")) == "true"
}

// AssertYamlFilesEqual validates YAML file names without worrying about ordering of keys
func AssertYamlFilesEqual(t *testing.T, expectedFile string, actualFile string, message string, args ...interface{}) {
	suffix := fmt.Sprintf(message, args...)

	require.FileExists(t, expectedFile, "expected file for %s", suffix)
	require.FileExists(t, actualFile, "actual file for %s", suffix)

	expectedData, err := os.ReadFile(expectedFile)
	require.NoError(t, err, "failed to load expected file %s for %s", expectedFile, suffix)

	actualData, err := os.ReadFile(actualFile)
	require.NoError(t, err, "failed to load expected file %s for %s", actualFile, suffix)

	AssertYamlEqual(t, string(expectedData), string(actualData), message, args...)
}

// AssertYamlEqual validates YAML without worrying about ordering of keys
func AssertYamlEqual(t *testing.T, expected string, actual string, message string, args ...interface{}) {
	expectedMap := map[string]interface{}{}
	actualMap := map[string]interface{}{}

	reason := fmt.Sprintf(message, args...)

	err := yaml.Unmarshal([]byte(expected), &expectedMap)
	require.NoError(t, err, "failed to unmarshal expected yaml: %s for %s", expected, reason)

	err = yaml.Unmarshal([]byte(actual), &actualMap)
	require.NoError(t, err, "failed to unmarshal actual yaml: %s for %s", actual, reason)

	assert.Equal(t, expectedMap, actualMap, "parsed YAML contents not equal for %s", reason)
}

// AssertTextFilesEqual asserts that the expected file matches the actual file contents
func AssertTextFilesEqual(t *testing.T, expected string, actual string, message string) {
	require.FileExists(t, expected, "expected file for %s", message)
	require.FileExists(t, actual, "actual file for %s", message)

	wantData, err := os.ReadFile(expected)
	require.NoError(t, err, "could not load expected file %s for %s", expected, message)

	gotData, err := os.ReadFile(actual)
	require.NoError(t, err, "could not load actual file %s for %s", actual, message)
	assert.NoError(t, err)

	want := string(wantData)
	got := string(gotData)
	if diff := cmp.Diff(strings.TrimSpace(got), strings.TrimSpace(want)); diff != "" {
		t.Errorf("Unexpected file contents %s for %s", actual, message)
		t.Log(diff)

		t.Logf("generated %s for %s:\n", actual, message)
		t.Logf("\n%s\n", got)
		t.Logf("expected %s for %s:\n", expected, message)
		t.Logf("\n%s\n", want)
	}
}

// AssertFileNotExists asserts that a file does not exist
func AssertFileNotExists(t *testing.T, path string) {
	exists, err := files.FileExists(path)
	require.NoError(t, err, "failed to check if file exists %s", path)
	assert.False(t, exists, "file should not exist %s", path)
}
