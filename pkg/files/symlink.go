package files

import (
	"os"

	"github.com/jenkins-x/jx-logging/pkg/log"
	"github.com/pkg/errors"
)

// RecreateSymLink removes any old symlink/binary if they exist and
// creates a new symlink from the given source
func RecreateSymLink(src, target string) error {
	exists, err := FileExists(target)
	if err != nil {
		return errors.Wrapf(err, "failed to check if %s exists", target)
	}
	if exists {
		err = os.Remove(target)
		if err != nil {
			return errors.Wrapf(err, "failed to remove %s", target)
		}
	}
	err = os.Symlink(src, target)
	if err != nil {
		return errors.Wrapf(err, "failed to create symlink from %s to %s", src, target)
	}
	log.Logger().Infof("created symlink from %s => %s", target, src)
	return nil
}
