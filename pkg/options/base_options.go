package options

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jenkins-x/jx-helpers/v3/pkg/signals"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"

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
	BatchMode bool   `env:"JX_BATCH_MODE"`
	Verbose   bool   `env:"JX_VERBOSE"`
	LogLevel  string `env:"JX_LOG_LEVEL"`
	Ctx       context.Context
	Command   *cobra.Command
	Out       io.Writer
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
			return fmt.Errorf("failed to set the log level %s: %w", o.LogLevel, err)
		}
		return nil
	}
	if o.Verbose {
		err := log.SetLevel("debug")
		if err != nil {
			return fmt.Errorf("failed to set debug level: %w", err)
		}
	}
	if o.Out == nil {
		o.Out = os.Stdout
	}
	log.SetOutput(o.Out)
	return nil
}

// GetContext lazily creates the context if its not already set
func (o *BaseOptions) GetContext() context.Context {
	if o.Ctx == nil {
		// handles ctrl-c
		o.Ctx = signals.NewContext()
	}
	return o.Ctx
}
