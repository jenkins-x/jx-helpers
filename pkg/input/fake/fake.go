package fake

import (
	"strings"

	"github.com/jenkins-x/jx-helpers/v3/pkg/input"
	"github.com/pkg/errors"
)

// FakeInput provide a fake provider for testing
type FakeInput struct {
	// Values the values to return indexed by the message
	Values map[string]string

	// OrderedValues are used if there is not a value indexed by message
	OrderedValues []string

	// Counter the number of fields that have been requested so far
	Counter int
}

var _ input.Interface = &FakeInput{}

// PickPassword gets a password (via hidden input) from a user's free-form input
func (f *FakeInput) PickPassword(message, help string) (string, error) {
	value := f.getValue(message)
	if value == "" {
		return "", errors.Errorf("missing fake value for message: %s", message)
	}
	return value, nil
}

// PickValue picks a value
func (f *FakeInput) PickValue(message, defaultValue string, required bool, help string) (string, error) {
	value := f.getValue(message)
	if value == "" {
		value = defaultValue
	}
	return value, nil
}

// PickValidValue gets an answer to a prompt from a user's free-form input with a given validator
func (f *FakeInput) PickValidValue(message, defaultValue string, validator func(val interface{}) error, help string) (string, error) {
	return f.PickValue(message, defaultValue, false, help)
}

// PickNameWithDefault picks a value
func (f *FakeInput) PickNameWithDefault(names []string, message, defaultValue, help string) (string, error) {
	value := f.getValue(message)
	if value == "" {
		value = defaultValue
	}
	return value, nil
}

func (f *FakeInput) SelectNamesWithFilter(names []string, message string, selectAll bool, filter, help string) ([]string, error) {
	return f.SelectNames(names, message, false, help)
}

func (f *FakeInput) SelectNames(names []string, message string, selectAll bool, help string) ([]string, error) {
	value, err := f.PickValue(message, "", false, help)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to ")
	}
	return []string{value}, nil
}

func (f *FakeInput) Confirm(message string, defaultValue bool, help string) (bool, error) {
	value, err := f.PickValue(message, "", false, help)
	if err != nil {
		return false, errors.Wrapf(err, "failed to ")
	}
	if value == "" {
		return defaultValue, nil
	}
	switch strings.ToLower(value) {
	case "true", "y", "yes":
		return true, nil
	default:
		return false, nil
	}
}

// getValue returns the value using the message key or the ordered value if not available
func (f *FakeInput) getValue(message string) string {
	if f.Values == nil {
		f.Values = map[string]string{}
	}
	if f.OrderedValues == nil {
		f.OrderedValues = []string{}
	}
	value := f.Values[message]
	if value == "" {
		// lets see if we have an ordered value
		if f.Counter < len(f.OrderedValues) {
			value = f.OrderedValues[f.Counter]
		}
	}
	f.Counter++
	return value
}
