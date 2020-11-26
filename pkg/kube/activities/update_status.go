package activities

import (
	"time"

	v1 "github.com/jenkins-x/jx-api/v4/pkg/apis/core/v4beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// UpdateStatus updates the pipeline activity status if any of the steps have failed or they have all completed
func UpdateStatus(activity *v1.PipelineActivity, containersTerminated bool, onCompleteCallback func(activity *v1.PipelineActivity)) {
	spec := &activity.Spec
	var biggestFinishedAt metav1.Time

	var stageSteps []v1.CoreActivityStep
	for i := range spec.Steps {
		step := &spec.Steps[i]
		stage := step.Stage
		if stage != nil {
			stage.Status = UpdateStepsStatus(stage.Status, stage.Steps)
			stageSteps = append(stageSteps, stage.CoreActivityStep)

			if stage.StartedTimestamp != nil && spec.StartedTimestamp == nil {
				spec.StartedTimestamp = stage.StartedTimestamp
			}
			if stage.CompletedTimestamp != nil {
				t := stage.CompletedTimestamp
				if !t.IsZero() {
					if biggestFinishedAt.IsZero() || t.After(biggestFinishedAt.Time) {
						biggestFinishedAt = *t
					}
				}
			}
		}
	}
	spec.Status = UpdateStepsStatus(spec.Status, stageSteps)

	// lets make sure we have a completed time
	if spec.Status.IsTerminated() {
		if spec.CompletedTimestamp == nil {
			if biggestFinishedAt.IsZero() {
				biggestFinishedAt.Time = time.Now()
			}
			spec.CompletedTimestamp = &biggestFinishedAt
		}
	}

	if containersTerminated {
		switch spec.Status {
		case v1.ActivityStatusTypeSucceeded, v1.ActivityStatusTypeAborted, v1.ActivityStatusTypeError, v1.ActivityStatusTypeFailed:

		default:
			spec.Status = v1.ActivityStatusTypeAborted
		}

		if spec.Status.IsTerminated() && spec.CompletedTimestamp == nil {
			if !biggestFinishedAt.IsZero() {
				spec.CompletedTimestamp = &biggestFinishedAt
			}

			// log that the build completed
			if onCompleteCallback != nil {
				onCompleteCallback(activity)
			}
		}
	}
}

// UpdateStepsStatus updates the status for the given status and steps
func UpdateStepsStatus(status v1.ActivityStatusType, steps []v1.CoreActivityStep) v1.ActivityStatusType {
	if status == v1.ActivityStatusTypeNone {
		status = v1.ActivityStatusTypePending
	}
	if len(steps) == 0 {
		return status
	}
	allCompleted := true
	failed := false
	running := false
	for _, s := range steps {
		switch s.Status {
		case v1.ActivityStatusTypeRunning:
			running = true
			allCompleted = false

		case v1.ActivityStatusTypePending, v1.ActivityStatusTypeNotExecuted, v1.ActivityStatusTypeNone, v1.ActivityStatusTypeWaitingForApproval:
			allCompleted = false

		case v1.ActivityStatusTypeAborted, v1.ActivityStatusTypeError, v1.ActivityStatusTypeFailed:
			failed = true
		}
	}

	if failed {
		return v1.ActivityStatusTypeFailed
	}
	if running {
		return v1.ActivityStatusTypeRunning
	}
	if allCompleted {
		return v1.ActivityStatusTypeSucceeded
	}
	return status
}
