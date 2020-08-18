package jobs

import (
	batchv1 "k8s.io/api/batch/v1"

)
// IsJobSucceeded returns true if the job completed and did not fail
func IsJobSucceeded(job *batchv1.Job) bool {
	return IsJobFinished(job) && job.Status.Succeeded > 0
}

// IsJobFinished returns true if the job has completed
func IsJobFinished(job *batchv1.Job) bool {
	BackoffLimit := job.Spec.BackoffLimit
	return job.Status.CompletionTime != nil || (job.Status.Active == 0 && BackoffLimit != nil && job.Status.Failed >= *BackoffLimit)
}

