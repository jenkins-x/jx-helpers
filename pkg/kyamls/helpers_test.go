package kyamls_test

import (
	"testing"

	"github.com/jenkins-x/jx-helpers/v3/pkg/kyamls"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestGetLabels(t *testing.T) {
	path := "test_data/helpers/labelled-secret.yaml"
	rNode, readErr := yaml.ReadFile(path)
	require.NoError(t, readErr)

	labels, err := kyamls.GetMap(rNode, path, []string{"metadata", "labels"})
	require.NoError(t, err)

	value, _ := labels["gitops/type"]
	assert.Equal(t, "\"top-secret\"", value)
}

func TestGetAnnotations(t *testing.T) {
	path := "test_data/helpers/labelled-secret.yaml"
	rNode, readErr := yaml.ReadFile(path)
	require.NoError(t, readErr)

	annotations, err := kyamls.GetMap(rNode, path, []string{"metadata", "annotations"})
	require.NoError(t, err)

	value, _ := annotations["size"]
	assert.Equal(t, "small", value)

	value, _ = annotations["what"]
	assert.Equal(t, "\"put.in\"", value)
}

func TestGetMetadataMap(t *testing.T) {
	type test struct {
		path                   string
		expectedAnnotationsErr bool
		expectedLabelsErr      bool
	}

	tests := []test{
		{
			path:                   "test_data/helpers/empty-file.yaml",
			expectedAnnotationsErr: true,
			expectedLabelsErr:      true,
		},
		{
			path:                   "test_data/helpers/invalid-value-type.yaml",
			expectedAnnotationsErr: true,
			expectedLabelsErr:      true,
		},
	}

	for _, test := range tests {
		rNode, _ := yaml.ReadFile(test.path)
		_, annotationsErr := kyamls.GetMap(rNode, test.path, []string{"metadata", "annotations"})
		_, labelsErr := kyamls.GetMap(rNode, test.path, []string{"metadata", "labels"})

		if test.expectedAnnotationsErr {
			assert.NotNil(t, labelsErr)
		}
		if test.expectedAnnotationsErr {
			assert.NotNil(t, annotationsErr)
		}
	}
}
