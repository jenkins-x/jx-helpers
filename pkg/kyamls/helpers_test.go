package kyamls

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	"testing"
)

func TestGetLabels(t *testing.T) {
	path := "test_data/helpers/labelled-secret.yaml"
	rNode, readErr := yaml.ReadFile(path)
	require.NoError(t, readErr)

	labels, err := GetLabels(rNode, path)
	require.NoError(t, err)

	value, _ := labels["gitops/type"]
	assert.Equal(t, "\"top-secret\"", value)
}

func TestGetAnnotations(t *testing.T) {
	path := "test_data/helpers/labelled-secret.yaml"
	rNode, readErr := yaml.ReadFile(path)
	require.NoError(t, readErr)

	annotations, err := GetAnnotations(rNode, path)
	require.NoError(t, err)

	value, _ := annotations["size"]
	assert.Equal(t, "small", value)

	value, _ = annotations["what"]
	assert.Equal(t, "\"put.in\"", value)
}

func TestGetMetadataMap(t *testing.T) {
	type test struct {
		path                      string
		expectedAnnotationsErrMsg string
		expectedLabelsErrMsg      string
	}

	tests := []test{
		{
			path:                      "test_data/helpers/empty-file.yaml",
			expectedAnnotationsErrMsg: "failed to get annotations: wrong Node Kind for  expected: MappingNode was : value: {null}",
			expectedLabelsErrMsg:      "failed to get labels: wrong Node Kind for  expected: MappingNode was : value: {null}",
		},
		{
			path:                      "test_data/helpers/invalid-value-type.yaml",
			expectedAnnotationsErrMsg: "failed to get annotations: wrong Node Kind for metadata expected: MappingNode was ScalarNode: value: {\"hello\"}",
			expectedLabelsErrMsg:      "failed to get labels: wrong Node Kind for metadata expected: MappingNode was ScalarNode: value: {\"hello\"}",
		},
	}

	for _, test := range tests {
		rNode, _ := yaml.ReadFile(test.path)
		_, annotationsErr := GetAnnotations(rNode, test.path)
		_, labelsErr := GetLabels(rNode, test.path)

		if test.expectedAnnotationsErrMsg != "" {
			assert.Equal(t, test.expectedAnnotationsErrMsg, annotationsErr.Error())
		}
		if test.expectedAnnotationsErrMsg != "" {
			assert.Equal(t, test.expectedLabelsErrMsg, labelsErr.Error())
		}
	}
}
