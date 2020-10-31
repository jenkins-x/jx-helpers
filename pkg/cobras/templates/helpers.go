package templates

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	jenkinsv1 "github.com/jenkins-x/jx-api/v3/pkg/apis/jenkins.io/v1"
	"github.com/jenkins-x/jx-api/v3/pkg/client/clientset/versioned"
	"github.com/jenkins-x/jx-helpers/v3/pkg/extensions"
	"github.com/jenkins-x/jx-helpers/v3/pkg/kube/jxclient"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

type Options struct {
	ManagedPluginsEnabled bool
	JXClient              versioned.Interface
	Namespace             string
}

// GetPluginCommandGroups returns the plugin groups
func (o *Options) GetPluginCommandGroups(verifier extensions.PathVerifier, localPlugins []jenkinsv1.Plugin) (PluginCommandGroups, error) {

	otherCommands := PluginCommandGroup{
		Message: "Other Commands",
	}
	groups := make(map[string]PluginCommandGroup, 0)

	o.addPlugins(localPlugins, otherCommands, groups)

	err := o.addManagedPlugins(otherCommands, groups)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to add managed plugins")
	}

	pathCommands := PluginCommandGroup{
		Message: "Locally Available Commands:",
	}

	path := "PATH"
	if runtime.GOOS == "windows" {
		path = "path"
	}

	paths := sets.NewString(filepath.SplitList(os.Getenv(path))...)
	for _, dir := range paths.List() {
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, f := range files {
			if f.IsDir() {
				continue
			}
			if !strings.HasPrefix(f.Name(), "jx-") {
				continue
			}

			pluginPath := filepath.Join(dir, f.Name())
			subCommand := strings.TrimPrefix(strings.Replace(filepath.Base(pluginPath), "-", " ", -1), "jx ")
			pc := &PluginCommand{
				PluginSpec: jenkinsv1.PluginSpec{
					SubCommand:  subCommand,
					Description: pluginPath,
				},
				Errors: make([]error, 0),
			}
			pathCommands.Commands = append(pathCommands.Commands, pc)
			if errs := verifier.Verify(filepath.Join(dir, f.Name())); len(errs) != 0 {
				for _, err := range errs {
					pc.Errors = append(pc.Errors, err)
				}
			}
		}
	}

	pcgs := PluginCommandGroups{}
	for _, g := range groups {
		pcgs = append(pcgs, g)
	}
	if len(otherCommands.Commands) > 0 {
		pcgs = append(pcgs, otherCommands)
	}
	if len(pathCommands.Commands) > 0 {
		pcgs = append(pcgs, pathCommands)
	}
	return pcgs, nil
}

func (o *Options) addManagedPlugins(otherCommands PluginCommandGroup, groups map[string]PluginCommandGroup) error {
	// Managed plugins
	var err error
	if o.ManagedPluginsEnabled {
		o.JXClient, o.Namespace, err = jxclient.LazyCreateJXClientAndNamespace(o.JXClient, o.Namespace)
		if err != nil {
			return errors.Wrapf(err, "failed to create jx client")
		}
		pluginList, err := o.JXClient.JenkinsV1().Plugins(o.Namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil && apierrors.IsNotFound(err) {
			err = nil
		}
		if err != nil {
			log.Logger().Debugf("failed to find Plugin CRDs in kubernetes namespace %s due to: %s", o.Namespace, err.Error())
		}
		o.addPlugins(pluginList.Items, otherCommands, groups)
	}
	return nil
}

func (o *Options) addPlugins(pluginSlice []jenkinsv1.Plugin, otherCommands PluginCommandGroup, groups map[string]PluginCommandGroup) {
	for _, plugin := range pluginSlice {
		pluginCommand := &PluginCommand{
			PluginSpec: plugin.Spec,
		}
		if plugin.Spec.Group == "" {
			otherCommands.Commands = append(otherCommands.Commands, pluginCommand)
		} else {
			if g, ok := groups[plugin.Spec.Group]; !ok {
				groups[plugin.Spec.Group] = PluginCommandGroup{
					Message: fmt.Sprintf("%s:", plugin.Spec.Group),
					Commands: []*PluginCommand{
						pluginCommand,
					},
				}
			} else {
				g.Commands = append(g.Commands, pluginCommand)
			}
		}
	}
}

// ActsAsRootCommand act as if the given set of plugins is a single root command
func ActsAsRootCommand(cmd *cobra.Command, filters []string, getPluginCommandGroups func() (PluginCommandGroups, bool), groups ...CommandGroup) FlagExposer {
	if cmd == nil {
		panic("nil root command")
	}
	templater := &templater{
		RootCmd:                cmd,
		UsageTemplate:          MainUsageTemplate(),
		HelpTemplate:           MainHelpTemplate(),
		GetPluginCommandGroups: getPluginCommandGroups,
		CommandGroups:          groups,
		Filtered:               filters,
	}
	cmd.SetUsageFunc(templater.UsageFunc())
	cmd.SetHelpFunc(templater.HelpFunc())
	return templater
}
