package cmdrunner

import (
	"strings"

	"github.com/jenkins-x/jx-helpers/v3/pkg/termcolor"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
)

// CommandRunner represents a command runner so that it can be stubbed out for testing
type CommandRunner func(*Command) (string, error)

// DefaultCommandRunner default runner if none is set
func DefaultCommandRunner(c *Command) (string, error) {
	if c.Dir == "" {
		log.Logger().Infof("about to run: %s", termcolor.ColorInfo(CLI(c)))
	} else {
		log.Logger().Infof("about to run: %s in dir %s", termcolor.ColorInfo(CLI(c)), termcolor.ColorInfo(c.Dir))
	}
	result, err := c.RunWithoutRetry()
	if result != "" {
		log.Logger().Info(termcolor.ColorStatus(result))
	}
	return result, err
}

// QuietCommandRunner uses debug level logging to output commands executed and results
func QuietCommandRunner(c *Command) (string, error) {
	if c.Dir == "" {
		log.Logger().Debugf("about to run: %s", termcolor.ColorInfo(CLI(c)))
	} else {
		log.Logger().Debugf("about to run: %s in dir %s", termcolor.ColorInfo(CLI(c)), termcolor.ColorInfo(c.Dir))
	}
	result, err := c.RunWithoutRetry()
	if result != "" {
		log.Logger().Debug(termcolor.ColorStatus(result))
	}
	return result, err
}

// DryRunCommandRunner output the commands to be run
func DryRunCommandRunner(c *Command) (string, error) {
	log.Logger().Info(CLI(c))
	return "", nil
}

// CLI returns the CLI string without the dir or env vars
func CLI(c *Command) string {
	var builder strings.Builder
	builder.WriteString(c.Name)
	for _, arg := range c.Args {
		builder.WriteString(" ")
		builder.WriteString(arg)
	}
	return builder.String()
}
