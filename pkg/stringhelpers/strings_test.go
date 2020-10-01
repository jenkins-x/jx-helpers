// +build unit

package stringhelpers_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jenkins-x/jx-helpers/v3/pkg/stringhelpers"
	"github.com/stretchr/testify/assert"
)

type regexSplitData struct {
	input     string
	separator string
	expected  []string
}

func TestRegexpSplit(t *testing.T) {
	testCases := []regexSplitData{
		{
			"foo/bar", ":|/", []string{"foo", "bar"},
		},
		{
			"foo:bar", ":|/", []string{"foo", "bar"},
		},
	}
	for _, data := range testCases {
		actual := stringhelpers.RegexpSplit(data.input, data.separator)
		assert.Equal(t, data.expected, actual, "Split did not match for input %s with separator %s", data.input, data.separator)
		//t.Logf("split %s with separator %s into %#v", data.input, data.separator, actual)
	}
}

func TestStringIndices(t *testing.T) {
	assertStringIndices(t, "foo/bar", "/", []int{3})
	assertStringIndices(t, "/foo/bar", "/", []int{0, 4})
}

func TestRemoveStringFromSlice(t *testing.T) {
	beatles := []string{"paul", "john", "ringo", "george"}
	betterBeatles := stringhelpers.RemoveStringFromSlice(beatles, "ringo")

	assert.NotContains(t, betterBeatles, "ringo", "Ringo shouldn't be in the beatles")
	assert.Equal(t, 3, len(betterBeatles))
}

func TestRemoveStringFromSlice_NotAMember(t *testing.T) {
	beatles := []string{"paul", "john", "ringo", "george"}
	betterBeatles := stringhelpers.RemoveStringFromSlice(beatles, "Freddy")

	assert.Equal(t, betterBeatles, beatles)
}

func assertStringIndices(t *testing.T, text string, sep string, expected []int) {
	actual := stringhelpers.StringIndexes(text, sep)
	assert.Equal(t, expected, actual, "Failed to evaluate StringIndices(%s, %s)", text, sep)
}

func TestDiffSlices(t *testing.T) {
	// no inserts or deletes
	assertDiffSlice(t, []string{"a", "b", "c"}, []string{"a", "b", "c"}, []string{}, []string{})

	// all inserts no deletes
	assertDiffSlice(t, []string{}, []string{"a", "b", "c"}, []string{}, []string{"a", "b", "c"})

	// no inserts all deletes
	assertDiffSlice(t, []string{"a", "b", "c"}, []string{}, []string{"a", "b", "c"}, []string{})

	// all inserts and all deletes
	assertDiffSlice(t, []string{"a", "b", "c"}, []string{"d", "e", "f"}, []string{"a", "b", "c"}, []string{"d", "e", "f"})

	// remove single in the middle
	assertDiffSlice(t, []string{"a", "b", "c"}, []string{"b"}, []string{"a", "c"}, []string{})

	// replace single in the middle
	assertDiffSlice(t, []string{"a", "b", "c"}, []string{"a", "x", "c"}, []string{"b"}, []string{"x"})
}

func assertDiffSlice(t *testing.T, originalSlice, newSlice, removed, added []string) {
	toDelete, toInsert := stringhelpers.DiffSlices(originalSlice, newSlice)
	assert.Equal(t, toDelete, removed, fmt.Sprintf("removal incorrect - original [%s] new [%s]", strings.Join(originalSlice, ", "), strings.Join(newSlice, ", ")))
	assert.Equal(t, toInsert, added, fmt.Sprintf("insert incorrect - original [%s] new [%s]", strings.Join(originalSlice, ", "), strings.Join(newSlice, ", ")))
}

func TestYesNo(t *testing.T) {
	assert.Equal(t, "Yes", stringhelpers.YesNo(true), "Yes boolean conversion")
	assert.Equal(t, "No", stringhelpers.YesNo(false), "No boolean conversion")
}

func TestExtractKeyValuePairs(t *testing.T) {
	type testData struct {
		keyValueArray []string
		keyValueMap   map[string]string
		expectError   bool
	}

	testCases := []testData{
		{
			[]string{}, map[string]string{}, false,
		},
		{
			[]string{"foo=bar"}, map[string]string{"foo": "bar"}, false,
		},
		{
			[]string{"foo=bar", "snafu=tarfu"}, map[string]string{"foo": "bar", "snafu": "tarfu"}, false,
		},
		{
			[]string{"foo=bar", "snafu"}, map[string]string{}, true,
		},
	}
	for _, data := range testCases {
		actual, err := stringhelpers.ExtractKeyValuePairs(data.keyValueArray, "=")
		if data.expectError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
		assert.Equal(t, data.keyValueMap, actual)
	}
}

func TestStripTrailingSlash(t *testing.T) {
	t.Parallel()

	url := "http://some.url.com/"
	assert.Equal(t, stringhelpers.StripTrailingSlash(url), "http://some.url.com")

	url = "http://some.other.url.com"
	assert.Equal(t, stringhelpers.StripTrailingSlash(url), "http://some.other.url.com")
}

func Test_ToCamelCase(t *testing.T) {
	assert.Equal(t, stringhelpers.ToCamelCase("my-super-name"), "MySuperName")
}

func TestHasPrefix(t *testing.T) {
	assert.True(t, stringhelpers.HasPrefix("some text", "ignored", "some ", "another"), "should have found prefix for 2nd prefix")
	assert.False(t, stringhelpers.HasPrefix("cheese text", "ignored", "some ", "another"), "should not have matched a prefix")
}

func TestHasSuffix(t *testing.T) {
	assert.True(t, stringhelpers.HasSuffix("some text", "ignored", " text", "another"), "should have found suffix for 2nd suffix")
	assert.False(t, stringhelpers.HasSuffix("cheese toast", "ignored", " something", "another"), "should not have matched a suffix")
}
