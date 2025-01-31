package jobs

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

// IsJobSucceeded returns true if the job completed and did not fail
func IsJobSucceeded(job *batchv1.Job) bool {
	for _, con := range job.Status.Conditions {
		if con.Type == batchv1.JobComplete && con.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

// IsJobFinished returns true if the job has completed
func IsJobFinished(job *batchv1.Job) bool {
	for _, con := range job.Status.Conditions {
		if con.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}
