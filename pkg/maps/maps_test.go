// +build unit

package maps_test

import (
	"testing"

	"github.com/jenkins-x/jx-helpers/pkg/maps"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeyValuesToMap(t *testing.T) {
	t.Parallel()

	m := maps.KeyValuesToMap([]string{"foo=bar", "whatnot=cheese"})
	require.NotNil(t, m, "util.KeysToMap() returned nil")
	assert.Equal(t, "bar", m["foo"], "map does not contain foo=bar")
	assert.Equal(t, "cheese", m["whatnot"], "map does not contain whatnot=cheese")
}

func TestMapToKeyValues(t *testing.T) {
	t.Parallel()

	values := maps.MapToKeyValues(map[string]string{
		"foo":     "bar",
		"whatnot": "cheese",
	})

	assert.Equal(t, []string{"foo=bar", "whatnot=cheese"}, values, "output of util.MapToKeyValues()")
}
