/*
Copyright 2016 The Kubernetes Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package templates

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/template"
	"unicode"

	"github.com/jenkins-x/jx-helpers/v3/pkg/table"
	"github.com/jenkins-x/jx-helpers/v3/pkg/termcolor"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

type FlagExposer interface {
	ExposeFlags(cmd *cobra.Command, flags ...string) FlagExposer
}

func UseOptionsTemplates(cmd *cobra.Command) {
	templater := &templater{
		UsageTemplate: OptionsUsageTemplate(),
		HelpTemplate:  OptionsHelpTemplate(),
	}
	cmd.SetUsageFunc(templater.UsageFunc())
	cmd.SetHelpFunc(templater.HelpFunc())
}

type templater struct {
	UsageTemplate          string
	HelpTemplate           string
	GetPluginCommandGroups func() (PluginCommandGroups, bool)
	RootCmd                *cobra.Command
	CommandGroups
	Filtered []string
}

func (templater *templater) ExposeFlags(cmd *cobra.Command, flags ...string) FlagExposer {
	cmd.SetUsageFunc(templater.UsageFunc(flags...))
	return templater
}

func (templater *templater) HelpFunc() func(*cobra.Command, []string) {
	return func(c *cobra.Command, s []string) {
		t := template.New("help")
		t.Funcs(templater.templateFuncs())
		template.Must(t.Parse(templater.HelpTemplate))
		out := newResponsiveWriter(c.OutOrStdout())
		err := t.Execute(out, c)
		if err != nil {
			log.Logger().Errorf("failed to generate help template: %s", err.Error())
		}
	}
}

func (templater *templater) UsageFunc(exposedFlags ...string) func(*cobra.Command) error {
	return func(c *cobra.Command) error {
		t := template.New("usage")
		t.Funcs(templater.templateFuncs(exposedFlags...))
		template.Must(t.Parse(templater.UsageTemplate))
		out := newResponsiveWriter(c.OutOrStderr())
		return t.Execute(out, c)
	}
}

func (templater *templater) templateFuncs(exposedFlags ...string) template.FuncMap {
	return template.FuncMap{
		"trim":                strings.TrimSpace,
		"trimRight":           func(s string) string { return strings.TrimRightFunc(s, unicode.IsSpace) },
		"trimLeft":            func(s string) string { return strings.TrimLeftFunc(s, unicode.IsSpace) },
		"gt":                  cobra.Gt,
		"eq":                  cobra.Eq,
		"rpad":                rpad,
		"appendIfNotPresent":  appendIfNotPresent,
		"flagsNotIntersected": flagsNotIntersected,
		"visibleFlags":        visibleFlags,
		"flagsUsages":         flagsUsages,
		"cmdGroups":           templater.cmdGroups,
		"cmdGroupsString":     templater.cmdGroupsString,
		"rootCmd":             templater.rootCmdName,
		"isRootCmd":           templater.isRootCmd,
		"optionsCmdFor":       templater.optionsCmdFor,
		"usageLine":           templater.usageLine,
		"exposed": func(c *cobra.Command) *flag.FlagSet {
			exposed := flag.NewFlagSet("exposed", flag.ContinueOnError)
			if len(exposedFlags) > 0 {
				for _, name := range exposedFlags {
					if f := c.Flags().Lookup(name); f != nil {
						exposed.AddFlag(f)
					}
				}
			}
			return exposed
		},
	}
}

func newResponsiveWriter(out io.Writer) io.Writer {
	// TODO term.NewResponsiveWriter(out)
	return out
}

func (templater *templater) cmdGroups(c *cobra.Command, all []*cobra.Command) []CommandGroup {
	if len(templater.CommandGroups) > 0 && c == templater.RootCmd {
		all = filter(all, templater.Filtered...)
		return AddAdditionalCommands(templater.CommandGroups, "Other Commands:", all)
	}
	all = filter(all, "options")
	return []CommandGroup{
		{
			Message:  "Available Commands:",
			Commands: all,
		},
	}
}

func (t *templater) groupPath(c *cobra.Command) string {
	parentText := ""
	parent := c.Parent()
	if parent != nil {
		parentText = t.groupPath(parent)
		if parentText != "" {
			parentText += " "
		}
	}
	return strings.TrimPrefix(parentText, "jx ") + c.Name()
}

func (t *templater) cmdGroupsString(c *cobra.Command) string {
	var groups []string
	maxLen := 1
	for _, cmdGroup := range t.cmdGroups(c, c.Commands()) {
		for _, cmd := range cmdGroup.Commands {
			if cmd.Runnable() {
				path := t.groupPath(cmd)
				l := len(path)
				if l > maxLen {
					maxLen = l
				}
			}
		}
	}
	pluginCommandGroups, _ := t.GetPluginCommandGroups()
	for _, cmdGroup := range pluginCommandGroups {
		for _, cmd := range cmdGroup.Commands {
			l := len(cmd.SubCommand)
			if l > maxLen {
				maxLen = l
			}
		}
	}
	for _, cmdGroup := range t.cmdGroups(c, c.Commands()) {
		cmds := []string{cmdGroup.Message}
		for _, cmd := range cmdGroup.Commands {
			if cmd.Runnable() {
				path := t.groupPath(cmd)
				//cmds = append(cmds, "  "+rpad(path, maxLen - len(path))+" "+cmd.Short)
				cmds = append(cmds, "  "+table.PadRight(termcolor.ColorInfo(path), " ", maxLen)+" "+cmd.Short)
			}
		}
		groups = append(groups, strings.Join(cmds, "\n"))
	}
	for _, cmdGroup := range pluginCommandGroups {
		cmds := []string{cmdGroup.Message}
		for _, cmd := range cmdGroup.Commands {
			//cmds = append(cmds, "  "+rpad(path, maxLen - len(path))+" "+cmd.Short)
			cmds = append(cmds, "  "+table.PadRight(termcolor.ColorInfo(cmd.SubCommand), " ", maxLen)+" "+fmt.Sprintf("%s (from plugin)", cmd.Description))
		}
		groups = append(groups, strings.Join(cmds, "\n"))
	}
	return fmt.Sprintf("%s\n", strings.Join(groups, "\n\n"))
}

func (t *templater) rootCmdName(c *cobra.Command) string {
	return t.rootCmd(c).CommandPath()
}

func (t *templater) isRootCmd(c *cobra.Command) bool {
	return t.rootCmd(c) == c
}

func (t *templater) parents(c *cobra.Command) []*cobra.Command {
	parents := []*cobra.Command{c}
	for current := c; !t.isRootCmd(current) && current.HasParent(); {
		current = current.Parent()
		parents = append(parents, current)
	}
	return parents
}

func (t *templater) rootCmd(c *cobra.Command) *cobra.Command {
	if c != nil && !c.HasParent() {
		return c
	}
	if t.RootCmd == nil {
		panic("nil root cmd")
	}
	return t.RootCmd
}

func (t *templater) optionsCmdFor(c *cobra.Command) string {
	if !c.Runnable() {
		return ""
	}
	rootCmdStructure := t.parents(c)
	for i := len(rootCmdStructure) - 1; i >= 0; i-- {
		cmd := rootCmdStructure[i]
		if _, _, err := cmd.Find([]string{"options"}); err == nil {
			return cmd.CommandPath() + " options"
		}
	}
	return ""
}

func (t *templater) usageLine(c *cobra.Command) string {
	usage := c.UseLine()
	suffix := "[options]"
	if c.HasFlags() && !strings.Contains(usage, suffix) {
		usage += " " + suffix
	}
	return usage
}

func flagsUsages(f *flag.FlagSet) string {
	x := new(bytes.Buffer)

	f.VisitAll(func(flag *flag.Flag) {
		if flag.Hidden {
			return
		}
		format := "--%s=%s: %s\n"

		if flag.Value.Type() == "string" {
			format = "--%s='%s': %s\n"
		}

		if len(flag.Shorthand) > 0 {
			format = "  -%s, " + format
		} else {
			format = "   %s   " + format
		}

		fmt.Fprintf(x, format, flag.Shorthand, flag.Name, flag.DefValue, flag.Usage)
	})

	return x.String()
}

func rpad(s string, padding int) string {
	t := fmt.Sprintf("%%-%ds", padding)
	return fmt.Sprintf(t, s)
}

func appendIfNotPresent(s, stringToAppend string) string {
	if strings.Contains(s, stringToAppend) {
		return s
	}
	return s + " " + stringToAppend
}

func flagsNotIntersected(l *flag.FlagSet, r *flag.FlagSet) *flag.FlagSet {
	f := flag.NewFlagSet("notIntersected", flag.ContinueOnError)
	f.AddFlagSet(l)
	r.VisitAll(func(flag *flag.Flag) {
		if l.Lookup(flag.Name) == nil {
			f.AddFlag(flag)
		}
	})
	return f
}

func visibleFlags(l *flag.FlagSet) *flag.FlagSet {
	hidden := "help"
	f := flag.NewFlagSet("visible", flag.ContinueOnError)
	l.VisitAll(func(flag *flag.Flag) {
		if flag.Name != hidden {
			f.AddFlag(flag)
		}
	})
	return f
}

func filter(cmds []*cobra.Command, names ...string) []*cobra.Command {
	var out []*cobra.Command
	for _, c := range cmds {
		if c.Hidden {
			continue
		}
		skip := false
		for _, name := range names {
			if name == c.Name() {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		out = append(out, c)
	}
	return out
}
