package survey_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Netflix/go-expect"
	pseudotty "github.com/creack/pty"
	"github.com/hinshun/vt10x"
	"github.com/jenkins-x/jx-helpers/v3/pkg/input/survey"
	"github.com/stretchr/testify/assert"
)

// Modified from https://github.com/suarezjulian/survey_test/blob/main/main_test.go

func TestPickNameWithDefault(t *testing.T) {
	names := []string{"first", "second", "third"}
	testCases := []struct {
		description string
		selection   string
		expect      string
	}{
		{"Select nothing, hence default", "", names[0]},
		{"Select first element", names[0], names[0]},
		{"Select Second element", names[1], names[1]},
		{"Select Last element", names[2], names[2]},
	}
	for _, v := range testCases {
		t.Log(v.description)
		pty, tty, err := pseudotty.Open()
		if err != nil {
			t.Fatalf("failed to open pseudotty: %v", err)
		}
		var donec chan struct{}
		term := vt10x.New(vt10x.WithWriter(tty))
		if err != nil {
			t.Errorf("Unexpected error %s", err)
		}

		console, err := expect.NewConsole(
			expect.WithStdin(pty),
			expect.WithStdout(term),
			expect.WithCloser(pty, tty),
			expectNoError(t),
			expect.WithDefaultTimeout(time.Second),
		)
		if err != nil {
			t.Fatalf("failed to create console: %v", err)
		}

		defer console.Close()
		donec = make(chan struct{})
		go func(selection, expected string) {
			defer close(donec)
			console.ExpectString("Pick")
			console.SendLine(selection)
			console.ExpectString(expected)
			console.ExpectEOF()
		}(v.selection, v.expect)

		c := survey.NewInputFrom(console.Tty(), console.Tty(), console.Tty())

		answer, err := c.PickNameWithDefault(names, "Pick", "", "helpful text")
		assert.NoError(t, err)

		if err := console.Tty().Close(); err != nil {
			t.Errorf("error closing Tty: %v", err)
		}
		<-donec

		output := strings.TrimSpace(expect.StripTrailingEmptyLines(term.String()))
		expectedOutput := fmt.Sprintf("? Pick %s", answer)
		if output != expectedOutput {
			t.Fatalf("Unexpected output.\nExpected: \n%s ; \nFound: \n%s", expectedOutput, output)
		}
	}
}

func expectNoError(t *testing.T) expect.ConsoleOpt {
	return expect.WithExpectObserver(
		func(matchers []expect.Matcher, buf string, err error) {
			if err == nil {
				return
			}
			if len(matchers) == 0 {
				t.Fatalf("Error occurred while matching %q: %s\n", buf, err)
			} else {
				var criteria []string
				for _, matcher := range matchers {
					criteria = append(criteria, fmt.Sprintf("%q", matcher.Criteria()))
				}
				t.Fatalf("Unexpected output; expected: %s ; got %q: \nError: %s\n", strings.Join(criteria, ", "), buf, err)
			}
		},
	)
}
