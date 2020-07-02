package fake

import (
	"strings"

	"github.com/jenkins-x/jx-helpers/pkg/input"
	"github.com/pkg/errors"
)

// FakeInput provide a fake provider for testing
type FakeInput struct {
	// Values the values to return indexed by the message
	Values map[string]string
}

var _ input.Interface = &FakeInput{}

// PickPassword gets a password (via hidden input) from a user's free-form input
func (f *FakeInput) PickPassword(message string, help string) (string, error) {
	if f.Values == nil {
		f.Values = map[string]string{}
	}
	value := f.Values[message]
	if value == "" {
		return "", errors.Errorf("missing fake value for message: %s", message)
	}
	return value, nil
}

// PickValue picks a value
func (f *FakeInput) PickValue(message string, defaultValue string, required bool, help string) (string, error) {
	if f.Values == nil {
		f.Values = map[string]string{}
	}
	value := f.Values[message]
	if value == "" {
		value = defaultValue
	}
	return value, nil
}

// PickNameWithDefault picks a value
func (f *FakeInput) PickNameWithDefault(names []string, message string, defaultValue string, help string) (string, error) {
	if f.Values == nil {
		f.Values = map[string]string{}
	}
	value := f.Values[message]
	if value == "" {
		value = defaultValue
	}
	return value, nil
}

func (f *FakeInput) SelectNamesWithFilter(names []string, message string, selectAll bool, filter string, help string) ([]string, error) {
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
