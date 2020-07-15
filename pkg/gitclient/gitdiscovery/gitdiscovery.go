package gitdiscovery

import (
	"fmt"

	"github.com/jenkins-x/jx-helpers/pkg/gitclient"
	"github.com/jenkins-x/jx-helpers/pkg/gitclient/gitconfig"
	"github.com/pkg/errors"
)

// FindGitURLFromDir tries to find the git clone URL from the given directory
func FindGitURLFromDir(dir string) (string, error) {
	_, gitConfDir, err := gitclient.FindGitConfigDir(dir)
	if err != nil {
		return "", errors.Wrapf(err, "there was a problem obtaining the git config dir of directory %s", dir)
	}

	if gitConfDir == "" {
		return "", fmt.Errorf("no .git directory could be found from dir %s", dir)
	}
	return gitconfig.DiscoverUpstreamGitURL(gitConfDir)
}
