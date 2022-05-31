package batch

import (
	"fmt"

	"github.com/jenkins-x/jx-helpers/v3/pkg/input"
)

type options struct{}

// NewBatchInput creates a new batch input implementation
func NewBatchInput() *options {
	return &options{}
}

var _ input.Interface = &options{}

// PickPassword gets a password (via hidden input) from a user's free-form input
func (f *options) PickPassword(message string, help string) (string, error) {
	return "", nil
}

// PickValue picks a value
func (f *options) PickValue(message string, defaultValue string, required bool, help string) (string, error) {
	return defaultValue, nil
}

// PickValidValue gets an answer to a prompt from a user's free-form input with a given validator
func (f *options) PickValidValue(message string, defaultValue string, validator func(val interface{}) error, help string) (string, error) {
	return f.PickValue(message, defaultValue, false, help)
}

// PickNameWithDefault picks a value
func (f *options) PickNameWithDefault(names []string, message string, defaultValue interface{}, help string) (string, error) {
	if fmt.Sprintf("%v", defaultValue) == "" && len(names) > 0 {
		defaultValue = names[0]
	}
	return fmt.Sprintf("%v", defaultValue), nil
}

func (f *options) SelectNamesWithFilter(names []string, message string, selectAll bool, filter string, help string) ([]string, error) {
	if len(names) == 0 {
		return nil, nil
	}
	return []string{names[0]}, nil
}

func (f *options) SelectNames(names []string, message string, selectAll bool, help string) ([]string, error) {
	if len(names) == 0 {
		return nil, nil
	}
	return []string{names[0]}, nil
}

func (f *options) Confirm(message string, defaultValue bool, help string) (bool, error) {
	return true, nil
}
