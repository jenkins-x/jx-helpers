package kyamls

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// ModifyFiles recursively walks the given directory and modifies any suitable file
func ModifyFiles(dir string, modifyFn func(node *yaml.RNode, path string) (bool, error), filter Filter) error {
	filterFn, err := filter.ToFilterFn()
	if err != nil {
		return fmt.Errorf("failed to create filter: %w", err)
	}

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

		modified, err := modifyFn(node, path)
		if err != nil {
			return fmt.Errorf("failed to modify file %s: %w", path, err)
		}

		if !modified {
			return nil
		}

		err = yaml.WriteFile(node, path)
		if err != nil {
			return fmt.Errorf("failed to save %s: %w", path, err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to modify files in dir %s: %w", dir, err)
	}
	return nil
}
