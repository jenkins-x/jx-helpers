package kyamls

import (
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	"testing"
)

func TestGetLabels(t *testing.T) {
	path := "test_data/helpers/labelled-secret.yaml"
	rNode, _ := yaml.ReadFile(path)

	labels, _ := GetLabels(rNode, path)

	value, _ := labels["gitops/type"]
	assert.Equal(t, "\"top-secret\"", value)
}

func TestGetAnnotations(t *testing.T) {
	path := "test_data/helpers/labelled-secret.yaml"
	rNode, _ := yaml.ReadFile(path)

	annotations, _ := GetAnnotations(rNode, path)

	value, _ := annotations["size"]
	assert.Equal(t, "small", value)

	value, _ = annotations["what"]
	assert.Equal(t, "\"put.in\"", value)
}
