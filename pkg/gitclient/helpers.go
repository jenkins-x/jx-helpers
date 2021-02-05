package gitclient

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jenkins-x/jx-api/v4/pkg/util"
	"github.com/jenkins-x/jx-helpers/v3/pkg/files"
	"github.com/jenkins-x/jx-helpers/v3/pkg/stringhelpers"
	"github.com/jenkins-x/jx-helpers/v3/pkg/termcolor"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/pkg/errors"
)

var (
	splitDescribeRegex = regexp.MustCompile(`(?:~|\^|-g)`)
)

// Init inits a git repository into the given directory
func Init(g Interface, dir string) error {
	_, err := g.Command(dir, "init")
	if err != nil {
		return errors.Wrapf(err, "failed to initialise git in dir %s", dir)
	}
	return nil
}

// Add does a git add for all the given arguments
func Add(g Interface, dir string, args ...string) error {
	add := append([]string{"add"}, args...)
	_, err := g.Command(dir, add...)
	if err != nil {
		return errors.Wrapf(err, "failed to add %s to git", strings.Join(args, ", "))
	}
	return nil
}

// AddAndCommitFiles add and commits files
func AddAndCommitFiles(gitter Interface, dir, message string) (bool, error) {
	_, err := gitter.Command(dir, "add", "*")
	if err != nil {
		return false, errors.Wrapf(err, "failed to add files to git")
	}
	changes, err := HasChanges(gitter, dir)
	if err != nil {
		return changes, errors.Wrapf(err, "failed to check if there are changes")
	}
	if !changes {
		return changes, nil
	}
	_, err = gitter.Command(dir, "commit", "-m", message)
	if err != nil {
		return changes, errors.Wrapf(err, "failed to git commit initial code changes")
	}
	return changes, nil
}

// CreateBranch creates a dynamic branch name and branch
func CreateBranch(gitter Interface, dir string) (string, error) {
	branchName := fmt.Sprintf("pr-%s", uuid.New().String())
	gitRef := branchName
	_, err := gitter.Command(dir, "branch", branchName)
	if err != nil {
		return branchName, errors.Wrapf(err, "create branch %s from %s", branchName, gitRef)
	}

	_, err = gitter.Command(dir, "checkout", branchName)
	if err != nil {
		return branchName, errors.Wrapf(err, "checkout branch %s", branchName)
	}
	return branchName, nil
}

// CreateBranchFrom creates a new branch called branchName from startPoint
func CreateBranchFrom(g Interface, dir string, branchName string, startPoint string) error {
	_, err := g.Command(dir, "branch", branchName, startPoint)
	if err != nil {
		return errors.Wrapf(err, "failed to create branch %s from %s", branchName, startPoint)
	}
	return nil
}

// SetUpstreamTo will set the given branch to track the origin branch with the same name
func SetUpstreamTo(g Interface, dir string, branch string) error {
	upstream := fmt.Sprintf("origin/%s", branch)
	_, err := g.Command(dir, "branch", "--set-upstream-to", upstream, branch)
	if err != nil {
		return errors.Wrapf(err, "failed to set upstream to %s for branch %s", upstream, branch)
	}
	return nil
}

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
	var answer []string
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
	remoteName := "origin"
	_, err := g.Command(dir, "init")
	if err != nil {
		return errors.Wrapf(err, "failed to init a new git repository in directory %s", dir)
	}
	if true {
		log.Logger().Infof("ran git init in %s", dir)
	}
	err = AddRemote(g, dir, "origin", gitURL)
	if err != nil {
		return errors.Wrapf(err, "failed to add remote %s with url %s in directory %s", remoteName, gitURL, dir)
	}
	if true {
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

			if true {
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

// CloneToDir clones the git repository to either the given directory or create a temporary
func CloneToDir(g Interface, gitURL, dir string) (string, error) {
	var err error
	if dir != "" {
		err = os.MkdirAll(dir, util.DefaultWritePermissions)
		if err != nil {
			return "", errors.Wrapf(err, "failed to create directory %s", dir)
		}
	} else {
		dir, err = ioutil.TempDir("", "jx-git-")
		if err != nil {
			return "", errors.Wrap(err, "failed to create temporary directory")
		}
	}

	log.Logger().Debugf("cloning %s to directory %s", termcolor.ColorInfo(gitURL), termcolor.ColorInfo(dir))

	parentDir := filepath.Dir(dir)
	_, err = g.Command(parentDir, "clone", gitURL, dir)
	if err != nil {
		return "", errors.Wrapf(err, "failed to clone repository %s to directory: %s", gitURL, dir)
	}
	return dir, nil
}

// GetLatestCommitSha returns the latest commit sha
func GetLatestCommitSha(g Interface, dir string) (string, error) {
	return g.Command(dir, "rev-parse", "HEAD")
}

// ForcePushBranch does a force push of the local branch into the remote branch of the repository at the given directory
func ForcePushBranch(g Interface, dir string, localBranch string, remoteBranch string) error {
	fullBranch := fmt.Sprintf("%s:%s", localBranch, remoteBranch)
	return Push(g, dir, "origin", true, fullBranch)
}

// Push pushes the changes from the repository at the given directory
func Push(g Interface, dir string, remote string, force bool, refspec ...string) error {
	args := []string{"push", remote}
	if force {
		args = append(args, "--force")
	}
	args = append(args, refspec...)
	_, err := g.Command(dir, args...)
	if err != nil {
		return errors.Wrapf(err, "failed to push to the branch %v", refspec)
	}
	return nil
}

// Branch returns the current branch of the repository located at the given directory
func Branch(g Interface, dir string) (string, error) {
	return g.Command(dir, "rev-parse", "--abbrev-ref", "HEAD")
}

// CloneOrPull performs a clone if the directory is empty otherwise a pull
func CloneOrPull(g Interface, url string, dir string) error {
	empty, err := files.IsEmpty(dir)
	if err != nil {
		return err
	}

	if !empty {
		return Pull(g, dir)
	}
	_, err = CloneToDir(g, url, dir)
	if err != nil {
		return errors.Wrapf(err, "failed to clone %s to %s", url, dir)
	}
	return nil

}

// Pull performs a git pull
func Pull(g Interface, dir string) error {
	_, err := g.Command(dir, "pull")
	if err != nil {
		return errors.Wrapf(err, "failed to git pull in dir %s", dir)
	}
	return nil
}

// FetchTags fetches all the tags
func FetchTags(g Interface, dir string) error {
	_, err := g.Command(dir, "fetch", "--tags")
	if err != nil {
		return errors.Wrapf(err, "failed to fetch tags")
	}
	return nil
}

// FetchRemoteTags fetches all the tags from a remote repository
func FetchRemoteTags(g Interface, dir string, repo string) error {
	_, err := g.Command(dir, "fetch", repo, "--tags")
	if err != nil {
		return errors.Wrapf(err, "failed to fetch remote tags")
	}
	return nil
}

// Tags returns all tags from the repository at the given directory
func Tags(g Interface, dir string) ([]string, error) {
	return FilterTags(g, dir, "")
}

// FilterTags returns all tags from the repository at the given directory that match the filter
func FilterTags(g Interface, dir string, filter string) ([]string, error) {
	args := []string{"tag"}
	if filter != "" {
		args = append(args, "--list", filter)
	}
	text, err := g.Command(dir, args...)
	if err != nil {
		return nil, err
	}
	text = strings.TrimSuffix(text, "\n")
	split := strings.Split(text, "\n")
	// Split will return the original string if it can't split it, and it may be empty
	if len(split) == 1 && split[0] == "" {
		return make([]string, 0), nil
	}
	return split, nil
}

// FetchBranch fetches the refspecs from the repo
func FetchBranch(g Interface, dir string, repo string, refspecs ...string) error {
	return fetchBranch(g, dir, repo, false, false, false, refspecs...)
}

// FetchBranch fetches the refspecs from the repo
func fetchBranch(g Interface, dir string, repo string, unshallow bool, shallow bool,
	verbose bool, refspecs ...string) error {
	args := []string{"fetch", repo}
	if shallow && unshallow {
		return errors.Errorf("cannot use --depth=1 and --unshallow at the same time")
	}
	if shallow {
		args = append(args, "--depth=1")
	}
	if unshallow {
		args = append(args, "--unshallow")
	}
	for _, refspec := range refspecs {
		args = append(args, refspec)
	}
	_, err := g.Command(dir, args...)
	if err != nil {
		return errors.WithStack(err)
	}
	if verbose {
		if shallow {
			log.Logger().Infof("ran git fetch %s --depth=1 %s in dir %s", repo, strings.Join(refspecs, " "), dir)
		} else if unshallow {
			log.Logger().Infof("ran git fetch %s unshallow %s in dir %s", repo, strings.Join(refspecs, " "), dir)
		} else {
			log.Logger().Infof("ran git fetch %s --depth=1 %s in dir %s", repo, strings.Join(refspecs, " "), dir)
		}

	}
	return nil
}

// Checkout checks out the given branch
func Checkout(g Interface, dir string, branch string) error {
	_, err := g.Command(dir, "checkout", branch)
	if err != nil {
		return errors.Wrapf(err, "failed to checkout %s", branch)
	}
	return nil
}

// CheckoutRemoteBranch checks out the given remote tracking branch
func CheckoutRemoteBranch(g Interface, dir string, branch string) error {
	remoteBranch := "origin/" + branch
	remoteBranches, err := RemoteBranches(g, dir)
	if err != nil {
		return err
	}
	if stringhelpers.StringArrayIndex(remoteBranches, remoteBranch) < 0 {
		_, err := g.Command(dir, "checkout", "-t", remoteBranch)
		if err != nil {
			return errors.Wrapf(err, "failed to checkout %s", remoteBranch)
		}
		return nil
	}
	cur, err := Branch(g, dir)
	if err != nil {
		return err
	}
	if cur == branch {
		return nil
	}
	return Checkout(g, dir, branch)
}

// GetLatestCommitMessage returns the latest git commit message
func GetLatestCommitMessage(g Interface, dir string) (string, error) {
	return g.Command(dir, "log", "-1", "--pretty=%B")
}

// Remove removes the given file from a Git repository located at the given directory
func Remove(g Interface, dir, fileName string) error {
	_, err := g.Command(dir, "rm", "-r", fileName)
	if err != nil {
		return errors.Wrapf(err, "failed to remove %s in dir %s", fileName, dir)
	}
	return nil
}

// Status returns the status of the git repository at the given directory
func Status(g Interface, dir string) (string, error) {
	return g.Command(dir, "status")
}

// Merge merges the commitish into the current branch
func Merge(g Interface, dir string, commitish string) error {
	_, err := g.Command(dir, "merge", commitish)
	if err != nil {
		return errors.Wrapf(err, "failed to merge %s", commitish)
	}
	return nil
}
