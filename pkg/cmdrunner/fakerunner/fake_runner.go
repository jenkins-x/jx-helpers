package fakerunner

import (
	"sort"
	"testing"

	"github.com/jenkins-x/jx-helpers/v3/pkg/cmdrunner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// FakeRunner for testing command runners
type FakeRunner struct {
	// Commands sorted commands
	Commands []*cmdrunner.Command

	// OrderedCommands recorded commands in order
	OrderedCommands []*cmdrunner.Command

	// CommandRunner if specified this callback returns the results and error
	CommandRunner cmdrunner.CommandRunner

	// ResultOutput default output if no CommandRunner
	ResultOutput string

	// ResultError default error output if no CommandRunner
	ResultError error
}

// FakeResult the expected results
type FakeResult struct {
	CLI string
	Dir string
	Env map[string]string
}

// Run the default implementation
func (f *FakeRunner) Run(c *cmdrunner.Command) (string, error) {
	f.Commands = append(f.Commands, c)
	f.OrderedCommands = append(f.OrderedCommands, c)

	if f.CommandRunner != nil {
		return f.CommandRunner(c)
	}
	return f.ResultOutput, f.ResultError
}

// Expects expects the given results
func (f *FakeRunner) ExpectResults(t *testing.T, results ...FakeResult) {
	commands := f.Commands
	for _, c := range commands {
		t.Logf("got command %s\n", cmdrunner.CLI(c))
	}

	require.Equal(t, len(results), len(commands), "expected command invocations")

	sort.Slice(commands, func(i, j int) bool {
		c1 := cmdrunner.CLI(commands[i])
		c2 := cmdrunner.CLI(commands[j])
		return c1 < c2
	})

	sort.Slice(results, func(i, j int) bool {
		c1 := results[i].CLI
		c2 := results[j].CLI
		return c1 < c2
	})

	for i, r := range results {
		c := commands[i]
		assert.Equal(t, r.CLI, cmdrunner.CLI(c), "command line for command %d", i+1)
		if r.Dir != "" {
			assert.Equal(t, r.Dir, c.Dir, "directory line for command %d", i+1)
		}
		if r.Env != nil {
			for k, v := range r.Env {
				actual := ""
				if c.Env != nil {
					actual = c.Env[k]
				}
				assert.Equal(t, v, actual, "$%s for command %d", k, i+1)
			}
		}
	}
}
