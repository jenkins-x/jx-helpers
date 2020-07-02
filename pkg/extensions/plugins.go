package extensions

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	jenkinsv1client "github.com/jenkins-x/jx-api/pkg/client/clientset/versioned"
	"github.com/jenkins-x/jx-helpers/pkg/files"
	"github.com/jenkins-x/jx-helpers/pkg/httphelpers"
	"github.com/jenkins-x/jx-helpers/pkg/termcolor"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/jenkins-x/jx-logging/pkg/log"

	jenkinsv1 "github.com/jenkins-x/jx-api/pkg/apis/jenkins.io/v1"

	"github.com/spf13/cobra"
)

const (
	// PluginCommandLabel is the label applied to plugins to allow them to be found
	PluginCommandLabel = "jenkins.io/pluginCommand"
)

// PathVerifier receives a path and determines if it is valid or not
type PathVerifier interface {
	// Verify determines if a given path is valid
	Verify(path string) []error
}

// CommandOverrideVerifier verifies a set of plugins
type CommandOverrideVerifier struct {
	Root        *cobra.Command
	SeenPlugins map[string]string
}

// Verify implements PathVerifier and determines if a given path
// is valid depending on whether or not it overwrites an existing
// jx command path, or a previously seen plugin.
func (v *CommandOverrideVerifier) Verify(path string) []error {
	if v.Root == nil {
		return []error{fmt.Errorf("unable to verify path with nil root")}
	}

	// extract the plugin binary name
	segs := strings.Split(path, "/")
	binName := segs[len(segs)-1]

	cmdPath := strings.Split(binName, "-")
	if len(cmdPath) > 1 {
		// the first argument is always "jx" for a plugin binary
		cmdPath = cmdPath[1:]
	}

	var errors []error

	if isExec, err := isExecutable(path); err == nil && !isExec {
		errors = append(errors, fmt.Errorf("warning: %s identified as a jx plugin, but it is not executable", path))
	} else if err != nil {
		errors = append(errors, fmt.Errorf("error: unable to identify %s as an executable file: %v", path, err))
	}

	if existingPath, ok := v.SeenPlugins[binName]; ok {
		errors = append(errors, fmt.Errorf("warning: %s is overshadowed by a similarly named plugin: %s", path, existingPath))
	} else {
		v.SeenPlugins[binName] = path
	}

	if cmd, _, err := v.Root.Find(cmdPath); err == nil {
		errors = append(errors, fmt.Errorf("warning: %s overwrites existing command: %q", binName, cmd.CommandPath()))
	}

	return errors
}

func isExecutable(fullPath string) (bool, error) {
	info, err := os.Stat(fullPath)
	if err != nil {
		return false, err
	}

	if runtime.GOOS == "windows" {
		if strings.HasSuffix(info.Name(), ".exe") {
			return true, nil
		}
		return false, nil
	}

	if m := info.Mode(); !m.IsDir() && m&0111 != 0 {
		return true, nil
	}

	return false, nil
}

// FindPluginUrl finds the download URL for the current platform for a plugin
func FindPluginUrl(plugin jenkinsv1.PluginSpec) (string, error) {
	u := ""
	for _, binary := range plugin.Binaries {
		if strings.ToLower(runtime.GOOS) == strings.ToLower(binary.Goos) && strings.ToLower(runtime.
			GOARCH) == strings.ToLower(binary.Goarch) {
			u = binary.URL
		}
	}
	if u == "" {
		return "", fmt.Errorf("unable to locate binary for %s %s for %s", runtime.GOARCH, runtime.GOOS,
			plugin.SubCommand)
	}
	return u, nil
}

// EnsurePluginInstalled ensures that the correct version of a plugin is installed locally.
// It will clean up old versions.
func EnsurePluginInstalled(plugin jenkinsv1.Plugin, pluginBinDir string) (string, error) {
	return EnsurePluginInstalledForAliasFile(plugin, pluginBinDir, "")
}

// EnsurePluginInstalledForAliasFile ensures that the correct version of a plugin is installed locally.
// It will clean up old versions.
func EnsurePluginInstalledForAliasFile(plugin jenkinsv1.Plugin, pluginBinDir string, aliasFileName string) (string, error) {
	var err error
	version := plugin.Spec.Version
	path := filepath.Join(pluginBinDir, fmt.Sprintf("%s-%s", plugin.Spec.Name, version))
	if _, err = os.Stat(path); os.IsNotExist(err) {
		u, err := FindPluginUrl(plugin.Spec)
		if err != nil {
			return "", err
		}
		log.Logger().Infof("Installing plugin %s version %s for command %s from %s into %s", termcolor.ColorInfo(plugin.Spec.Name),
			termcolor.ColorInfo(version), termcolor.ColorInfo(fmt.Sprintf("jx %s", plugin.Spec.SubCommand)), termcolor.ColorInfo(u), pluginBinDir)

		// Look for other versions to cleanup
		fileObs, err := ioutil.ReadDir(pluginBinDir)
		if err != nil {
			return path, err
		}
		deleted := make([]string, 0)
		// lets only delete plugins for this major version so we can keep, say, helm 2 and 3 around
		prefix := plugin.Name + "-"
		if len(version) > 0 {
			prefix += version[0:1]
		}
		for _, f := range fileObs {
			if strings.HasPrefix(f.Name(), prefix) {
				err = os.Remove(filepath.Join(pluginBinDir, f.Name()))
				if err != nil {
					log.Logger().Warnf("Unable to delete old version of plugin %s installed at %s because %v", plugin.Name, f.Name(), err)
				} else {
					deleted = append(deleted, strings.TrimPrefix(f.Name(), fmt.Sprintf("%s-", plugin.Name)))
				}
			}
		}
		if len(deleted) > 0 {
			log.Logger().Infof("Deleted old plugin versions: %v", termcolor.ColorInfo(deleted))
		}

		httpClient := httphelpers.GetClientWithTimeout(time.Minute * 20)

		// Get the file
		pluginURL, err := url.Parse(u)
		if err != nil {
			return "", err
		}
		filename := filepath.Base(pluginURL.Path)
		tmpDir, err := ioutil.TempDir("", plugin.Spec.Name)
		defer func() {
			err := os.RemoveAll(tmpDir)
			if err != nil {
				log.Logger().Errorf("Error cleaning up tmpdir %s because %v", tmpDir, err)
			}
		}()
		if err != nil {
			return "", err
		}
		downloadFile := filepath.Join(tmpDir, filename)
		// Create the file
		out, err := os.Create(downloadFile)
		if err != nil {
			return path, err
		}
		defer out.Close()
		requestU := u
		if pluginURL.User != nil {
			c := *pluginURL
			c.User = nil
			requestU = c.String()
		}
		req, err := http.NewRequest("GET", requestU, nil)
		req.Header.Add("Accept", "application/octet-stream")
		if pluginURL.User != nil {
			pwd, ok := pluginURL.User.Password()
			if ok {
				req.Header.Add("Authorization", fmt.Sprintf("token %s", pwd))
			}
		}
		resp, err := httpClient.Do(req)
		if err != nil {
			return path, err
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return "", fmt.Errorf("unable to install plugin %s because %s getting %s", plugin.Name, resp.Status, u)
		}
		defer resp.Body.Close()

		// Write the body to file
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return path, err
		}

		oldPath := downloadFile
		if strings.HasSuffix(filename, ".tar.gz") || strings.HasSuffix(aliasFileName, ".tar.gz") {
			err = files.UnTargz(downloadFile, tmpDir, make([]string, 0))
			if err != nil {
				return "", err
			}
			oldPath = filepath.Join(tmpDir, plugin.Spec.Name)
		}
		if strings.HasSuffix(filename, ".zip") || strings.HasSuffix(aliasFileName, ".zip") {
			err = files.Unzip(downloadFile, tmpDir)
			if err != nil {
				return "", err
			}
			oldPath = filepath.Join(tmpDir, plugin.Spec.Name)
		}

		err = files.CopyFile(oldPath, path)
		if err != nil {
			return "", err
		}
		// Make the file executable
		err = os.Chmod(path, 0755)
		if err != nil {
			return path, err
		}
	}
	return path, nil
}

// ValidatePlugins tells the user about any problems with plugins installed
func ValidatePlugins(jxClient jenkinsv1client.Interface, ns string) error {
	// TODO needs a test
	// Validate installed plugins
	plugins, err := jxClient.JenkinsV1().Plugins(ns).List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	seenSubCommands := make(map[string][]jenkinsv1.Plugin, 0)
	for _, plugin := range plugins.Items {
		if _, ok := seenSubCommands[plugin.Spec.SubCommand]; !ok {
			seenSubCommands[plugin.Spec.SubCommand] = make([]jenkinsv1.Plugin, 0)
		}
		seenSubCommands[plugin.Spec.SubCommand] = append(seenSubCommands[plugin.Spec.SubCommand], plugin)
	}
	for subCommand, ps := range seenSubCommands {
		if len(ps) > 1 {
			log.Logger().Warnf("More than one extension has installed a plugin which will be called for jx %s. These extensions are:", termcolor.ColorWarning(subCommand))
			for _, p := range ps {
				for _, o := range p.ObjectMeta.OwnerReferences {
					if o.Kind == "Extension" {
						log.Logger().Warnf("  %s", termcolor.ColorWarning(o.Name))
					}
				}
			}
			log.Logger().Warnf("\nUnpredictable behavior will occur. Contact the extension authors and ask them to resolve the conflict.")
		}

	}
	return nil
}
