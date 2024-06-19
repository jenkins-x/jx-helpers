package kyamls

import (
	"fmt"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// SetStringValue sets the string value at the given path
func SetStringValue(node *yaml.RNode, path string, value string, fields ...string) error {
	err := node.PipeE(yaml.LookupCreate(yaml.ScalarNode, fields...), yaml.FieldSetter{StringValue: value})
	if err != nil {
		return fmt.Errorf("failed to set field %s to %s at path %s: %w", JSONPath(fields...), value, path, err)
	}
	return nil
}
