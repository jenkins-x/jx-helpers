package gitclient

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jenkins-x/jx-logging/pkg/log"
	"github.com/pkg/errors"
)

var (
	splitDescribeRegex = regexp.MustCompile(`(?:~|\^|-g)`)
)

// RefIsBranch looks for remove branches in ORIGIN for the provided directory and returns true if ref is found
func RefIsBranch(gitter Interface, dir string, ref string) (bool, error) {
	remoteBranches, err := RemoteBranches(gitter, dir)
	if err != nil {
		return false, errors.Wrapf(err, "error getting remote branches to find provided ref %s", ref)
	}
	for _, b := range remoteBranches {
		if strings.Contains(b, ref) {
			return true, nil
		}
	}
	return false, nil
}

// RemoteBranches returns the remote branches
func RemoteBranches(gitter Interface, dir string) ([]string, error) {
	answer := []string{}
	text, err := gitter.Command(dir, "branch", "-r")
	if err != nil {
		return answer, err
	}
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		columns := strings.Split(line, " ")
		for _, col := range columns {
			if col != "" {
				answer = append(answer, col)
				break
			}
		}
	}
	return answer, nil
}

// ShallowCloneBranch clones a single branch of the given git URL into the given directory
func ShallowCloneBranch(g Interface, gitURL string, branch string, dir string) error {
	verbose := true
	remoteName := "origin"
	_, err := g.Command(dir, "init")
	if err != nil {
		return errors.Wrapf(err, "failed to init a new git repository in directory %s", dir)
	}
	if verbose {
		log.Logger().Infof("ran git init in %s", dir)
	}
	err = AddRemote(g, dir, "origin", gitURL)
	if err != nil {
		return errors.Wrapf(err, "failed to add remote %s with url %s in directory %s", remoteName, gitURL, dir)
	}
	if verbose {
		log.Logger().Infof("ran git add remote %s %s in %s", remoteName, gitURL, dir)
	}

	_, err = g.Command(dir, "fetch", remoteName, branch, "--depth=1")
	if err != nil {
		return errors.Wrapf(err, "failed to fetch %s from %s in directory %s", branch, gitURL,
			dir)
	}
	_, err = g.Command(dir, "checkout", "-t", fmt.Sprintf("%s/%s", remoteName, branch))
	if err != nil {
		log.Logger().Warnf("failed to checkout remote tracking branch %s/%s in directory %s due to: %s", remoteName,
			branch, dir, err.Error())
		if branch != "master" {
			// git init checks out the master branch by default
			_, err = g.Command(dir, "branch", branch)
			if err != nil {
				return errors.Wrapf(err, "failed to create branch %s in directory %s", branch, dir)
			}

			if verbose {
				log.Logger().Infof("ran git branch %s in directory %s", branch, dir)
			}
		}
		_, err = g.Command(dir, "reset", "--hard", fmt.Sprintf("%s/%s", remoteName, branch))
		if err != nil {
			return errors.Wrapf(err, "failed to reset hard to %s in directory %s", branch, dir)
		}
		_, err = g.Command(dir, "branch", "--set-upstream-to", fmt.Sprintf("%s/%s", remoteName, branch), branch)
		if err != nil {
			return errors.Wrapf(err, "failed to set tracking information to %s/%s %s in directory %s", remoteName,
				branch, branch, dir)
		}
	}
	return nil
}

// AddRemote adds a remote repository at the given URL and with the given name
func AddRemote(g Interface, dir string, name string, url string) error {
	_, err := g.Command(dir, "remote", "add", name, url)
	if err != nil {
		_, err = g.Command(dir, "remote", "set-url", name, url)
		if err != nil {
			return err
		}
	}
	return nil
}

// Describe does a git describe of commitish, optionally adding the abbrev arg if not empty, falling back to just the commit ref if it's untagged
func Describe(g Interface, dir string, contains bool, commitish string, abbrev string, fallback bool) (string, string, error) {
	args := []string{"describe", commitish}
	if abbrev != "" {
		args = append(args, fmt.Sprintf("--abbrev=%s", abbrev))
	}
	if contains {
		args = append(args, "--contains")
	}
	out, err := g.Command(dir, args...)
	if err != nil {
		if fallback {
			// If the commit-ish is untagged, it'll fail with "fatal: cannot describe '<commit-ish>'". In those cases, just return
			// the original commit-ish.
			if strings.Contains(err.Error(), "fatal: cannot describe") {
				return commitish, "", nil
			}
		}
		return "", "", errors.Wrapf(err, "running git %s", strings.Join(args, " "))
	}
	trimmed := strings.TrimSpace(strings.Trim(out, "\n"))
	parts := splitDescribeRegex.Split(trimmed, -1)
	if len(parts) == 2 {
		return parts[0], parts[1], nil
	}
	return parts[0], "", nil
}
