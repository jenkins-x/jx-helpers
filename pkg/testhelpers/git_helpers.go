package testhelpers

import (
	"strings"
	"testing"

	"github.com/jenkins-x/jx-helpers/v3/pkg/cmdrunner"
	"github.com/jenkins-x/jx-helpers/v3/pkg/gitclient/cli"
)

// GetDefaultBranch gets the default branch name for tests
func GetDefaultBranch(t *testing.T) string {
	gitter := cli.NewCLIClient("", cmdrunner.QuietCommandRunner)

	defaultBranch := "master"
	text, err := gitter.Command(".", "config", "--global", "--get", "init.defaultBranch")
	if err == nil {
		text := strings.TrimSpace(text)
		if text != "" {
			defaultBranch = text
		}
	}
	t.Logf("using default branch: %s\n", defaultBranch)
	return defaultBranch
}
