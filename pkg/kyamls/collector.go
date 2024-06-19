package kyamls

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func Collect(dir string, filter Filter) ([]yaml.RNode, error) {
	filterFn, err := filter.ToFilterFn()
	if err != nil {
		return nil, fmt.Errorf("failed to create filter: %w", err)
	}

	var resources []yaml.RNode
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info == nil || info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
			return nil
		}
		node, err := yaml.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to load file %s: %w", path, err)
		}

		if filterFn != nil {
			flag, err := filterFn(node, path)
			if err != nil {
				return fmt.Errorf("failed to evaluate filter on file %s: %w", path, err)
			}
			if !flag {
				return nil
			}
		}
		resources = append(resources, *node)
		return nil
	})
	return resources, nil
}
