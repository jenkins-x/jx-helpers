package linter

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jenkins-x/jx-api/v4/pkg/util"
	"github.com/jenkins-x/jx-helpers/v3/pkg/files"
	"github.com/jenkins-x/jx-helpers/v3/pkg/options"
	"github.com/jenkins-x/jx-helpers/v3/pkg/table"
	"github.com/jenkins-x/jx-helpers/v3/pkg/termcolor"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"

	"github.com/spf13/cobra"
	yaml2 "gopkg.in/yaml.v2"
	"sigs.k8s.io/yaml"
)

const (
	// FormatTap for the tap based output format
	FormatTap = "tap"
)

// Options contains the command line options
type Options struct {
	options.BaseOptions

	OutFile string
	Format  string
	Tests   []*Test
}

type Test struct {
	File    string
	Error   error
	Message string
}

// Linter represents a lint check if a file exists matching the path
type Linter struct {
	Path   string
	Linter func(path string, test *Test) error
}

var (
	info = termcolor.ColorInfo
)

// AddFlags adds command line flags
func (o *Options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.OutFile, "out", "o", "", "The TAP format file to output with the results. If not specified the tap file is output to the terminal")
	cmd.Flags().StringVarP(&o.Format, "format", "", "", "If specify 'tap' lets use the TAP output otherwise use simple text output")
}

// Lint lints the given linters and renders the results
func (o *Options) Lint(ls []Linter, dir string) error {
	for _, l := range ls {
		path := filepath.Join(dir, l.Path)
		exists, err := files.FileExists(path)
		if err != nil {
			return fmt.Errorf("failed to check if file exists %s: %w", path, err)
		}
		if !exists {
			continue
		}

		test := &Test{
			File: l.Path,
		}
		o.Tests = append(o.Tests, test)

		err = l.Linter(path, test)
		if err != nil {
			return fmt.Errorf("failed to lint %s: %w", path, err)
		}
	}
	return o.LogResults()
}

// LogResults logs the results
func (o *Options) LogResults() error {
	if o.Format == FormatTap || o.OutFile != "" {
		return o.logTapResults()
	}

	t := table.CreateTable(os.Stdout)
	t.AddRow("FILE", "STATUS")

	for _, test := range o.Tests {
		name := test.File
		err := test.Error
		status := info("OK")
		if err != nil {
			status = termcolor.ColorWarning(err.Error())
		}
		t.AddRow(name, status)
	}
	t.Render()
	return nil
}

func (o *Options) logTapResults() error {
	buf := strings.Builder{}
	buf.WriteString("TAP version 13\n")
	count := len(o.Tests)
	buf.WriteString(fmt.Sprintf("1..%d\n", count))
	var failed []string
	for i, test := range o.Tests {
		n := i + 1
		if test.Error != nil {
			failed = append(failed, strconv.Itoa(n))
			buf.WriteString(fmt.Sprintf("not ok %d - %s\n", n, test.File))
		} else {
			buf.WriteString(fmt.Sprintf("ok %d - %s\n", n, test.File))
		}
	}
	failedCount := len(failed)
	if failedCount > 0 {
		buf.WriteString(fmt.Sprintf("FAILED tests %s\n", strings.Join(failed, ", ")))
	}
	var p float32
	if count > 0 {
		p = float32(100 * (count - failedCount) / count)
	}
	buf.WriteString(fmt.Sprintf("Failed %d/%d tests, %.2f", failedCount, count, p))
	buf.WriteString("%% okay\n")

	text := buf.String()
	if o.OutFile != "" {
		err := os.WriteFile(o.OutFile, []byte(text), files.DefaultFileWritePermissions)
		if err != nil {
			return fmt.Errorf("failed to save file %s: %w", o.OutFile, err)
		}
		log.Logger().Infof("saved file %s", info(o.OutFile))
		return nil
	}
	log.Logger().Infof(text)
	return nil
}

// LintResource lints a resource
func (o *Options) LintResource(path string, test *Test, resource interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", path, err)
	}

	err = yaml.UnmarshalStrict(data, resource)
	if err != nil {
		test.Error = err
		return nil
	}

	validationErrors, err := util.ValidateYaml(resource, data)
	if err != nil {
		return fmt.Errorf("failed to validate %s: %w", path, err)
	}

	for _, msg := range validationErrors {
		if test.Message == "" {
			test.Message = msg
		} else {
			test.Message = test.Message + "\n" + msg
		}
	}
	return nil
}

// LintYaml2Resource lints a resource
func (o *Options) LintYaml2Resource(path string, test *Test, resource interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", path, err)
	}

	err = yaml2.UnmarshalStrict(data, resource)
	if err != nil {
		test.Error = err
		return nil
	}

	validationErrors, err := util.ValidateYaml(resource, data)
	if err != nil {
		return fmt.Errorf("failed to validate %s: %w", path, err)
	}

	for _, msg := range validationErrors {
		if test.Message == "" {
			test.Message = msg
		} else {
			test.Message = test.Message + "\n" + msg
		}
	}
	return nil
}
