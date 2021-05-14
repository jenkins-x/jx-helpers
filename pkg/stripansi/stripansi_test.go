package stripansi_test

import (
	"testing"

	"github.com/fatih/color"
	"github.com/jenkins-x/jx-helpers/v3/pkg/stripansi"
	"github.com/jenkins-x/jx-helpers/v3/pkg/termcolor"
	"github.com/stretchr/testify/assert"
)

func TestStripAnsiColors(t *testing.T) {
	old := color.NoColor
	color.NoColor = false

	testCases := []struct {
		input    string
		expected string
	}{
		{
			input:    termcolor.ColorInfo("hello"),
			expected: "hello",
		},
	}

	for _, tc := range testCases {
		got := stripansi.Strip(tc.input)
		assert.Equal(t, tc.expected, got, "for input %s", tc.input)
		t.Logf("stripped %s to get %s\n", tc.input, got)
	}

	color.NoColor = old
}
