package kyamls

import (
	"strings"

	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/pkg/errors"
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

// GetLabels returns the labels for the given file
func GetLabels(node *yaml.RNode, path string) (map[string]string, error) {
	labels, err := node.Pipe(yaml.Lookup("metadata", "labels"))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get labels")
	}
	m := map[string]string{}
	if labels == nil {
		return m, nil
	}
	err = labels.VisitFields(func(node *yaml.MapNode) error {
		v := ""
		k, err := node.Key.String()
		if err != nil {
			return errors.Wrapf(err, "failed to find label key for path %s", path)
		}
		v, err = node.Value.String()
		if err != nil {
			return errors.Wrapf(err, "failed to find label %s value for path %s", k, path)
		}
		m[strings.TrimSpace(k)] = strings.TrimSpace(v)
		return nil
	})
	return m, nil
}

/// GetStringField returns the given field from the node or returns a blank string if the field cannot be found
func GetStringField(node *yaml.RNode, path string, fields ...string) string {
	answer := ""
	valueNode, err := node.Pipe(yaml.Lookup(fields...))
	if err != nil {
		log.Logger().Warnf("failed to read field %s for path %s", JSONPath(fields...), path)
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
