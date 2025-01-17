package cobras

import (
	"github.com/spf13/cobra"
)

// SplitCommand helper command to ignore the options object
func SplitCommand(cmd *cobra.Command, options interface{}) *cobra.Command {
	return cmd
}
