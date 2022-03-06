package options

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/jenkins-x/jx-helpers/v3/pkg/termcolor"
)

const (
	// DefaultErrorExitCode is the default exit code in case of an error
	DefaultErrorExitCode = 1

	// DefaultSuggestionsMinimumDistance default distance when generating suggestions if an option/arg is wrong
	DefaultSuggestionsMinimumDistance = 2
)

var (
	fatalErrHandler = fatal
	// ErrExit can be used to exit with a non 0 exit code without any error message
	ErrExit = fmt.Errorf("exit")
)

func InvalidOption(name, value string, values []string) error {
	suggestions := SuggestionsFor(value, values, DefaultSuggestionsMinimumDistance)
	if len(suggestions) > 0 {
		if len(suggestions) == 1 {
			return InvalidOptionf(name, value, "Did you mean:  --%s %s", name, termcolor.ColorInfo(suggestions[0]))
		}
		return InvalidOptionf(name, value, "Did you mean one of: %s", termcolor.ColorInfo(strings.Join(suggestions, ", ")))
	}
	sort.Strings(values)
	return InvalidOptionf(name, value, "Possible values: %s", strings.Join(values, ", "))
}

func InvalidArg(value string, values []string) error {
	suggestions := SuggestionsFor(value, values, DefaultSuggestionsMinimumDistance)
	if len(suggestions) > 0 {
		if len(suggestions) == 1 {
			return InvalidArgf(value, "Did you mean: %s", suggestions[0])
		}
		return InvalidArgf(value, "Did you mean one of: %s", strings.Join(suggestions, ", "))
	}
	sort.Strings(values)
	return InvalidArgf(value, "Possible values: %s", strings.Join(values, ", "))
}

func InvalidArgError(value string, err error) error {
	return InvalidArgf(value, "%s", err)
}

func InvalidArgf(value, message string, a ...interface{}) error {
	text := fmt.Sprintf(message, a...)
	return fmt.Errorf("Invalid argument: %s\n%s", termcolor.ColorInfo(value), text)
}

// InvalidOptionf returns an error that shows the invalid option.
func InvalidOptionf(option string, value interface{}, message string, a ...interface{}) error {
	text := fmt.Sprintf(message, a...)
	return fmt.Errorf("invalid option: --%s %v\n%s", option, value, text)
}

// MissingOption reports a missing command line option using the full name expression.
func MissingOption(name string) error {
	return fmt.Errorf("missing option: --%s", name)
}

// CheckErr prints a user friendly error to STDERR and exits with a non-zero exit code.
func CheckErr(err error) {
	checkErr(err, fatalErrHandler)
}

// checkErr formats a given error as a string and calls the passed handleErr func.
func checkErr(err error, handleErr func(string, int)) {
	switch {
	case err == nil:
		return
	case err == ErrExit:
		handleErr("", DefaultErrorExitCode)
		return
	default:
		handleErr(err.Error(), DefaultErrorExitCode)
	}
}

func fatal(msg string, code int) {
	if len(msg) > 0 {
		// add newline if needed
		if !strings.HasSuffix(msg, "\n") {
			msg += "\n"
		}
		fmt.Fprint(os.Stderr, msg)
	}
	os.Exit(code)
}

func SuggestionsFor(typedName string, values []string, suggestionsMinimumDistance int, explicitSuggestions ...string) []string {
	suggestions := []string{}
	for _, value := range values {
		levenshteinDistance := ld(typedName, value, true)
		suggestByLevenshtein := levenshteinDistance <= suggestionsMinimumDistance
		suggestByPrefix := strings.HasPrefix(strings.ToLower(value), strings.ToLower(typedName))
		if suggestByLevenshtein || suggestByPrefix {
			suggestions = append(suggestions, value)
		}
		for _, explicitSuggestion := range explicitSuggestions {
			if strings.EqualFold(typedName, explicitSuggestion) {
				suggestions = append(suggestions, value)
			}
		}
	}
	return suggestions
}

// ld compares two strings and returns the levenshtein distance between them.
//
// this was copied from vendor/github.com/spf13/cobra/command.go as its not public
func ld(s, t string, ignoreCase bool) int {
	if ignoreCase {
		s = strings.ToLower(s)
		t = strings.ToLower(t)
	}
	d := make([][]int, len(s)+1)
	for i := range d {
		d[i] = make([]int, len(t)+1)
	}
	for i := range d {
		d[i][0] = i
	}
	for j := range d[0] {
		d[0][j] = j
	}
	for j := 1; j <= len(t); j++ {
		for i := 1; i <= len(s); i++ {
			if s[i-1] == t[j-1] {
				d[i][j] = d[i-1][j-1]
			} else {
				min := d[i-1][j]
				if d[i][j-1] < min {
					min = d[i][j-1]
				}
				if d[i-1][j-1] < min {
					min = d[i-1][j-1]
				}
				d[i][j] = min + 1
			}
		}
	}
	return d[len(s)][len(t)]
}

// ArgumentsOptionValue returns the argument
func ArgumentsOptionValue(args []string, flag, option string) string {
	flags := []string{"--" + option, "-" + flag}
	for i := 0; i < len(args); i++ {
		arg := args[i]

		for _, f := range flags {
			feq := f + "="
			if strings.HasPrefix(arg, feq) {
				return arg[len(feq):]
			}

			if arg == f {
				j := i + 1
				if j < len(args) {
					return args[j]
				}
				return ""
			}
		}
	}
	return ""
}
