package helmer

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"

	"github.com/google/uuid"
	"github.com/jenkins-x/jx-helpers/v3/pkg/files"
	"github.com/jenkins-x/jx-helpers/v3/pkg/termcolor"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"

	"sigs.k8s.io/yaml"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
)

// copied from helm to minimise dependencies...

// Dependency describes a chart upon which another chart depends.
//
// Dependencies can be used to express developer intent, or to capture the state
// of a chart.
type Dependency struct {
	// Name is the name of the dependency.
	//
	// This must mach the name in the dependency's Chart.yaml.
	Name string `json:"name"`
	// Version is the version (range) of this chart.
	//
	// A lock file will always produce a single version, while a dependency
	// may contain a semantic version range.
	Version string `json:"version,omitempty"`
	// The URL to the repository.
	//
	// Appending `index.yaml` to this string should result in a URL that can be
	// used to fetch the repository index.
	Repository string `json:"repository"`
	// A yaml path that resolves to a boolean, used for enabling/disabling charts (e.g. subchart1.enabled )
	Condition string `json:"condition,omitempty"`
	// Tags can be used to group charts for enabling/disabling together
	Tags []string `json:"tags,omitempty"`
	// Enabled bool determines if chart should be loaded
	Enabled bool `json:"enabled,omitempty"`
	// ImportValues holds the mapping of source values to parent key to be imported. Each item can be a
	// string or pair of child/parent sublist items.
	ImportValues []interface{} `json:"import-values,omitempty"`
	// Alias usable alias to be used for the chart
	Alias string `json:"alias,omitempty"`
}

// ErrNoRequirementsFile to detect error condition
type ErrNoRequirementsFile error

// Requirements is a list of requirements for a chart.
//
// Requirements are charts upon which this chart depends. This expresses
// developer intent.
type Requirements struct {
	Dependencies []*Dependency `json:"dependencies"`
}

// AddHelmRepoIfMissing will add the helm repo if there is no helm repo with that url present.
// It will generate the repoName from the url (using the host name) if the repoName is empty.
// The repo name may have a suffix added in order to prevent name collisions, and is returned for this reason.
// The username and password will be stored in vault for the URL (if vault is enabled).
func AddHelmRepoIfMissing(helmer Helmer, helmURL, repoName, username, password string) (string, error) {
	missing, existingName, err := helmer.IsRepoMissing(helmURL)
	if err != nil {
		return "", fmt.Errorf("failed to check if the repository with URL '%s' is missing: %w", helmURL, err)
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
			return "", fmt.Errorf("listing helm repos: %w", err)
		}
		baseName := repoName
		for i := 0; true; i++ {
			if _, exists := existingRepos[repoName]; exists {
				repoName = fmt.Sprintf("%s-%d", baseName, i)
			} else {
				break
			}
		}
		log.Logger().Debugf("Adding missing Helm repo: %s %s", termcolor.ColorInfo(repoName), termcolor.ColorInfo(helmURL))
		err = helmer.AddRepo(repoName, helmURL, username, password)
		if err != nil {
			return "", fmt.Errorf("failed to add the repository '%s' with URL '%s': %w", repoName, helmURL, err)
		}
		log.Logger().Debugf("Successfully added Helm repository %s.", repoName)
	} else {
		repoName = existingName
	}
	return repoName, nil
}

// DepSorter Used to avoid merge conflicts by sorting deps by name
type DepSorter []*Dependency

func (a DepSorter) Len() int           { return len(a) }
func (a DepSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a DepSorter) Less(i, j int) bool { return a[i].Name < a[j].Name }

// SetAppVersion sets the version of the app to use
func (r *Requirements) SetAppVersion(app string, version string, repository string, alias string) {
	if r.Dependencies == nil {
		r.Dependencies = []*Dependency{}
	}
	for _, dep := range r.Dependencies {
		if dep != nil && dep.Name == app {
			if version != dep.Version {
				dep.Version = version
			}
			if repository != "" {
				dep.Repository = repository
			}
			if alias != "" {
				dep.Alias = alias
			}
			return
		}
	}
	r.Dependencies = append(r.Dependencies, &Dependency{
		Name:       app,
		Version:    version,
		Repository: repository,
		Alias:      alias,
	})
	sort.Sort(DepSorter(r.Dependencies))
}

// RemoveApplication removes the given app name. Returns true if a dependency was removed
func (r *Requirements) RemoveApplication(app string) bool {
	for i, dep := range r.Dependencies {
		if dep != nil && dep.Name == app {
			r.Dependencies = append(r.Dependencies[:i], r.Dependencies[i+1:]...)
			sort.Sort(DepSorter(r.Dependencies))
			return true
		}
	}
	return false
}

// FindRequirementsFileName returns the default requirements.yaml file name
func FindRequirementsFileName(dir string) (string, error) {
	return findFileName(dir, RequirementsFileName)
}

// FindChartFileName returns the default chart.yaml file name
func FindChartFileName(dir string) (string, error) {
	return findFileName(dir, ChartFileName)
}

// FindValuesFileName returns the default values.yaml file name
func FindValuesFileName(dir string) (string, error) {
	return findFileName(dir, ValuesFileName)
}

// FindValuesFileNameForChart returns the values.yaml file name for a given chart within the environment or the default if the chart name is empty
func FindValuesFileNameForChart(dir string, chartName string) (string, error) {
	//Chart name and file name are joined here to avoid hard coding the environment
	//The chart name is ignored in the path if it's empty
	return findFileName(dir, filepath.Join(chartName, ValuesFileName))
}

// FindTemplatesDirName returns the default templates/ dir name
func FindTemplatesDirName(dir string) (string, error) {
	return findFileName(dir, TemplatesDirName)
}

func findFileName(dir string, fileName string) (string, error) {
	names := []string{
		filepath.Join(dir, DefaultEnvironmentChartDir, fileName),
		filepath.Join(dir, fileName),
	}
	for _, name := range names {
		exists, err := files.FileExists(name)
		if err != nil {
			return "", err
		}
		if exists {
			return name, nil
		}
	}
	myfiles, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, f := range myfiles {
		if f.IsDir() {
			name := filepath.Join(dir, f.Name(), fileName)
			exists, err := files.FileExists(name)
			if err != nil {
				return "", err
			}
			if exists {
				return name, nil
			}
		}
	}
	dirs := []string{
		filepath.Join(dir, DefaultEnvironmentChartDir),
		dir,
	}
	for _, d := range dirs {
		name := filepath.Join(d, fileName)
		exists, err := files.DirExists(d)
		if err != nil {
			return "", err
		}
		if exists {
			return name, nil
		}
	}
	return "", fmt.Errorf("could not deduce the default requirements.yaml file name")
}

// LoadRequirementsFile loads the requirements file or creates empty requirements if the file does not exist
func LoadRequirementsFile(fileName string) (*Requirements, error) {
	exists, err := files.FileExists(fileName)
	if err != nil {
		return nil, err
	}
	if exists {
		data, err := os.ReadFile(fileName)
		if err != nil {
			return nil, err
		}
		return LoadRequirements(data)
	}
	r := &Requirements{}
	return r, nil
}

// LoadChartFile loads the chart file or creates empty chart if the file does not exist
func LoadChartFile(fileName string) (*chart.Metadata, error) {
	exists, err := files.FileExists(fileName)
	if err != nil {
		return nil, err
	}
	if exists {
		data, err := os.ReadFile(fileName)
		if err != nil {
			return nil, err
		}
		return LoadChart(data)
	}
	return &chart.Metadata{}, nil
}

// LoadRequirements loads the requirements from some data
func LoadRequirements(data []byte) (*Requirements, error) {
	r := &Requirements{}
	return r, yaml.Unmarshal(data, r)
}

// LoadChart loads the requirements from some data
func LoadChart(data []byte) (*chart.Metadata, error) {
	r := &chart.Metadata{}
	return r, yaml.Unmarshal(data, r)
}

// LoadValues loads the values from some data
func LoadValues(data []byte) (map[string]interface{}, error) {
	r := map[string]interface{}{}
	if data == nil || len(data) == 0 {
		return r, nil
	}
	return r, yaml.Unmarshal(data, &r)
}

// SaveFile saves contents (a pointer to a data structure) to a file
func SaveFile(fileName string, contents interface{}) error {
	data, err := yaml.Marshal(contents)
	if err != nil {
		return fmt.Errorf("failed to marshal helm file %s: %w", fileName, err)
	}
	err = os.WriteFile(fileName, data, files.DefaultFileWritePermissions)
	if err != nil {
		return fmt.Errorf("failed to save helm file %s: %w", fileName, err)
	}
	return nil
}

func LoadChartName(chartFile string) (string, error) {
	chart, err := chartutil.LoadChartfile(chartFile)
	if err != nil {
		return "", err
	}
	return chart.Name, nil
}
