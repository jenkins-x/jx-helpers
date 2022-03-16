package kyamls

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/jenkins-x/jx-helpers/v3/pkg/files"
	"github.com/jenkins-x/jx-helpers/v3/pkg/testhelpers"
	"github.com/stretchr/testify/assert"
)

func TestCopyFiles(t *testing.T) {
	tmpDir := t.TempDir()
	err := files.CopyDir("test_data/copy_delete_filter_test", tmpDir, true)
	assert.NoError(t, err)

	err = CopyFiles(tmpDir, Filter{}, ".bak", map[string]string{
		"clone": "true",
	})

	assert.NoError(t, err)

	for _, filename := range []string{
		"configmap",
		"deployment",
		"service",
	} {
		assert.FileExists(t, filepath.Join(tmpDir, fmt.Sprintf("%s.bak.yaml", filename)))
		testhelpers.AssertTextFilesEqual(t, filepath.Join(tmpDir, "copy_expected", fmt.Sprintf("%s.bak.yaml", filename)), filepath.Join(tmpDir, fmt.Sprintf("%s.bak.yaml", filename)), "generated file: "+filename)

	}

}
