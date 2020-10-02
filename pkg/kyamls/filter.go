package kyamls

import (
	"strings"

	"github.com/jenkins-x/jx-helpers/v3/pkg/stringhelpers"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// Filter for filtering
type Filter struct {
	Kinds       []string
	KindsIgnore []string
	Names       []string
	Selector    map[string]string
}

// ToFilterFn creates a filter function
func (f *Filter) ToFilterFn() (func(node *yaml.RNode, path string) (bool, error), error) {
	kf := f.Parse()
	return func(node *yaml.RNode, path string) (bool, error) {
		name := GetName(node, path)
		if len(f.Names) > 0 && stringhelpers.StringArrayIndex(f.Names, name) < 0 {
			return false, nil
		}

		for _, filter := range kf.KindsIgnore {
			if filter.Matches(node, path) {
				return false, nil
			}
		}

		if len(kf.Kinds) > 0 {
			matched := false
			for _, filter := range kf.Kinds {
				if filter.Matches(node, path) {
					matched = true
					break
				}
			}
			if !matched {
				return false, nil
			}
		}

		// lets check if there's a selector
		if f.Selector == nil {
			return true, nil
		}
		labels, err := GetLabels(node, path)
		if err != nil {
			return false, errors.Wrapf(err, "failed to get labels for %s", path)
		}
		if labels == nil {
			return false, nil
		}
		for k, v := range f.Selector {
			actual := labels[k]
			if trimQuotes(actual) != trimQuotes(v) {
				return false, nil
			}
		}
		return true, nil

	}, nil
}

func trimQuotes(text string) string {
	for _, q := range []string{"'", "\""} {
		if strings.HasPrefix(text, q) && strings.HasSuffix(text, q) {
			return text[1 : len(text)-1]
		}
	}
	return text
}

// AddFlags add CLI flags for specifying a filter
func (f *Filter) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringArrayVarP(&f.Kinds, "kind", "k", nil, "adds Kubernetes resource kinds to filter on. For kind expressions see: https://github.com/jenkins-x/jx-helpers/v3/tree/master/docs/kind_filters.md")
	cmd.Flags().StringArrayVarP(&f.KindsIgnore, "kind-ignore", "", nil, "adds Kubernetes resource kinds to exclude. For kind expressions see: https://github.com/jenkins-x/jx-helpers/v3/tree/master/docs/kind_filters.md")
}

// Parse parses the filter strings
func (f *Filter) Parse() APIVersionKindsFilter {
	r := APIVersionKindsFilter{}
	for _, text := range f.Kinds {
		r.Kinds = append(r.Kinds, ParseKindFilter(text))
	}
	for _, text := range f.KindsIgnore {
		r.KindsIgnore = append(r.KindsIgnore, ParseKindFilter(text))
	}
	return r
}

// APIVersionKindsFilter a filter of kinds and/or API versions
type APIVersionKindsFilter struct {
	Kinds       []KindFilter
	KindsIgnore []KindFilter
}

// KindFilter a filter on a kind and an optional APIVersion
type KindFilter struct {
	APIVersion *string
	Kind       *string
}

// ParseKindFilter parses a kind filter
func ParseKindFilter(text string) KindFilter {
	idx := strings.LastIndex(text, "/")
	if idx >= 0 {
		apiVersion := text[0:idx]
		kind := text[idx+1:]
		if len(kind) == 0 {
			return KindFilter{
				APIVersion: &apiVersion,
			}
		}
		return KindFilter{
			APIVersion: &apiVersion,
			Kind:       &kind,
		}
	}
	return KindFilter{
		Kind: &text,
	}
}

// Matches returns true if this node matches the filter
func (f *KindFilter) Matches(node *yaml.RNode, path string) bool {
	if f.Kind != nil {
		kind := GetKind(node, path)
		if kind != *f.Kind {
			return false
		}
	}
	if f.APIVersion != nil {
		apiVersion := GetAPIVersion(node, path)
		actual := *f.APIVersion
		if apiVersion != actual && !strings.HasPrefix(apiVersion, actual) {
			return false
		}
	}
	return true
}
