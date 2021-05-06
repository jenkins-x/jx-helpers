package options_test

import (
	"github.com/jenkins-x/jx-helpers/v3/pkg/options"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestArgumentsOptionValue(t *testing.T) {
	testCases := []struct {
		args     []string
		flag     string
		option   string
		expected string
	}{
		{
			args:     []string{"-xflag", "-o", "cheese", "something"},
			flag:     "o",
			option:   "output",
			expected: "cheese",
		},
		{
			args:     []string{"-xflag", "-o=cheese", "something"},
			flag:     "o",
			option:   "output",
			expected: "cheese",
		},
		{
			args:     []string{"-xflag", "--output", "cheese", "something"},
			flag:     "o",
			option:   "output",
			expected: "cheese",
		},
		{
			args:     []string{"-xflag", "--output", "last"},
			flag:     "o",
			option:   "output",
			expected: "last",
		},
		{
			args:     []string{"-xflag", "--output=cheese", "something"},
			flag:     "o",
			option:   "output",
			expected: "cheese",
		},
		{
			args:     []string{"cheese"},
			flag:     "o",
			option:   "output",
			expected: "",
		},
		{
			args:     []string{"-xflag", "--o", "cheese", "something"},
			flag:     "o",
			option:   "output",
			expected: "",
		},
		{
			args:     []string{"-xflag", "--output"},
			flag:     "o",
			option:   "output",
			expected: "",
		},
	}

	for _, tc := range testCases {
		got := options.ArgumentsOptionValue(tc.args, tc.flag, tc.option)
		assert.Equal(t, tc.expected, got, "for -%s --%s with args: %v", tc.flag, tc.option, tc.args)

		t.Logf("got value: '%s' for -%s --%s with args: %v\n", got, tc.flag, tc.option, tc.args)
	}
}
