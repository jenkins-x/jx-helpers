package podlogs

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/jenkins-x/jx-logging/v3/pkg/log"
)

// TailLogs will tail the logs for the pod in ns with containerName,
// returning when the logs are complete. It writes to errOut and out.
func TailLogs(ns string, pod string, containerName string, errOut io.Writer, out io.Writer) error {
	args := []string{"logs", "-n", ns, "-f"}
	if containerName != "" {
		args = append(args, "-c", containerName)
	}
	args = append(args, pod)
	name := "kubectl"

	log.Logger().Debugf("about to run: kubectl %s\n", strings.Join(args, " "))

	e := exec.Command(name, args...)
	e.Stderr = errOut
	stdout, _ := e.StdoutPipe()

	err := e.Start()
	if err != nil {
		return fmt.Errorf("failed to run command: %s %s: %w", name, strings.Join(args, " "), err)
	}

	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		m := scanner.Text()
		fmt.Fprintln(out, m)
		if m == "Finished: FAILURE" {
			os.Exit(1)
		}
	}
	e.Wait()
	return nil
}
