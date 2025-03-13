package jobs

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

// IsJobSucceeded returns true if the job completed
func IsJobSucceeded(job *batchv1.Job) bool {
	for _, con := range job.Status.Conditions {
		if con.Type == batchv1.JobComplete && con.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

// IsJobFinished returns true if the job has finished, i.e. is completed, failed or suspended
// Technically a suspended job can be resumed, but the logic in jx assume this doesn't happen
// Reference: https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#jobstatus-v1-batch
func IsJobFinished(job *batchv1.Job) bool {
	for _, con := range job.Status.Conditions {
		if (con.Type == batchv1.JobComplete || con.Type == batchv1.JobFailed || con.Type == batchv1.JobSuspended) &&
			con.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}
