package jxenv

import (
	"fmt"
	"os/user"
	"sort"
	"strings"

	"github.com/jenkins-x/jx-helpers/pkg/kube"
	"github.com/pkg/errors"

	v1 "github.com/jenkins-x/jx-api/pkg/apis/jenkins.io/v1"
	"github.com/jenkins-x/jx-api/pkg/client/clientset/versioned"
	"github.com/jenkins-x/jx-logging/pkg/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var useForkForEnvGitRepo = false

// ResolveChartMuseumURLFn used to resolve the chart repository URL if using remote environments
type ResolveChartMuseumURLFn func() (string, error)

// GetDevEnvTeamSettings gets the team settings from the specified namespace.
func GetDevEnvTeamSettings(jxClient versioned.Interface, ns string) (*v1.TeamSettings, error) {
	devEnv, err := GetDevEnvironment(jxClient, ns)
	if err != nil {
		log.Logger().Errorf("Error loading team settings. %v", err)
		return nil, err
	}
	if devEnv != nil {
		return &devEnv.Spec.TeamSettings, nil
	}
	return nil, fmt.Errorf("unable to find development environment in %s to get team settings", ns)
}

// GetDevEnvGitOwner gets the default GitHub owner/organisation to use for Environment repos. This takes the setting
// from the 'jx' Dev Env to get the one that was selected at installation time.
func GetDevEnvGitOwner(jxClient versioned.Interface) (string, error) {
	adminDevEnv, err := GetDevEnvironment(jxClient, "jx")
	if err != nil {
		log.Logger().Errorf("Error loading team settings. %v", err)
		return "", err
	}
	if adminDevEnv != nil {
		return adminDevEnv.Spec.TeamSettings.EnvOrganisation, nil
	}
	return "", errors.New("Unable to find development environment in 'jx' to take git owner from")
}

// GetEnvironmentNames returns the sorted list of environment names
func GetEnvironmentNames(jxClient versioned.Interface, ns string) ([]string, error) {
	var envNames []string
	envs, err := jxClient.JenkinsV1().Environments(ns).List(metav1.ListOptions{})
	if err != nil {
		return envNames, err
	}
	SortEnvironments(envs.Items)
	for _, env := range envs.Items {
		n := env.Name
		if n != "" {
			envNames = append(envNames, n)
		}
	}
	sort.Strings(envNames)
	return envNames, nil
}

func IsPreviewEnvironment(env *v1.Environment) bool {
	return env != nil && env.Spec.Kind == v1.EnvironmentKindTypePreview
}

// GetFilteredEnvironmentNames returns the sorted list of environment names
func GetFilteredEnvironmentNames(jxClient versioned.Interface, ns string, fn func(environment *v1.Environment) bool) ([]string, error) {
	var envNames []string
	envs, err := jxClient.JenkinsV1().Environments(ns).List(metav1.ListOptions{})
	if err != nil {
		return envNames, err
	}
	SortEnvironments(envs.Items)
	for _, e := range envs.Items {
		env := e
		n := env.Name
		if n != "" && fn(&env) {
			envNames = append(envNames, n)
		}
	}
	sort.Strings(envNames)
	return envNames, nil
}

// GetOrderedEnvironments returns a map of the environments along with the correctly ordered  names
func GetOrderedEnvironments(jxClient versioned.Interface, ns string) (map[string]*v1.Environment, []string, error) {
	m := map[string]*v1.Environment{}

	var envNames []string
	envs, err := jxClient.JenkinsV1().Environments(ns).List(metav1.ListOptions{})
	if err != nil {
		return m, envNames, err
	}
	SortEnvironments(envs.Items)
	for _, env := range envs.Items {
		n := env.Name
		c := env
		m[n] = &c
		if n != "" {
			envNames = append(envNames, n)
		}
	}
	return m, envNames, nil
}

// GetEnvironments returns a map of the environments along with a sorted list of names
func GetEnvironments(jxClient versioned.Interface, ns string) (map[string]*v1.Environment, []string, error) {
	m := map[string]*v1.Environment{}

	var envNames []string
	envs, err := jxClient.JenkinsV1().Environments(ns).List(metav1.ListOptions{})
	if err != nil {
		return m, envNames, err
	}
	for _, env := range envs.Items {
		n := env.Name
		c := env
		m[n] = &c
		if n != "" {
			envNames = append(envNames, n)
		}
	}
	sort.Strings(envNames)
	return m, envNames, nil
}

// GetEnvironment find an environment by name
func GetEnvironment(jxClient versioned.Interface, ns string, name string) (*v1.Environment, error) {
	envs, err := jxClient.JenkinsV1().Environments(ns).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, env := range envs.Items {
		if env.GetName() == name {
			return &env, nil
		}
	}
	return nil, fmt.Errorf("no environment with name '%s' found", name)
}

// GetEnvironmentsByPrURL find an environment by a pull request URL
func GetEnvironmentsByPrURL(jxClient versioned.Interface, ns string, prURL string) (*v1.Environment, error) {
	envs, err := jxClient.JenkinsV1().Environments(ns).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, env := range envs.Items {
		if env.Spec.PullRequestURL == prURL {
			return &env, nil
		}
	}
	return nil, fmt.Errorf("no environment found for PR '%s'", prURL)
}

// GetEnvironments returns the namespace name for a given environment
func GetEnvironmentNamespace(jxClient versioned.Interface, ns, environment string) (string, error) {
	env, err := jxClient.JenkinsV1().Environments(ns).Get(environment, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	if env == nil {
		return "", fmt.Errorf("no environment found called %s, try running `jx get env`", environment)
	}
	return env.Spec.Namespace, nil
}

// GetEditEnvironmentNamespace returns the namespace of the current users edit environment
func GetEditEnvironmentNamespace(jxClient versioned.Interface, ns string) (string, error) {
	envs, err := jxClient.JenkinsV1().Environments(ns).List(metav1.ListOptions{})
	if err != nil {
		return "", err
	}
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	for _, env := range envs.Items {
		if env.Spec.Kind == v1.EnvironmentKindTypeEdit && env.Spec.PreviewGitSpec.User.Username == u.Username {
			return env.Spec.Namespace, nil
		}
	}
	return "", fmt.Errorf("the user %s does not have an Edit environment in home namespace %s", u.Username, ns)
}

// GetDevNamespace returns the developer environment namespace
// which is the namespace that contains the Environments and the developer tools like Jenkins
func GetDevNamespace(kubeClient kubernetes.Interface, ns string) (string, string, error) {
	env := ""
	namespace, err := kubeClient.CoreV1().Namespaces().Get(ns, metav1.GetOptions{})
	if err != nil {
		return ns, env, err
	}
	if namespace == nil {
		return ns, env, fmt.Errorf("no namespace found for %s", ns)
	}
	if namespace.Labels != nil {
		answer := namespace.Labels[kube.LabelTeam]
		if answer != "" {
			ns = answer
		}
		env = namespace.Labels[kube.LabelEnvironment]
	}
	return ns, env, nil
}

// GetTeams returns the Teams the user is a member of
func GetTeams(kubeClient kubernetes.Interface) ([]*corev1.Namespace, []string, error) {
	var names []string
	var answer []*corev1.Namespace
	namespaceList, err := kubeClient.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return answer, names, err
	}
	for idx, namespace := range namespaceList.Items {
		if namespace.Labels[kube.LabelEnvironment] == kube.LabelValueDevEnvironment {
			answer = append(answer, &namespaceList.Items[idx])
			names = append(names, namespace.Name)
		}
	}
	sort.Strings(names)
	return answer, names, nil
}

type ByOrder []v1.Environment

func (a ByOrder) Len() int      { return len(a) }
func (a ByOrder) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByOrder) Less(i, j int) bool {
	env1 := a[i]
	env2 := a[j]
	o1 := env1.Spec.Order
	o2 := env2.Spec.Order
	if o1 == o2 {
		return env1.Name < env2.Name
	}
	return o1 < o2
}

func SortEnvironments(environments []v1.Environment) {
	sort.Sort(ByOrder(environments))
}

// ByTimestamp is used to fileter a list of PipelineActivities by their given timestamp
type ByTimestamp []v1.PipelineActivity

func (a ByTimestamp) Len() int      { return len(a) }
func (a ByTimestamp) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByTimestamp) Less(i, j int) bool {
	act1 := a[i]
	act2 := a[j]
	t1 := act1.Spec.StartedTimestamp
	if t1 == nil {
		return false
	}
	t2 := act2.Spec.StartedTimestamp
	if t2 == nil {
		return true
	}

	return t1.Before(t2)
}

// SortActivities sorts a list of PipelineActivities
func SortActivities(activities []v1.PipelineActivity) {
	sort.Sort(ByTimestamp(activities))
}

// NewPermanentEnvironment creates a new permanent environment for testing
func NewPermanentEnvironment(name string) *v1.Environment {
	return &v1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "jx",
		},
		Spec: v1.EnvironmentSpec{
			Label:             strings.Title(name),
			Namespace:         "jx-" + name,
			PromotionStrategy: v1.PromotionStrategyTypeAutomatic,
			Kind:              v1.EnvironmentKindTypePermanent,
		},
	}
}

// NewPermanentEnvironment creates a new permanent environment for testing
func NewPermanentEnvironmentWithGit(name string, gitUrl string) *v1.Environment {
	env := NewPermanentEnvironment(name)
	env.Spec.Source.URL = gitUrl
	env.Spec.Source.Ref = "master"
	return env
}

// NewPreviewEnvironment creates a new preview environment for testing
func NewPreviewEnvironment(name string) *v1.Environment {
	return &v1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "jx",
		},
		Spec: v1.EnvironmentSpec{
			Label:             strings.Title(name),
			Namespace:         "jx-preview-" + name,
			PromotionStrategy: v1.PromotionStrategyTypeAutomatic,
			Kind:              v1.EnvironmentKindTypePreview,
		},
	}
}

// GetDevEnvironment returns the current development environment using the jxClient for the given ns.
// If the Dev Environment cannot be found, returns nil Environment (rather than an error). A non-nil error is only
// returned if there is an error fetching the Dev Environment.
func GetDevEnvironment(jxClient versioned.Interface, ns string) (*v1.Environment, error) {
	//Find the settings for the team
	environmentInterface := jxClient.JenkinsV1().Environments(ns)
	name := kube.LabelValueDevEnvironment
	answer, err := environmentInterface.Get(name, metav1.GetOptions{})
	if err == nil {
		return answer, nil
	}
	selector := "env=dev"
	envList, err := environmentInterface.List(metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, err
	}
	if len(envList.Items) == 1 {
		return &envList.Items[0], nil
	}
	if len(envList.Items) == 0 {
		return nil, nil
	}
	return nil, fmt.Errorf("error fetching dev environment resource definition in namespace %s, No Environment called: %s or with selector: %s found %d entries: %v",
		ns, name, selector, len(envList.Items), envList.Items)
}

// GetPreviewEnvironmentReleaseName returns the (helm) release name for the given (preview) environment
// or the empty string is the environment is not a preview environment, or has no release name associated with it
func GetPreviewEnvironmentReleaseName(env *v1.Environment) string {
	if !IsPreviewEnvironment(env) {
		return ""
	}
	return env.Annotations[kube.AnnotationReleaseName]
}

// IsPermanentEnvironment indicates if an environment is permanent
func IsPermanentEnvironment(env *v1.Environment) bool {
	return env.Spec.Kind == v1.EnvironmentKindTypePermanent
}

// GetPermanentEnvironments returns a list with the current permanent environments
func GetPermanentEnvironments(jxClient versioned.Interface, ns string) ([]*v1.Environment, error) {
	var result []*v1.Environment
	envs, err := jxClient.JenkinsV1().Environments(ns).List(metav1.ListOptions{})
	if err != nil {
		return result, errors.Wrapf(err, "listing the environments in namespace %q", ns)
	}
	for i := range envs.Items {
		env := &envs.Items[i]
		if IsPermanentEnvironment(env) {
			result = append(result, env)
		}
	}
	return result, nil
}

// ModifyDevEnvironment modifies the dev environment
func ModifyDevEnvironment(kubeClient kubernetes.Interface, jxClient versioned.Interface, ns string, callback func(env *v1.Environment) error) error {
	err := EnsureDevNamespaceCreatedWithoutEnvironment(kubeClient, ns)
	if err != nil {
		return errors.Wrapf(err, "failed to create the %s Dev namespace", ns)
	}

	env, err := EnsureDevEnvironmentSetup(jxClient, ns)
	if err != nil {
		return errors.Wrapf(err, "failed to setup the dev environment for namespace '%s'", ns)
	}

	if env == nil {
		return fmt.Errorf("No Development environment found for namespace %s", ns)
	}

	err = callback(env)
	if err != nil {
		return errors.Wrap(err, "failed to call the callback function for dev environment")
	}
	_, err = jxClient.JenkinsV1().Environments(ns).PatchUpdate(env)
	if err != nil {
		return fmt.Errorf("Failed to update Development environment in namespace %s: %s", ns, err)
	}
	return nil
}
