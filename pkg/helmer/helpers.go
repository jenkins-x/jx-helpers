package helmer

import (
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/jenkins-x/jx-helpers/pkg/termcolor"
	"github.com/jenkins-x/jx-logging/pkg/log"
	"github.com/pkg/errors"
)

// AddHelmRepoIfMissing will add the helm repo if there is no helm repo with that url present.
// It will generate the repoName from the url (using the host name) if the repoName is empty.
// The repo name may have a suffix added in order to prevent name collisions, and is returned for this reason.
// The username and password will be stored in vault for the URL (if vault is enabled).
func AddHelmRepoIfMissing(helmer Helmer, helmURL, repoName, username, password string) (string, error) {
	missing, existingName, err := helmer.IsRepoMissing(helmURL)
	if err != nil {
		return "", errors.Wrapf(err, "failed to check if the repository with URL '%s' is missing", helmURL)
	}
	if missing {
		if repoName == "" {
			// Generate the name
			uri, err := url.Parse(helmURL)
			if err != nil {
				repoName = uuid.New().String()
				log.Logger().Warnf("Unable to parse %s as URL so assigning random name %s", helmURL, repoName)
			} else {
				repoName = uri.Hostname()
			}
		}
		// Avoid collisions
		existingRepos, err := helmer.ListRepos()
		if err != nil {
			return "", errors.Wrapf(err, "listing helm repos")
		}
		baseName := repoName
		for i := 0; true; i++ {
			if _, exists := existingRepos[repoName]; exists {
				repoName = fmt.Sprintf("%s-%d", baseName, i)
			} else {
				break
			}
		}
		log.Logger().Infof("Adding missing Helm repo: %s %s", termcolor.ColorInfo(repoName), termcolor.ColorInfo(helmURL))
		err = helmer.AddRepo(repoName, helmURL, username, password)
		if err != nil {
			return "", errors.Wrapf(err, "failed to add the repository '%s' with URL '%s'", repoName, helmURL)
		}
		log.Logger().Infof("Successfully added Helm repository %s.", repoName)
	} else {
		repoName = existingName
	}
	return repoName, nil
}
