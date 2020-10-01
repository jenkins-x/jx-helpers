package inputfactory

import (
	"github.com/jenkins-x/jx-helpers/v3/pkg/input"
	"github.com/jenkins-x/jx-helpers/v3/pkg/input/batch"
	"github.com/jenkins-x/jx-helpers/v3/pkg/input/survey"
	"github.com/jenkins-x/jx-helpers/v3/pkg/options"
)

// NewInput creates a new input interface depending on if batch mode is enabled or not
func NewInput(o *options.BaseOptions) input.Interface {
	if o != nil && o.BatchMode {
		return batch.NewBatchInput()
	}
	return survey.NewInput()
}
