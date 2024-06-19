package files

import (
	"fmt"
	"os"

	"github.com/jenkins-x/jx-logging/v3/pkg/log"
)

// RecreateSymLink removes any old symlink/binary if they exist and
// creates a new symlink from the given source
func RecreateSymLink(src, target string) error {
	exists, err := FileExists(target)
	if err != nil {
		return fmt.Errorf("failed to check if %s exists: %w", target, err)
	}
	if exists {
		err = os.Remove(target)
		if err != nil {
			return fmt.Errorf("failed to remove %s: %w", target, err)
		}
	}
	err = os.Symlink(src, target)
	if err != nil {
		return fmt.Errorf("failed to create symlink from %s to %s: %w", src, target, err)
	}
	log.Logger().Infof("created symlink from %s => %s", target, src)
	return nil
}
