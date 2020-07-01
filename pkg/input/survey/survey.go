package survey

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/jenkins-x/jx-helpers/pkg/input"
	"github.com/jenkins-x/jx-logging/pkg/log"
	"gopkg.in/AlecAivazis/survey.v1"
	"gopkg.in/AlecAivazis/survey.v1/terminal"
)

type client struct {
	in  terminal.FileReader
	out terminal.FileWriter
	err io.Writer
}

// NewInput creates a new input using std in/out/err
func NewInput() input.Interface {
	return NewInputFrom(os.Stdin, os.Stdout, os.Stderr)
}

// NewInputFrom creates a new input from the given in/out/err
func NewInputFrom(in terminal.FileReader, out terminal.FileWriter, err io.Writer) input.Interface {
	return &client{
		in:  in,
		out: out,
		err: err,
	}
}

// PickPassword gets a password (via hidden input) from a user's free-form input
func (c *client) PickPassword(message string, help string) (string, error) {
	answer := ""
	prompt := &survey.Password{
		Message: message,
		Help:    help,
	}
	validator := survey.Required
	surveyOpts := survey.WithStdio(c.in, c.out, c.err)
	err := survey.AskOne(prompt, &answer, validator, surveyOpts)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(answer), nil
}

// PickValue gets an answer to a prompt from a user's free-form input
func (c *client) PickValue(message string, defaultValue string, required bool, help string) (string, error) {
	answer := ""
	prompt := &survey.Input{
		Message: message,
		Default: defaultValue,
		Help:    help,
	}
	validator := survey.Required
	if !required {
		validator = nil
	}
	surveyOpts := survey.WithStdio(c.in, c.out, c.err)
	err := survey.AskOne(prompt, &answer, validator, surveyOpts)
	if err != nil {
		return "", err
	}
	return answer, nil
}

// PickNameWithDefault gets the user to pick an option from a list of options, with a default option specified
func (c *client) PickNameWithDefault(names []string, message string, defaultValue string, help string) (string, error) {
	name := ""
	if len(names) == 0 {
		return "", nil
	} else if len(names) == 1 {
		name = names[0]
	} else {
		prompt := &survey.Select{
			Message: message,
			Options: names,
			Default: defaultValue,
		}
		surveyOpts := survey.WithStdio(c.in, c.out, c.err)
		err := survey.AskOne(prompt, &name, nil, surveyOpts)
		if err != nil {
			return "", err
		}
	}
	return name, nil
}

// SelectNamesWithFilter selects from a list of names with a given filter. Optionally selecting them all
func (c *client) SelectNamesWithFilter(names []string, message string, selectAll bool, filter string, help string) ([]string, error) {
	filtered := []string{}
	for _, name := range names {
		if filter == "" || strings.Index(name, filter) >= 0 {
			filtered = append(filtered, name)
		}
	}
	if len(filtered) == 0 {
		return nil, fmt.Errorf("No names match filter: %s", filter)
	}
	return c.SelectNames(filtered, message, selectAll, help)
}

// SelectNames select which names from the list should be chosen
func (c *client) SelectNames(names []string, message string, selectAll bool, help string) ([]string, error) {
	answer := []string{}
	if len(names) == 0 {
		return answer, fmt.Errorf("No names to choose from")
	}
	sort.Strings(names)

	prompt := &survey.MultiSelect{
		Message: message,
		Options: names,
		Help:    help,
	}
	if selectAll {
		prompt.Default = names
	}
	surveyOpts := survey.WithStdio(c.in, c.out, c.err)
	err := survey.AskOne(prompt, &answer, nil, surveyOpts)
	return answer, err
}

// Confirm prompts the user to confirm something
func (c *client) Confirm(message string, defaultValue bool, help string) (bool, error) {
	answer := defaultValue
	prompt := &survey.Confirm{
		Message: message,
		Default: defaultValue,
		Help:    help,
	}
	surveyOpts := survey.WithStdio(c.in, c.out, c.err)
	err := survey.AskOne(prompt, &answer, nil, surveyOpts)
	if err != nil {
		return false, err
	}
	log.Logger().Info("")
	return answer, nil
}
