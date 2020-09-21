package inputfactory

import (
	"github.com/jenkins-x/jx-helpers/pkg/input"
	"github.com/jenkins-x/jx-helpers/pkg/input/batch"
	"github.com/jenkins-x/jx-helpers/pkg/input/survey"
	"github.com/jenkins-x/jx-helpers/pkg/options"
)

// NewInput creates a new input interface depending on if batch mode is enabled or not
func NewInput(o *options.BaseOptions) input.Interface {
	if o != nil && o.BatchMode {
		return batch.NewBatchInput()
	}
	return survey.NewInput()
}
