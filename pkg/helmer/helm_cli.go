package helmer

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/jenkins-x/jx-helpers/v3/pkg/cmdrunner"
	"github.com/jenkins-x/jx-helpers/v3/pkg/files"
	"github.com/jenkins-x/jx-helpers/v3/pkg/termcolor"

	"github.com/jenkins-x/jx-logging/v3/pkg/log"
)

// HelmCLI implements common helm actions based on helm CLI
type HelmCLI struct {
	Binary  string
	CWD     string
	Runner  cmdrunner.CommandRunner
	Command *cmdrunner.Command
	Debug   bool
}

// NewHelmCLI creates a new CLI
func NewHelmCLI(cwd string) *HelmCLI {
	return NewHelmCLIWithRunner(nil, "helm", cwd, false)
}

// NewHelmCLIWithRunner creates a new HelmCLI interface for the given runner
func NewHelmCLIWithRunner(runner cmdrunner.CommandRunner, binary, cwd string, debug bool) *HelmCLI {
	command := &cmdrunner.Command{
		Name: binary,
		Dir:  cwd,
	}
	if runner == nil {
		runner = cmdrunner.QuietCommandRunner
	}
	cli := &HelmCLI{
		Binary:  binary,
		CWD:     cwd,
		Command: command,
		Runner:  runner,
		Debug:   debug,
	}
	return cli
}

// SetCWD configures the common working directory of helm CLI
func (h *HelmCLI) SetCWD(dir string) {
	h.CWD = dir
}

// HelmBinary return the configured helm CLI
func (h *HelmCLI) HelmBinary() string {
	return h.Binary
}

// SetHelmBinary configure a new helm CLI
func (h *HelmCLI) SetHelmBinary(binary string) {
	h.Binary = binary
}

func (h *HelmCLI) runHelm(args ...string) error {
	h.Command.SetDir(h.CWD)
	h.Command.SetName(h.Binary)
	h.Command.SetArgs(args)
	_, err := h.Runner(h.Command)
	return err
}

func (h *HelmCLI) runHelmWithOutput(args ...string) (string, error) {
	h.Command.SetDir(h.CWD)
	h.Command.SetName(h.Binary)
	h.Command.SetArgs(args)
	return h.Runner(h.Command)
}

// Init executes the helm init command according with the given flags
func (h *HelmCLI) Init(clientOnly bool, serviceAccount, tillerNamespace string, upgrade bool) error {
	var args []string
	args = append(args, "init")
	if clientOnly {
		args = append(args, "--client-only")
	}
	if serviceAccount != "" {
		args = append(args, "--service-account", serviceAccount)
	}
	if tillerNamespace != "" {
		args = append(args, "--tiller-namespace", tillerNamespace)
	}
	if upgrade {
		args = append(args, "--upgrade", "--wait", "--force-upgrade")
	}

	if h.Debug {
		log.Logger().Debugf("Initialising Helm '%s'", termcolor.ColorInfo(strings.Join(args, " ")))
	}

	return h.runHelm(args...)
}

// AddRepo adds a new helm repo with the given name and URL
func (h *HelmCLI) AddRepo(repoName, repoURL, username, password string) error {
	args := []string{"repo", "add", repoName, repoURL}
	if username != "" {
		args = append(args, "--username", username)
	}
	if password != "" {
		args = append(args, "--password", password)
	}
	return h.runHelm(args...)
}

// RemoveRepo removes the given repo from helm
func (h *HelmCLI) RemoveRepo(repo string) error {
	return h.runHelm("repo", "remove", repo)
}

// ListRepos list the installed helm repos together with their URL
func (h *HelmCLI) ListRepos() (map[string]string, error) {
	output, err := h.runHelmWithOutput("repo", "list")
	repos := map[string]string{}
	if err != nil {
		// helm3 now returns an error if there are no repos
		return repos, nil
		// return nil, shouldError.Wrap(err, "failed to list repositories")
	}
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		fields := strings.Fields(line)
		if len(fields) > 1 {
			repos[strings.TrimSpace(fields[0])] = fields[1]
		} else if len(fields) > 0 {
			repos[fields[0]] = ""
		}
	}
	return repos, nil
}

// SearchCharts searches for all the charts matching the given filter
func (h *HelmCLI) SearchCharts(filter string, allVersions bool) ([]ChartSummary, error) {
	var answer []ChartSummary
	args := []string{"search", "repo", filter}
	if allVersions {
		args = append(args, "--versions")
	}
	output, err := h.runHelmWithOutput(args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search charts: %w", err)
	}
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "NAME") || line == "" {
			continue
		}
		line = strings.TrimSpace(line)
		fields := strings.Split(line, "\t")
		chart := ChartSummary{}
		l := len(fields)
		if l == 0 {
			continue
		}
		chart.Name = strings.TrimSpace(fields[0])
		if l > 1 {
			chart.ChartVersion = strings.TrimSpace(fields[1])
		}
		if l > 2 {
			chart.AppVersion = strings.TrimSpace(fields[2])
		}
		if l > 3 {
			chart.Description = strings.TrimSpace(fields[3])
		}
		answer = append(answer, chart)
	}
	return answer, nil
}

// IsRepoMissing checks if the repository with the given URL is missing from helm.
// If the repo is found, the name of the repo will be returned
func (h *HelmCLI) IsRepoMissing(repoURL string) (bool, string, error) {
	repos, err := h.ListRepos()
	if err != nil {
		return true, "", fmt.Errorf("failed to list the repositories: %w", err)
	}
	searchedURL, err := url.Parse(repoURL)
	if err != nil {
		return true, "", fmt.Errorf("provided repo IR: is invalid: %w", err)
	}
	for name, repoURL := range repos {
		if len(repoURL) > 0 {
			ru, err := url.Parse(repoURL)
			if err != nil {
				return true, "", fmt.Errorf("failed to parse the repo URL: %w", err)
			}
			// match on the whole repoURL as helm dep build requires on username + passowrd in the URL
			if ru.Host == searchedURL.Host && ru.Path == searchedURL.Path {
				return false, name, nil
			}
		}
	}
	return true, "", nil
}

// UpdateRepo updates the helm repositories
func (h *HelmCLI) UpdateRepo() error {
	return h.runHelm("repo", "update")
}

// RemoveRequirementsLock removes the requirements.lock file from the current working directory
func (h *HelmCLI) RemoveRequirementsLock() error {
	dir := h.CWD
	path := filepath.Join(dir, "requirements.lock")
	exists, err := files.FileExists(path)
	if err != nil {
		return fmt.Errorf("no requirements.lock file found in directory '%s': %w", dir, err)
	}
	if exists {
		err = os.Remove(path)
		if err != nil {
			return fmt.Errorf("failed to remove the requirements.lock file: %w", err)
		}
	}
	return nil
}

// BuildDependency builds the helm dependencies of the helm chart from the current working directory
func (h *HelmCLI) BuildDependency() error {
	if h.Debug {
		log.Logger().Infof("Running %s dependency build in %s\n", h.Binary, termcolor.ColorInfo(h.CWD))
		out, err := h.runHelmWithOutput("dependency", "build")
		log.Logger().Infof(out)
		return err
	}
	return h.runHelm("dependency", "build")
}

// InstallChart installs a helm chart according with the given flags
func (h *HelmCLI) InstallChart(chart, releaseName, ns, version string, timeout int,
	values, valueStrings, valueFiles []string, repo, username, password string,
) error {
	var err error
	var args []string
	args = append(args, "install", "--wait", "--name", releaseName, "--namespace", ns, chart)
	repo, err = addUsernamePasswordToURL(repo, username, password)
	if err != nil {
		return err
	}

	if timeout != -1 {
		args = append(args, "--timeout", fmt.Sprintf("%ss", strconv.Itoa(timeout)))
	}
	if version != "" {
		args = append(args, "--version", version)
	}
	for _, value := range values {
		args = append(args, "--set", value)
	}
	for _, value := range valueStrings {
		args = append(args, "--set-string", value)
	}
	for _, valueFile := range valueFiles {
		args = append(args, "--values", valueFile)
	}
	if repo != "" {
		args = append(args, "--repo", repo)
	}
	if username != "" {
		args = append(args, "--username", username)
	}
	if password != "" {
		args = append(args, "--password", password)
	}
	logLevel := os.Getenv("JX_HELM_VERBOSE")
	if logLevel != "" {
		args = append(args, "-v", logLevel)
	}
	if h.Debug {
		log.Logger().Infof("Installing Chart '%s'", termcolor.ColorInfo(strings.Join(args, " ")))
	}

	err = h.runHelm(args...)
	if err != nil {
		return err
	}

	return nil
}

// FetchChart fetches a Helm Chart
func (h *HelmCLI) FetchChart(chart, version string, untar bool, untardir, repo,
	username, password string,
) error {
	var args []string
	args = append(args, "fetch", chart)
	repo, err := addUsernamePasswordToURL(repo, username, password)
	if err != nil {
		return err
	}

	if untardir != "" {
		args = append(args, "--untardir", untardir)
	}
	if untar {
		args = append(args, "--untar")
	}

	if username != "" {
		args = append(args, "--username", username)
	}
	if password != "" {
		args = append(args, "--password", password)
	}

	if version != "" {
		args = append(args, "--version", version)
	}

	if repo != "" {
		args = append(args, "--repo", repo)
	}

	if h.Debug {
		log.Logger().Infof("Fetching Chart '%s'", termcolor.ColorInfo(strings.Join(args, " ")))
	}

	return h.runHelm(args...)
}

// Template generates the YAML from the chart template to the given directory
func (h *HelmCLI) Template(chart, releaseName, ns, outDir string, upgrade bool,
	values, valueStrings, valueFiles []string,
) error {
	args := []string{"template", "--name", releaseName, "--namespace", ns, chart, "--output-dir", outDir, "--debug"}
	if upgrade {
		args = append(args, "--is-upgrade")
	}
	for _, value := range values {
		args = append(args, "--set", value)
	}
	for _, value := range valueStrings {
		args = append(args, "--set-string", value)
	}
	for _, valueFile := range valueFiles {
		args = append(args, "--values", valueFile)
	}

	if h.Debug {
		log.Logger().Debugf("Generating Chart Template '%s'", termcolor.ColorInfo(strings.Join(args, " ")))
	}
	err := h.runHelm(args...)
	if err != nil {
		return fmt.Errorf("Failed to run helm %s: %w", strings.Join(args, " "), err)
	}
	return err
}

// UpgradeChart upgrades a helm chart according with given helm flags
func (h *HelmCLI) UpgradeChart(chart, releaseName, ns, version string,
	install bool, timeout int, force, wait bool, values, valueStrings,
	valueFiles []string, repo, username, password string,
) error {
	var err error
	var args []string
	args = append(args, "upgrade", "--namespace", ns)
	repo, err = addUsernamePasswordToURL(repo, username, password)
	if err != nil {
		return err
	}

	if install {
		args = append(args, "--install")
	}
	if wait {
		args = append(args, "--wait")
	}
	if force {
		args = append(args, "--force")
	}
	if timeout != -1 {
		if h.Binary == "helm3" {
			args = append(args, "--timeout", fmt.Sprintf("%ss", strconv.Itoa(timeout)))
		} else {
			args = append(args, "--timeout", strconv.Itoa(timeout))
		}
	}
	if version != "" {
		args = append(args, "--version", version)
	}
	for _, value := range values {
		args = append(args, "--set", value)
	}
	for _, value := range valueStrings {
		args = append(args, "--set-string", value)
	}
	for _, valueFile := range valueFiles {
		args = append(args, "--values", valueFile)
	}
	if repo != "" {
		args = append(args, "--repo", repo)
	}
	if username != "" {
		args = append(args, "--username", username)
	}
	if password != "" {
		args = append(args, "--password", password)
	}
	logLevel := os.Getenv("JX_HELM_VERBOSE")
	if logLevel != "" {
		args = append(args, "-v", logLevel)
	}
	args = append(args, releaseName, chart)

	if h.Debug {
		log.Logger().Infof("Upgrading Chart '%s'", termcolor.ColorInfo(strings.Join(args, " ")))
	}

	err = h.runHelm(args...)
	if err != nil {
		return err
	}

	return nil
}

// DeleteRelease removes the given release
func (h *HelmCLI) DeleteRelease(ns, releaseName string, purge bool) error {
	var args []string
	args = append(args, "delete")
	if purge {
		args = append(args, "--purge")
	}
	args = append(args, releaseName)
	return h.runHelm(args...)
}

// ListReleases lists the releases in ns
func (h *HelmCLI) ListReleases(ns string) (map[string]ReleaseSummary, []string, error) {
	output, err := h.runHelmWithOutput("list", "--all", "--namespace", ns)
	if err != nil {
		return nil, nil, fmt.Errorf("running helm list --all --namespace %s: %w", ns, err)
	}
	lines := strings.Split(strings.TrimSpace(output), "\n")
	result := make(map[string]ReleaseSummary)
	keys := make([]string, 0)
	if len(lines) > 1 {
		if h.Binary == "helm" {
			for _, line := range lines[1:] {
				fields := strings.Fields(line)
				if len(fields) == 10 || len(fields) == 11 {
					chartFullName := fields[8]
					lastDash := strings.LastIndex(chartFullName, "-")
					releaseName := fields[0]
					keys = append(keys, releaseName)
					result[releaseName] = ReleaseSummary{
						ReleaseName: fields[0],
						Revision:    fields[1],
						Updated: fmt.Sprintf("%s %s %s %s %s", fields[2], fields[3], fields[4], fields[5],
							fields[6]),
						Status:        fields[7],
						ChartFullName: chartFullName,
						Namespace:     ns,
						Chart:         chartFullName[:lastDash],
						ChartVersion:  chartFullName[lastDash+1:],
					}
				} else {
					return nil, nil, fmt.Errorf("Cannot parse %s as helm list output", line)
				}
			}
		} else {
			for _, line := range lines[1:] {
				fields := strings.Fields(line)
				if len(fields) == 9 {
					chartFullName := fields[8]
					lastDash := strings.LastIndex(chartFullName, "-")
					releaseName := fields[0]
					keys = append(keys, releaseName)
					result[releaseName] = ReleaseSummary{
						ReleaseName:   fields[0],
						Revision:      fields[2],
						Updated:       fmt.Sprintf("%s %s %s %s", fields[3], fields[4], fields[5], fields[6]),
						Status:        strings.ToUpper(fields[7]),
						ChartFullName: chartFullName,
						Namespace:     ns,
						Chart:         chartFullName[:lastDash],
						ChartVersion:  chartFullName[lastDash+1:],
					}
				} else {
					return nil, nil, fmt.Errorf("Cannot parse %s as helm3 list output", line)
				}
			}
		}
	}
	sort.Strings(keys)
	return result, keys, nil
}

// FindChart find a chart in the current working directory, if no chart file is found an error is returned
func (h *HelmCLI) FindChart() (string, error) {
	dir := h.CWD
	chartFile := filepath.Join(dir, ChartFileName)
	exists, err := files.FileExists(chartFile)
	if err != nil {
		return "", fmt.Errorf("no Chart.yaml file found in directory '%s': %w", dir, err)
	}
	if !exists {
		files, err := filepath.Glob(filepath.Join(dir, "*", "Chart.yaml"))
		if err != nil {
			return "", fmt.Errorf("no Chart.yaml file found: %w", err)
		}
		if len(files) > 0 {
			chartFile = files[0]
		} else {
			files, err = filepath.Glob(filepath.Join(dir, "*", "*", "Chart.yaml"))
			if err != nil {
				return "", fmt.Errorf("no Chart.yaml file found: %w", err)
			}
			if len(files) > 0 {
				for _, file := range files {
					if !strings.HasSuffix(file, "/preview/Chart.yaml") {
						return file, nil
					}
				}
			}
		}
	}
	return chartFile, nil
}

// StatusRelease returns the output of the helm status command for a given release
func (h *HelmCLI) StatusRelease(ns, releaseName string) error {
	return h.runHelm("status", releaseName)
}

// StatusReleaseWithOutput returns the output of the helm status command for a given release
func (h *HelmCLI) StatusReleaseWithOutput(ns, releaseName, outputFormat string) (string, error) {
	if outputFormat == "" {
		return h.runHelmWithOutput("status", releaseName)
	}
	return h.runHelmWithOutput("status", releaseName, "--output", outputFormat)
}

// Lint lints the helm chart from the current working directory and returns the warnings in the output
func (h *HelmCLI) Lint(valuesFiles []string) (string, error) {
	args := []string{
		"lint",
		"--set", "tags.jx-lint=true",
		"--set", "global.jxLint=true",
		"--set-string", "global.jxTypeEnv=lint",
	}
	for _, valueFile := range valuesFiles {
		if valueFile != "" {
			args = append(args, "--values", valueFile)
		}
	}
	return h.runHelmWithOutput(args...)
}

// Env returns the environment variables for the helmer
func (h *HelmCLI) Env() map[string]string {
	return h.Command.CurrentEnv()
}

// Version executes the helm version command and returns its output
func (h *HelmCLI) Version(tls bool) (string, error) {
	versionString, err := h.VersionWithArgs(tls)
	if err != nil {
		return "", fmt.Errorf("unable to query helm version: %w", err)
	}

	return h.extractSemanticVersion(versionString)
}

// VersionWithArgs executes the helm version command and returns its output
func (h *HelmCLI) VersionWithArgs(tls bool, extraArgs ...string) (string, error) {
	args := []string{"version", "--short", "--client"}
	if tls {
		args = append(args, "--tls")
	}
	args = append(args, extraArgs...)
	return h.runHelmWithOutput(args...)
}

// PackageChart packages the chart from the current working directory
func (h *HelmCLI) PackageChart() error {
	return h.runHelm("package", h.CWD)
}

// DecryptSecrets decrypt secrets
func (h *HelmCLI) DecryptSecrets(location string) error {
	return h.runHelm("secrets", "dec", location)
}

// Helm really prefers to have the username and password embedded in the URL (ugh) so this
// function makes that happen
func addUsernamePasswordToURL(urlStr, username, password string) (string, error) {
	if urlStr != "" && username != "" && password != "" {
		u, err := url.Parse(urlStr)
		if err != nil {
			return "", err
		}
		u.User = url.UserPassword(username, password)
		return u.String(), nil
	}
	return urlStr, nil
}

// extractSemanticVersion tries to extract a semantic version string substring from the specified string
func (h *HelmCLI) extractSemanticVersion(versionString string) (string, error) {
	r := regexp.MustCompile(`.*v?(?P<SemVer>\d+\.\d+\.\d+).*`)
	match := r.FindStringSubmatch(versionString)
	if match == nil {
		return "", fmt.Errorf("unable to extract a semantic version from %s", versionString)
	}

	for i, name := range r.SubexpNames() {
		if name == "SemVer" {
			return match[i], nil
		}
	}

	return "", fmt.Errorf("unable to extract a semantic version from %s", versionString)
}
