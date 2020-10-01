package versionstreamrepo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	v1 "github.com/jenkins-x/jx-api/pkg/apis/jenkins.io/v1"
	"github.com/jenkins-x/jx-api/pkg/config"
	"github.com/jenkins-x/jx-helpers/v3/pkg/files"
	"github.com/jenkins-x/jx-helpers/v3/pkg/gitclient"
	"github.com/jenkins-x/jx-helpers/v3/pkg/termcolor"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/pkg/errors"
	"gopkg.in/AlecAivazis/survey.v1"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

// CloneJXVersionsRepoToDir clones the version stream to the given directory
func CloneJXVersionsRepoToDir(configDir string, versionRepository string, versionRef string, settings *v1.TeamSettings, gitter gitclient.Interface, batchMode bool, advancedMode bool, handles files.IOFileHandles) (string, string, error) {
	dir, versionRef, err := cloneJXVersionsRepo(configDir, versionRepository, versionRef, settings, gitter, batchMode, advancedMode, handles)
	if err != nil {
		return "", "", errors.Wrapf(err, "")
	}
	if versionRef != "" {
		resolved, err := resolveRefToTag(dir, versionRef, gitter)
		if err != nil {
			return "", "", errors.WithStack(err)
		}
		return dir, resolved, nil
	}
	return dir, "", nil
}

func cloneJXVersionsRepo(configDir string, versionRepository string, versionRef string, settings *v1.TeamSettings, gitter gitclient.Interface, batchMode bool, advancedMode bool, handles files.IOFileHandles) (string, string, error) {
	surveyOpts := survey.WithStdio(handles.In, handles.Out, handles.Err)
	wrkDir := filepath.Join(configDir, "jenkins-x-versions")

	if settings != nil {
		if versionRepository == "" {
			versionRepository = settings.VersionStreamURL
		}
		if versionRef == "" {
			versionRef = settings.VersionStreamRef
		}
	}
	if versionRepository == "" {
		versionRepository = config.DefaultVersionsURL
	}
	if versionRef == "" {
		versionRef = config.DefaultVersionsRef
	}

	log.Logger().Debugf("Current configuration dir: %s", configDir)
	log.Logger().Debugf("VersionRepository: %s git ref: %s", versionRepository, versionRef)

	// If the repo already exists let's try to fetch the latest version
	if exists, err := files.DirExists(wrkDir); err == nil && exists {
		pullLatest := false
		if batchMode {
			pullLatest = true
		} else if advancedMode {
			confirm := &survey.Confirm{
				Message: "A local Jenkins X versions repository already exists, pull the latest?",
				Default: true,
			}
			err = survey.AskOne(confirm, &pullLatest, nil, surveyOpts)
			if err != nil {
				log.Logger().Errorf("Error confirming if we should pull latest, skipping %s", wrkDir)
			}
		} else {
			pullLatest = true
			// TODO
			//log.Logger().Debugf(util.QuestionAnswer("A local Jenkins X versions repository already exists, pulling the latest", util.YesNo(pullLatest)))
		}
		if pullLatest {
			_, err = gitter.Command(wrkDir, "fetch", versionRepository, "--tags")
			if err != nil {
				dir, err := deleteAndReClone(wrkDir, versionRepository, versionRef, gitter)
				if err != nil {
					return "", "", errors.WithStack(err)
				}
				return dir, versionRef, nil
			}
			_, err = gitter.Command(wrkDir, "fetch", versionRepository, versionRef)
			if err != nil {
				dir, err := deleteAndReClone(wrkDir, versionRepository, versionRef, gitter)
				if err != nil {
					return "", "", errors.WithStack(err)
				}
				return dir, versionRef, nil
			}

			isBranch, err := gitclient.RefIsBranch(gitter, wrkDir, versionRef)
			if err != nil {
				return "", "", err
			}

			if versionRef == config.DefaultVersionsRef || isBranch {
				_, err = gitter.Command(wrkDir, "checkout", versionRef)
				if err != nil {
					dir, err := deleteAndReClone(wrkDir, versionRepository, versionRef, gitter)
					if err != nil {
						return "", "", errors.WithStack(err)
					}
					return dir, versionRef, nil
				}
				_, err = gitter.Command(wrkDir, "reset", "--hard", "FETCH_HEAD")
				if err != nil {
					dir, err := deleteAndReClone(wrkDir, versionRepository, versionRef, gitter)
					if err != nil {
						return "", "", errors.WithStack(err)
					}
					return dir, versionRef, nil
				}
			} else {
				_, err = gitter.Command(wrkDir, "checkout", "FETCH_HEAD")
				if err != nil {
					dir, err := deleteAndReClone(wrkDir, versionRepository, versionRef, gitter)
					if err != nil {
						return "", "", errors.WithStack(err)
					}
					return dir, versionRef, nil
				}
			}
		}
		return wrkDir, versionRef, err
	}
	dir, err := deleteAndReClone(wrkDir, versionRepository, versionRef, gitter)
	if err != nil {
		return "", "", errors.WithStack(err)
	}
	return dir, versionRef, nil
}

func deleteAndReClone(wrkDir string, versionRepository string, referenceName string, gitter gitclient.Interface) (string, error) {
	log.Logger().Debug("Deleting and cloning the Jenkins X versions repo")
	err := os.RemoveAll(wrkDir)
	if err != nil {
		return "", errors.Wrapf(err, "failed to delete dir %s: %s\n", wrkDir, err.Error())
	}
	err = os.MkdirAll(wrkDir, files.DefaultDirWritePermissions)
	if err != nil {
		return "", errors.Wrapf(err, "failed to ensure directory is created %s", wrkDir)
	}
	_, err = clone(wrkDir, versionRepository, referenceName, gitter)
	if err != nil {
		return "", err
	}
	return wrkDir, err
}

func clone(wrkDir string, versionRepository string, referenceName string, gitter gitclient.Interface) (string, error) {
	if referenceName == "" || referenceName == "master" {
		referenceName = "refs/heads/master"
	} else if !strings.Contains(referenceName, "/") {
		if strings.HasPrefix(referenceName, "PR-") {
			prNumber := strings.TrimPrefix(referenceName, "PR-")

			log.Logger().Debugf("Cloning the Jenkins X versions repo %s with PR: %s to %s", termcolor.ColorInfo(versionRepository), termcolor.ColorInfo(referenceName), termcolor.ColorInfo(wrkDir))

			referenceName = fmt.Sprintf("refs/pull/%s/head", prNumber)

			// TODO
			//return "", shallowCloneGitRepositoryToDir(wrkDir, versionRepository, prNumber, "", gitter)
		}
		log.Logger().Debugf("Cloning the Jenkins X versions repo %s with revision %s to %s", termcolor.ColorInfo(versionRepository), termcolor.ColorInfo(referenceName), termcolor.ColorInfo(wrkDir))

		parentDir := filepath.Dir(wrkDir)

		_, err := gitter.Command(parentDir, "clone", versionRepository, wrkDir)
		if err != nil {
			return "", errors.Wrapf(err, "failed to clone repository: %s to dir %s", versionRepository, wrkDir)
		}

		_, err = gitter.Command(wrkDir, "fetch", "origin", referenceName)
		if err != nil {
			return "", errors.Wrapf(err, "failed to git fetch origin %s for repo: %s in dir %s", referenceName, versionRepository, wrkDir)
		}
		isBranch, err := gitclient.RefIsBranch(gitter, wrkDir, referenceName)
		if err != nil {
			return "", err
		}
		if isBranch {
			_, err = gitter.Command(wrkDir, "checkout", referenceName)
			if err != nil {
				return "", errors.Wrapf(err, "failed to checkout %s of repo: %s in dir %s", referenceName, versionRepository, wrkDir)
			}
			return "", nil
		}
		_, err = gitter.Command(wrkDir, "checkout", "FETCH_HEAD")
		if err != nil {
			return "", errors.Wrapf(err, "failed to checkout FETCH_HEAD of repo: %s in dir %s", versionRepository, wrkDir)
		}
		return "", nil
	}
	log.Logger().Infof("Cloning the Jenkins X versions repo %s with ref %s to %s", termcolor.ColorInfo(versionRepository), termcolor.ColorInfo(referenceName), termcolor.ColorInfo(wrkDir))
	// TODO: Change this to use gitter instead, but need to understand exactly what it's doing first.
	_, err := git.PlainClone(wrkDir, false, &git.CloneOptions{
		URL:           versionRepository,
		ReferenceName: plumbing.ReferenceName(referenceName),
		SingleBranch:  true,
		Progress:      nil,
	})
	if err != nil {
		return "", errors.Wrapf(err, "failed to clone reference: %s", referenceName)
	}
	return "", err
}

/*
func shallowCloneGitRepositoryToDir(dir string, gitURL string, pullRequestNumber string, revision string, gitter gitclient.Interface) error {
	if pullRequestNumber != "" {
		log.Logger().Infof("shallow cloning pull request %s of repository %s to temp dir %s", gitURL,
			pullRequestNumber, dir)
		err := gitclient.ShallowClone(gitter, gitURL, "", pullRequestNumber, dir)
		if err != nil {
			return errors.Wrapf(err, "shallow cloning pull request %s of repository %s to temp dir %s\n", gitURL,
				pullRequestNumber, dir)
		}
	} else if revision != "" {
		log.Logger().Infof("shallow cloning revision %s of repository %s to temp dir %s", gitURL,
			revision, dir)
		err := gitter.ShallowClone(dir, gitURL, revision, "")
		if err != nil {
			return errors.Wrapf(err, "shallow cloning revision %s of repository %s to temp dir %s\n", gitURL,
				revision, dir)
		}
	} else {
		log.Logger().Infof("shallow cloning master of repository %s to temp dir %s", gitURL, dir)
		err := gitter.ShallowClone(dir, gitURL, "", "")
		if err != nil {
			return errors.Wrapf(err, "shallow cloning master of repository %s to temp dir %s\n", gitURL, dir)
		}
	}

	return nil
}
*/

func resolveRefToTag(dir string, commitish string, gitter gitclient.Interface) (string, error) {
	_, err := gitter.Command(dir, "fetch", "--tags")
	if err != nil {
		return "", errors.Wrapf(err, "fetching tags")
	}
	resolved, _, err := gitclient.Describe(gitter, dir, true, commitish, "0", true)
	if err != nil {
		return "", errors.Wrapf(err, "running git describe %s --abbrev=0", commitish)
	}
	if resolved != "" {
		return resolved, nil
	}
	return resolved, nil
}
