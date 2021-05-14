package requirements_test

import (
	"path/filepath"
	"testing"

	"github.com/jenkins-x/jx-helpers/v3/pkg/requirements"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadSettings(t *testing.T) {
	testCases := []struct {
		path           string
		failOnError    bool
		expectError    bool
		expectNil      bool
		expectedGitURL string
	}{
		{
			path:        "bad_header",
			failOnError: true,
			expectError: true,
		},
		{
			path:        "bad_spec",
			failOnError: true,
			expectError: true,
		},
		{
			path:           "good",
			failOnError:    true,
			expectError:    false,
			expectedGitURL: "https://something.com/cheese.git",
		},
	}

	for _, tc := range testCases {
		dir := filepath.Join("test_data", tc.path)
		settings, err := requirements.LoadSettings(dir, tc.failOnError)

		expectNil := tc.expectNil
		if tc.expectError {
			expectNil = true
			require.Error(t, err, "expected error for %s", tc.path)
			t.Logf("got expected error %s for %s\n", err.Error(), tc.path)
		} else {
			require.NoError(t, err, "should not fail for %s", tc.path)
		}
		if expectNil {
			require.Nil(t, settings, "should have no settings for %s", tc.path)
		} else {
			require.NotNil(t, settings, "should have settings for %s", tc.path)
		}
		if tc.expectedGitURL != "" {
			assert.Equal(t, tc.expectedGitURL, settings.Spec.GitURL, "spec.gitUrl for %s", tc.path)
			t.Logf("test %s got gitUrl: %s\n", tc.path, settings.Spec.GitURL)
		}
	}
}
