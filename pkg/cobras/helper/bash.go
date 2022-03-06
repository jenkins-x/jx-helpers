package helper

import "fmt"

// BashExample returns markdown for a bash script expression for the CLI help
func BashExample(binaryCommand, cli string) string {
	return fmt.Sprintf("\n```bash \n%s %s\n```\n", binaryCommand, cli)
}
