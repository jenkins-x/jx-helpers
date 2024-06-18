package kyamls

import (
	"fmt"
	"strings"

	"github.com/jenkins-x/jx-logging/v3/pkg/log"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var quotes = []string{"'", "\""}

// GetKind finds the Kind of the node at the given path
func GetKind(node *yaml.RNode, path string) string {
	return GetStringField(node, path, "kind")
}

// GetAPIVersion finds the API Version of the node at the given path
func GetAPIVersion(node *yaml.RNode, path string) string {
	return GetStringField(node, path, "apiVersion")
}

// GetName returns the name from the metadata
func GetName(node *yaml.RNode, path string) string {
	return GetStringField(node, path, "metadata", "name")
}

// GetNamespace returns the namespace from the metadata
func GetNamespace(node *yaml.RNode, path string) string {
	return GetStringField(node, path, "metadata", "namespace")
}

// GetMap return the content of mapPath in node as a map
func GetMap(node *yaml.RNode, filePath string, mapPath []string) (map[string]string, error) {
	mapNode, err := node.Pipe(yaml.Lookup(mapPath...))
	if err != nil {
		return nil, fmt.Errorf("failed to get %v: %w", mapPath, err)
	}
	m := map[string]string{}
	if mapNode == nil {
		return m, nil
	}
	err = mapNode.VisitFields(func(node *yaml.MapNode) error {
		v := ""
		k, err := node.Key.String()
		if err != nil {
			return fmt.Errorf("failed to find %v key for path %s: %w", mapPath, filePath, err)
		}
		v, err = node.Value.String()
		if err != nil {
			return fmt.Errorf("failed to find %v %s value for path %s: %w", mapPath, k, filePath, err)
		}
		m[strings.TrimSpace(k)] = strings.TrimSpace(v)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get %v: %w", mapPath, err)
	}
	return m, nil
}

// GetStringField returns the given field from the node or returns a blank string if the field cannot be found
func GetStringField(node *yaml.RNode, path string, fields ...string) string {
	answer := ""
	valueNode, err := node.Pipe(yaml.Lookup(fields...))
	if err != nil {
		log.Logger().Debugf("failed to read field %s for path %s", JSONPath(fields...), path)
	}
	if valueNode != nil {
		var err error
		answer, err = valueNode.String()
		if err != nil {
			log.Logger().Warnf("failed to get string value of field %s for path %s", JSONPath(fields...), path)
		}
	}
	return TrimSpaceAndQuotes(answer)
}

// TrimSpaceAndQuotes trims any whitespace and quotes around a value
func TrimSpaceAndQuotes(answer string) string {
	text := strings.TrimSpace(answer)
	for _, q := range quotes {
		if strings.HasPrefix(text, q) && strings.HasSuffix(text, q) {
			return strings.TrimPrefix(strings.TrimSuffix(text, q), q)
		}
	}
	return text
}

// IsClusterKind returns true if the kind is a cluster kind
func IsClusterKind(kind string) bool {
	return kind == "" || kind == "CustomResourceDefinition" || kind == "Namespace" || strings.HasPrefix(kind, "Cluster")
}

// IsCustomResourceDefinition returns true if the kind is a customresourcedefinition
func IsCustomResourceDefinition(kind string) bool {
	return kind == "CustomResourceDefinition"
}

// JSONPath returns the fields separated by dots
func JSONPath(fields ...string) string {
	return strings.Join(fields, ".")
}
