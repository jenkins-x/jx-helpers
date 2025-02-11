package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	jxCore "github.com/jenkins-x/jx-api/v4/pkg/apis/jenkins.io/v1"
	"github.com/jenkins-x/jx-helpers/v3/pkg/extensions"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/sets"
)

// GetPluginCommandGroups returns the plugin groups
func GetPluginCommandGroups(verifier extensions.PathVerifier, localPlugins []jxCore.Plugin) (PluginCommandGroups, error) {

	otherCommands := PluginCommandGroup{
		Message: "Other Commands",
	}
	groups := make(map[string]PluginCommandGroup, 0)

	addPlugins(localPlugins, &otherCommands, groups)

	pathCommands := PluginCommandGroup{
		Message: "Locally Available Commands:",
	}

	path := "PATH"
	if runtime.GOOS == "windows" {
		path = "path"
	}

	paths := sets.NewString(filepath.SplitList(os.Getenv(path))...)
	for _, dir := range paths.List() {
		files, err := os.ReadDir(dir)
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
				PluginSpec: jxCore.PluginSpec{
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

func addPlugins(pluginSlice []jxCore.Plugin, otherCommands *PluginCommandGroup, groups map[string]PluginCommandGroup) {
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
func ActsAsRootCommand(cmd *cobra.Command, filters []string, getPluginCommandGroups func() PluginCommandGroups, groups ...CommandGroup) FlagExposer {
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
