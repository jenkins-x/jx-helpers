package helper_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/jenkins-x/jx-helpers/v3/pkg/cobras/helper"
	"github.com/jenkins-x/jx-helpers/v3/pkg/files"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func makeCommand(retryTestFunction func(error) bool) *cobra.Command {
	type options struct {
		outputfile string
		testfile   string
	}
	o := options{}
	command := &cobra.Command{
		RunE: func(cmd *cobra.Command, args []string) error {

			f, err := os.OpenFile(o.outputfile,
				os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return fmt.Errorf("error opening file %s", o.outputfile)
			}
			defer f.Close()
			if _, err := f.WriteString("test line\n"); err != nil {
				return fmt.Errorf("error outputting test line")
			}

			exists, err := files.FileExists(o.testfile)
			if !exists {
				return fmt.Errorf("error finding file")
			}
			if err != nil {
				return err
			}

			return nil
		},
	}
	command.Flags().StringVarP(&o.outputfile, "outputfile", "", "", "test file to write to")
	command.Flags().StringVarP(&o.testfile, "testfile", "", "", "test file to check exists")

	return helper.RetryOnErrorCommand(command, retryTestFunction)
}

func TestRetryCommand(t *testing.T) {
	helper.BehaviorOnFatal(func(s string, i int) {})
	retries := 5

	tmpDir := t.TempDir()

	testCases := map[string]struct {
		filename          string
		createTestFile    bool
		outputFile        string
		lineCount         int
		retryTestFunction func(error) bool
	}{
		"exists": {
			filename:          filepath.Join(tmpDir, "exists"),
			createTestFile:    true,
			outputFile:        filepath.Join(tmpDir, "exists_output"),
			lineCount:         1,
			retryTestFunction: nil,
		},
		"elusive": {
			filename:          filepath.Join(tmpDir, "elusive"),
			createTestFile:    false,
			outputFile:        filepath.Join(tmpDir, "elusive_output"),
			lineCount:         5,
			retryTestFunction: nil,
		},
		"exists_but_no_retry": {
			filename:          filepath.Join(tmpDir, "exists_but_no_retry"),
			createTestFile:    true,
			outputFile:        filepath.Join(tmpDir, "exists_but_no_retry_output"),
			lineCount:         1,
			retryTestFunction: func(e error) bool { return false },
		},
		"does_not_exists_but_regex_causes_retry": {
			filename:          filepath.Join(tmpDir, "does_not_exists_but_regex_causes_retry"),
			createTestFile:    false,
			outputFile:        filepath.Join(tmpDir, "does_not_exists_but_regex_causes_retry_output"),
			lineCount:         5,
			retryTestFunction: helper.RegexRetryFunction([]string{"error f...ing file"}),
		},
	}

	for _, tc := range testCases {
		if tc.createTestFile {
			err := ioutil.WriteFile(tc.filename, []byte{105, 110}, os.ModePerm)
			if err != nil {
				t.Fatalf("failed to write test file %s", tc.filename)
			}
		}
	}

	for _, file := range testCases {
		command := makeCommand(file.retryTestFunction)
		command.SetArgs([]string{
			"--outputfile",
			file.outputFile,
			"--testfile",
			file.filename,
			"--retries",
			strconv.Itoa(retries),
		})

		command.Execute()

		f, err := os.Open(file.outputFile)
		if err != nil {
			t.Fatalf("failed to open outpu file %s", file.outputFile)
		}
		lineCount, err := lineCounter(f)
		if err != nil {
			t.Fatalf("failed to count output file lines %s", file.outputFile)
		}
		assert.Equal(t, file.lineCount, lineCount)
	}

}

func lineCounter(r io.Reader) (int, error) {
	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}
