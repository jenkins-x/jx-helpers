package activities

import (
	v1 "github.com/jenkins-x/jx-api/pkg/apis/jenkins.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// UpdateStatus updates the pipeline activity status if any of the steps have failed or they have all completed
func UpdateStatus(activity *v1.PipelineActivity, containersTerminated bool, onCompleteCallback func (activity *v1.PipelineActivity)) {
	spec := &activity.Spec
	var biggestFinishedAt metav1.Time

	allCompleted := true
	failed := false
	running := false
	for i := range spec.Steps {
		step := &spec.Steps[i]
		stage := step.Stage
		if stage != nil {
			UpdateStageStatus(stage)
			stageFinished := stage.Status.IsTerminated()
			if stage.StartedTimestamp != nil && spec.StartedTimestamp == nil {
				spec.StartedTimestamp = stage.StartedTimestamp
			}
			if stage.CompletedTimestamp != nil {
				t := stage.CompletedTimestamp
				if !t.IsZero() {
					stageFinished = true
					if biggestFinishedAt.IsZero() || t.After(biggestFinishedAt.Time) {
						biggestFinishedAt = *t
					}
				}
			}
			if stageFinished {
				switch stage.Status {
				case v1.ActivityStatusTypeSucceeded, v1.ActivityStatusTypeNotExecuted:
					// stage did not fail
				default:
					failed = true
				}
			} else {
				allCompleted = false
			}
			if stage.Status == v1.ActivityStatusTypeRunning {
				running = true
			}
			if stage.Status == v1.ActivityStatusTypeRunning || stage.Status == v1.ActivityStatusTypePending {
				allCompleted = false
			}
		}
	}

	if !allCompleted && containersTerminated {
		allCompleted = true
	}
	if allCompleted {
		if failed {
			spec.Status = v1.ActivityStatusTypeFailed
		} else {
			spec.Status = v1.ActivityStatusTypeSucceeded
		}
		if !biggestFinishedAt.IsZero() {
			spec.CompletedTimestamp = &biggestFinishedAt
		}

		// log that the build completed
		if onCompleteCallback != nil {
			onCompleteCallback(activity)
		}

	} else {
		if running {
			spec.Status = v1.ActivityStatusTypeRunning
		} else {
			spec.Status = v1.ActivityStatusTypePending
		}
	}
}

// UpdateStageStatus updates the status of a stage
func UpdateStageStatus(stage *v1.StageActivityStep) {
	allCompleted := true
	failed := false
	running := false
	for _, s := range stage.Steps {
		switch s.Status {
		case   v1.ActivityStatusTypeRunning:
			running = true
			allCompleted = false

		case   v1.ActivityStatusTypePending, v1.ActivityStatusTypeNotExecuted, v1.ActivityStatusTypeNone, v1.ActivityStatusTypeWaitingForApproval:
			allCompleted = false

		case   v1.ActivityStatusTypeAborted, v1.ActivityStatusTypeError, v1.ActivityStatusTypeFailed:
			failed = true
		}
	}

	if failed {
		stage.Status = v1.ActivityStatusTypeFailed
		return
	}
	if running {
		stage.Status = v1.ActivityStatusTypeRunning
		return
	}
	if allCompleted {
		stage.Status = v1.ActivityStatusTypeSucceeded
	}
}

