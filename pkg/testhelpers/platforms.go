package testhelpers

import (
	"runtime"
	"testing"
)

// SkipForWindows skips tests if they are running on Windows
// This is to be used for valid tests that just don't work on windows for whatever reason
func SkipForWindows(t *testing.T, reason string) {
	if runtime.GOOS == "windows" {
		t.Skipf("Test skipped on windows. Reason: %s", reason)
	}
}
