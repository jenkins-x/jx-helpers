package kyamls

import (
	"path/filepath"
	"testing"

	"github.com/jenkins-x/jx-helpers/v3/pkg/files"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestDeleteFiles(t *testing.T) {
	tmpDir := t.TempDir()
	err := files.CopyDir("test_data", tmpDir, true)
	assert.NoError(t, err)

	err = DeleteFiles(tmpDir, func(node *yaml.RNode, path string) (bool, error) {
		return true, nil
	}, Filter{Kinds: []string{"Deployment"}})

	assert.NoError(t, err)
	assert.FileExists(t, filepath.Join(tmpDir, "configmap.yaml"))
	assert.FileExists(t, filepath.Join(tmpDir, "service.yaml"))
	assert.NoFileExists(t, filepath.Join(tmpDir, "deployment.yaml"))

}

func TestDeleteFilesWithDeleteFn(t *testing.T) {
	tmpDir := t.TempDir()
	err := files.CopyDir("test_data", tmpDir, true)
	assert.NoError(t, err)

	deleteFn := func(node *yaml.RNode, path string) (bool, error) {
		name := GetName(node, path)
		if name == "onion" {
			return true, nil
		}
		return false, nil
	}

	err = DeleteFiles(tmpDir, deleteFn, Filter{})

	assert.NoError(t, err)
	assert.NoFileExists(t, filepath.Join(tmpDir, "configmap.yaml"))
	assert.FileExists(t, filepath.Join(tmpDir, "service.yaml"))
	assert.FileExists(t, filepath.Join(tmpDir, "deployment.yaml"))

}
