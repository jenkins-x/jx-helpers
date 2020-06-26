package cli

import (
	"github.com/jenkins-x/jx-helpers/pkg/cmdrunner"
	"github.com/jenkins-x/jx-helpers/pkg/gitclient"
)

type client struct {
	binary string
	runner cmdrunner.CommandRunner
}

// NewCLIClient creates a new CLI based client
// if no binary or runner is supplied then use the defaults
func NewCLIClient(binary string, runner cmdrunner.CommandRunner) gitclient.Interface {
	if binary == "" {
		binary = "git"
	}
	if runner == nil {
		runner = cmdrunner.DefaultCommandRunner
	}
	return &client{
		binary: binary,
		runner: runner,
	}
}

// Command invokes the git command
func (c *client) Command(dir string, args ...string) (string, error) {
	cmd := &cmdrunner.Command{
		Dir:  dir,
		Name: c.binary,
		Args: args,
	}
	return c.runner(cmd)
}
