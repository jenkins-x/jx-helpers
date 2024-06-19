package kyamls

import (
	"fmt"
	"strings"

	"github.com/jenkins-x/jx-helpers/v3/pkg/stringhelpers"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// Filter for filtering
type Filter struct {
	Kinds          []string
	KindsIgnore    []string
	Names          []string
	SelectTarget   string
	Selector       map[string]string
	InvertSelector bool
}

// ToFilterFn creates a filter function
func (f *Filter) ToFilterFn() (func(node *yaml.RNode, path string) (bool, error), error) {
	mapPath := []string{"metadata", "labels"}
	if f.SelectTarget != "" {
		mapPath = strings.Split(f.SelectTarget, ".")
	}
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
		labels, err := GetMap(node, path, mapPath)
		if err != nil {
			return false, fmt.Errorf("failed to get labels for %s: %w", path, err)
		}
		if labels == nil && f.InvertSelector {
			return true, nil
		} else if labels == nil && !f.InvertSelector {
			return false, nil
		}

		matchCount := 0
		for k, v := range f.Selector {
			actual := labels[k]
			match := trimQuotes(actual) == trimQuotes(v)
			if !f.InvertSelector && !match {
				return false, nil
			} else if match {
				matchCount++
			}

		}
		if f.InvertSelector && matchCount == len(f.Selector) {
			return false, nil
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
	f.AddKindFlags(cmd)
	f.AddSelectorFlags(cmd)
}

// AddKindFlags add CLI flags for specifying the kind part of a filter
func (f *Filter) AddKindFlags(cmd *cobra.Command) {
	cmd.Flags().StringArrayVarP(&f.Kinds, "kind", "k", nil,
		"adds Kubernetes resource kinds to filter on. For kind expressions see:"+
			" https://github.com/jenkins-x/jx-helpers/tree/master/docs/kind_filters.md")
	cmd.Flags().StringArrayVarP(&f.KindsIgnore, "kind-ignore", "", nil,
		"adds Kubernetes resource kinds to exclude. For kind expressions see:"+
			" https://github.com/jenkins-x/jx-helpers/tree/master/docs/kind_filters.md")
}

// AddSelectorFlags add CLI flags for specifying the selector part of a filter
func (f *Filter) AddSelectorFlags(cmd *cobra.Command) {
	cmd.Flags().StringToStringVarP(&f.Selector, "selector", "", nil,
		"adds Kubernetes label selector to filter on, e.g. --selector app=pusher-wave,heritage=Helm")
	cmd.Flags().StringVar(&f.SelectTarget, "selector-target", "",
		"sets which path in the Kubernetes resources to select on instead of metadata.labels.")
	cmd.Flags().BoolVarP(&f.InvertSelector, "invert-selector", "", false,
		"inverts the effect of selector to exclude resources matched by selector")
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
