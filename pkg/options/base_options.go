package options

import (
	"fmt"
	"os"
	"strings"

	"github.com/jenkins-x/jx-logging/pkg/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const (
	// OptionBatchMode command line option to enable batch mode
	OptionBatchMode = "batch-mode"

	// OptionVerbose command line option to enable verbose logging
	OptionVerbose = "verbose"
)

// BaseOptions a few common options we tend to use in command line tools
type BaseOptions struct {
	BatchMode bool
	Verbose   bool
	LogLevel  string
	Command   *cobra.Command
}

// AddBaseFlags adds the base flags for all commands
func (o *BaseOptions) AddBaseFlags(cmd *cobra.Command) {
	defaultBatchMode := false
	if os.Getenv("JX_BATCH_MODE") == "true" {
		defaultBatchMode = true
	}
	cmd.Flags().StringVarP(&o.LogLevel, "log-level", "", "", "Sets the logging level. If not specified defaults to $JX_LOG_LEVEL")
	cmd.Flags().BoolVarP(&o.BatchMode, OptionBatchMode, "b", defaultBatchMode, "Runs in batch mode without prompting for user input")
	levels := strings.Join([]string{"panic", "fatal", "error", "warn", "info", "debug", "trace"}, ", ")
	cmd.Flags().BoolVarP(&o.Verbose, OptionVerbose, "", false, fmt.Sprintf("Enables verbose output. The environment variable JX_LOG_LEVEL has precedence over this flag and allows setting the logging level to any value of: %s", levels))
	o.Command = cmd
}

// Validate verifies settings
func (o *BaseOptions) Validate() error {
	if o.LogLevel == "" {
		o.LogLevel = os.Getenv("JX_LOG_LEVEL")
	}
	if o.LogLevel != "" {
		err := log.SetLevel(o.LogLevel)
		if err != nil {
			return errors.Wrapf(err, "failed to set the log level %s", o.LogLevel)
		}
		return nil
	}
	if o.Verbose {
		err := log.SetLevel("debug")
		if err != nil {
			return errors.Wrapf(err, "failed to set debug level")
		}

	}
	return nil
}
