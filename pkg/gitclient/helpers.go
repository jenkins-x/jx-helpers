package gitclient

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jenkins-x/jx-helpers/pkg/files"
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

// HasChanges indicates if there are any changes in the repository from the given directory
func HasChanges(g Interface, dir string) (bool, error) {
	text, err := g.Command(dir, "status", "-s")
	if err != nil {
		return false, err
	}
	text = strings.TrimSpace(text)
	return len(text) > 0, nil
}

// HasFileChanged indicates if there are any changes to a file in the repository from the given directory
func HasFileChanged(g Interface, dir string, fileName string) (bool, error) {
	text, err := g.Command(dir, "status", "-s", fileName)
	if err != nil {
		return false, err
	}
	text = strings.TrimSpace(text)
	return len(text) > 0, nil
}

// CommiIfChanges does a commit if there are any changes in the repository at the given directory
func CommitIfChanges(g Interface, dir string, message string) error {
	changed, err := HasChanges(g, dir)
	if err != nil {
		return err
	}
	if !changed {
		return nil
	}
	_, err = g.Command(dir, "commit", "-m", message)
	if err != nil {
		return errors.Wrap(err, "failed to commit to git")
	}
	return nil
}

// FindGitConfigDir tries to find the `.git` directory either in the current directory or in parent directories
func FindGitConfigDir(dir string) (string, string, error) {
	d := dir
	var err error
	if d == "" {
		d, err = os.Getwd()
		if err != nil {
			return "", "", err
		}
	}
	for {
		gitDir := filepath.Join(d, ".git/config")
		exists, err := files.FileExists(gitDir)
		if err != nil {
			return "", "", err
		}
		if exists {
			return d, gitDir, nil
		}
		dirPath := strings.TrimSuffix(d, "/")
		if dirPath == "" {
			return "", "", nil
		}
		p, _ := filepath.Split(dirPath)
		if d == "/" || p == d {
			return "", "", nil
		}
		d = p
	}
}

// GetCommitPointedToByLatestTag return the SHA of the commit pointed to by the latest git tag as well as the tag name
// for the git repo in dir
func GetCommitPointedToByLatestTag(g Interface, dir string) (string, string, error) {
	tagSHA, tagName, err := NthTag(g, dir, 1)
	if err != nil {
		return "", "", errors.Wrapf(err, "getting commit pointed to by latest tag in %s", dir)
	}
	if tagSHA == "" {
		return tagSHA, tagName, nil
	}
	commitSHA, err := g.Command(dir, "rev-list", "-n", "1", tagSHA)
	if err != nil {
		return "", "", errors.Wrapf(err, "running for git rev-list -n 1 %s", tagSHA)
	}
	return commitSHA, tagName, err
}

// NthTag return the SHA and tag name of nth tag in reverse chronological order from the repository at the given directory.
// If the nth tag does not exist empty strings without an error are returned.
func NthTag(g Interface, dir string, n int) (string, string, error) {
	out, err := g.Command(dir, "for-each-ref", "--sort=-creatordate",
		"--format=%(objectname)%00%(refname:short)", fmt.Sprintf("--count=%d", n), "refs/tags")
	if err != nil {
		return "", "", errors.Wrapf(err, "running git")
	}
	tagList := strings.Split(out, "\n")

	if len(tagList) < n {
		return "", "", nil
	}

	fields := strings.Split(tagList[n-1], "\x00")

	if len(fields) != 2 {
		return "", "", errors.Errorf("Unexpected format for returned tag and sha: '%s'", tagList[n-1])
	}

	return fields[0], fields[1], nil
}
