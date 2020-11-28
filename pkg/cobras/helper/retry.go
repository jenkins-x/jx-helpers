package helper

import (
	"regexp"

	flag "github.com/spf13/pflag"

	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/spf13/cobra"
)

const (
	DefaultCommandAttempts = 1
)

type retryOptions struct {
	retryAttempts int
}

func (o *retryOptions) run(command *cobra.Command, fn func(error) bool) {
	var err error
	for i := 1; i <= o.retryAttempts; i++ {
		err = command.Execute()
		if err == nil {
			break
		}
		retry := true
		if fn != nil {
			retry = fn(err)
		}
		if !retry {
			break
		}
		log.Logger().Warnf("error running %s, retry %d of %d", command.Use, i, o.retryAttempts)
	}
	if err != nil {
		CheckErr(err)
	}
}

//RetryOnErrorCommand accepts an existing cobra Command with an optional function to determine which errors constitute
//ones which will cause the command to be reattempted
func RetryOnErrorCommand(command *cobra.Command, fn func(error) bool) *cobra.Command {

	o := retryOptions{}
	cmd := &cobra.Command{
		Use:     command.Use,
		Short:   command.Short,
		Long:    command.Long,
		Example: command.Example,
		Run: func(cmd *cobra.Command, args []string) {
			o.run(command, fn)
		},
	}

	command.Flags().VisitAll(func(flag *flag.Flag) {
		cmd.Flags().AddFlag(flag)
	})

	cmd.Flags().IntVarP(&o.retryAttempts, "retries", "", DefaultCommandAttempts, "Specify the number of times the command should be reattempted on failure")

	return cmd
}

//RegexRetryFunction uses a list of supplied regexes and returns a function that will accept an error and return true if
//if the error matches against any supplied regexes
func RegexRetryFunction(retryErrorRegexes []string) func(error) bool {
	return func(e error) bool {
		for _, retriableErr := range retryErrorRegexes {
			re, err := regexp.Compile(retriableErr)
			if err != nil {
				return false
			}
			if re.MatchString(e.Error()) {
				return true
			}
		}
		return false
	}
}
