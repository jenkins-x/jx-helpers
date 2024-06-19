package kyamls

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func CopyFiles(dir string, filter Filter, suffix string, labels map[string]string) error {
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

		for l, v := range labels {
			label := yaml.SetLabel(l, v)
			_, err = label.Filter(node)
		}

		extension := filepath.Ext(path)
		existingFilenameWithoutExtension := strings.TrimSuffix(filepath.Base(path), extension)
		copyPath := filepath.Join(filepath.Dir(path), strings.Join([]string{existingFilenameWithoutExtension, suffix, extension}, ""))

		err = yaml.WriteFile(node, copyPath)
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
