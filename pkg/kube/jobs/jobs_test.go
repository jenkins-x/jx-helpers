package jobs_test

import (
	"path/filepath"
	"testing"

	"github.com/jenkins-x/jx-helpers/v3/pkg/kube/jobs"
	"github.com/jenkins-x/jx-helpers/v3/pkg/yamls"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
)

func TestJobSucceeded(t *testing.T) {
	t.Parallel()

	fileName := filepath.Join("test_data", "succeeded.yaml")
	job := &batchv1.Job{}
	err := yamls.LoadFile(fileName, job)
	require.NoError(t, err, "failed to load file %s", fileName)

	assert.Equal(t, true, jobs.IsJobFinished(job), "IsFinished")
	assert.Equal(t, true, jobs.IsJobSucceeded(job), "IsJobSucceeded")
}

func TestJobFailed(t *testing.T) {
	t.Parallel()

	fileName := filepath.Join("test_data", "failed.yaml")
	job := &batchv1.Job{}
	err := yamls.LoadFile(fileName, job)
	require.NoError(t, err, "failed to load file %s", fileName)

	assert.Equal(t, true, jobs.IsJobFinished(job), "IsFinished")
	assert.Equal(t, false, jobs.IsJobSucceeded(job), "IsJobSucceeded")
}

func TestJobRunning(t *testing.T) {
	t.Parallel()

	fileName := filepath.Join("test_data", "running.yaml")
	job := &batchv1.Job{}
	err := yamls.LoadFile(fileName, job)
	require.NoError(t, err, "failed to load file %s", fileName)

	assert.Equal(t, false, jobs.IsJobFinished(job), "IsFinished")
	assert.Equal(t, false, jobs.IsJobSucceeded(job), "IsJobSucceeded")
}
