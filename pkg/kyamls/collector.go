package kyamls

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func Collect(dir string, filter Filter) ([]yaml.RNode, error) {
	filterFn, err := filter.ToFilterFn()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create filter")
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
		resources = append(resources, *node)
		return nil
	})
	return resources, nil
}
