package templater_test

import (
	"testing"

	"github.com/jenkins-x/jx-helpers/pkg/templater"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvaluate(t *testing.T) {
	templateData := map[string]interface{}{
		"Name": "World",
	}

	output, err := templater.Evaluate(nil, templateData, "hello {{ .Name }}!", "myfile.gotmpl", "test template")
	require.NoError(t, err, "failed to evaluate template")
	assert.Equal(t, "hello World!", output, "template output")

	t.Logf("template generated %s\n", output)
}
