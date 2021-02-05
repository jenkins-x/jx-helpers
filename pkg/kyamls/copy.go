package kyamls

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func CopyFiles(dir string, filter Filter, suffix string, labels map[string]string) error {
	filterFn, err := filter.ToFilterFn()
	if err != nil {
		return errors.Wrap(err, "failed to create filter")
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
			return errors.Wrapf(err, "failed to load file %s", path)
		}

		if filterFn != nil {
			flag, err := filterFn(node, path)
			if err != nil {
				return errors.Wrapf(err, "failed to evaluate filter on file %s", path)
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
			return errors.Wrapf(err, "failed to save %s", path)
		}

		return nil
	})
	if err != nil {
		return errors.Wrapf(err, "failed to modify files in dir %s", dir)
	}
	return nil
}
