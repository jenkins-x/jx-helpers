// +build unit

package pods_test

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jenkins-x/jx-helpers/v3/pkg/kube/pods"
	"github.com/jenkins-x/jx-helpers/v3/pkg/yamls"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetPodConditionPodReady(t *testing.T) {
	t.Parallel()

	var condition v1.PodConditionType
	condition = v1.PodReady

	status := v1.PodStatus{
		Phase: v1.PodRunning,
		Conditions: []v1.PodCondition{
			{
				Type:   condition,
				Status: v1.ConditionTrue,
			},
		},
	}

	resStatus, res := pods.GetPodCondition(&status, condition)

	assert.Equal(t, 0, resStatus)
	assert.Equal(t, condition, res.Type)
}

func TestPodFailed(t *testing.T) {
	t.Parallel()

	sourceData := filepath.Join("test_data", "pod_failed")
	fileNames, err := ioutil.ReadDir(sourceData)
	assert.NoError(t, err)

	for _, f := range fileNames {
		name := f.Name()
		if f.IsDir() || !strings.HasSuffix(name, ".yaml") {
			continue
		}
		fileName := filepath.Join(sourceData, name)
		pod := &v1.Pod{}
		err := yamls.LoadFile(fileName, pod)
		require.NoError(t, err, "failed to load file %s", fileName)

		assert.Equal(t, true, pods.IsPodFailed(pod), "IsPodFailed for %s", name)
		assert.Equal(t, false, pods.IsPodReady(pod), "IsPodReady for %s", name)

		t.Logf("file %s has failed pod\n", name)
	}
}

func TestPodNotFailed(t *testing.T) {
	t.Parallel()

	sourceData := filepath.Join("test_data", "pod_ready")

	testCases := []struct {
		file     string
		validate func(string, *v1.Pod)
	}{
		{
			file: "pod_ready.yaml",
			validate: func(name string, pod *v1.Pod) {
				assert.Equal(t, true, pods.IsPodReady(pod), "IsPodReady for %s", name)
				assert.Equal(t, false, pods.IsPodPending(pod), "IsPodPending for %s", name)
				assert.Equal(t, false, pods.IsPodFailed(pod), "IsPodFailed for %s", name)
			},
		},
		{
			file: "pod_pending.yaml",
			validate: func(name string, pod *v1.Pod) {
				assert.Equal(t, false, pods.IsPodReady(pod), "IsPodReady for %s", name)
				assert.Equal(t, true, pods.IsPodPending(pod), "IsPodPending for %s", name)
				assert.Equal(t, false, pods.IsPodFailed(pod), "IsPodFailed for %s", name)

				t.Logf("pod %s ready %v\n", name, pods.IsPodReady(pod))
				t.Logf("pod %s pending %v\n", name, pods.IsPodPending(pod))
			},
		},
	}

	for _, tc := range testCases {
		name := tc.file
		fileName := filepath.Join(sourceData, name)
		pod := &v1.Pod{}
		err := yamls.LoadFile(fileName, pod)
		require.NoError(t, err, "failed to load file %s", fileName)

		assert.Equal(t, false, pods.IsPodFailed(pod), "IsPodFailed for %s", name)

		tc.validate(name, pod)
		t.Logf("pod from %s processed\n", name)
	}
}

func TestGetPodConditionFailures(t *testing.T) {
	t.Parallel()

	var condition v1.PodConditionType
	condition = v1.PodReady

	status := v1.PodStatus{
		Phase: v1.PodRunning,
		Conditions: []v1.PodCondition{
			{
				Status: v1.ConditionTrue,
			},
		},
	}

	resStatus, _ := pods.GetPodCondition(nil, condition)
	assert.Equal(t, -1, resStatus)

	// Status missing type fails
	resStatus, _ = pods.GetPodCondition(&status, condition)
	assert.Equal(t, -1, resStatus)
}

func TestGetPodReadyCondition(t *testing.T) {
	t.Parallel()

	status := v1.PodStatus{
		Phase: v1.PodRunning,
		Conditions: []v1.PodCondition{
			{
				Type:   v1.PodReady,
				Status: v1.ConditionTrue,
			},
		},
	}

	res := pods.GetPodReadyCondition(status)
	assert.Equal(t, status.Conditions[0].Status, res.Status)
	assert.Equal(t, status.Conditions[0].Type, res.Type)
}

func TestGetPodReadyConditionFailures(t *testing.T) {
	t.Parallel()

	status := v1.PodStatus{
		Phase: v1.PodRunning,
		Conditions: []v1.PodCondition{
			{
				Status: v1.ConditionTrue,
			},
		},
	}

	var expectedCondition *v1.PodCondition
	res := pods.GetPodReadyCondition(status)
	assert.Equal(t, expectedCondition, res)
}

func TestIsPodReadyConditionTrue(t *testing.T) {
	t.Parallel()

	status := v1.PodStatus{
		Phase: v1.PodRunning,
		Conditions: []v1.PodCondition{
			{
				Type:   v1.PodReady,
				Status: v1.ConditionTrue,
			},
		},
	}

	res := pods.IsPodReadyConditionTrue(status)
	assert.Equal(t, true, res)
}

func TestIsPodReadyConditionTrueFailures(t *testing.T) {
	t.Parallel()

	status := v1.PodStatus{
		Phase: v1.PodRunning,
		Conditions: []v1.PodCondition{
			{
				Status: v1.ConditionTrue,
			},
		},
	}

	res := pods.IsPodReadyConditionTrue(status)
	assert.Equal(t, false, res)

	status = v1.PodStatus{
		Phase: v1.PodRunning,
		Conditions: []v1.PodCondition{
			{
				Type: v1.PodReady,
			},
		},
	}

	res = pods.IsPodReadyConditionTrue(status)
	assert.Equal(t, false, res)
}

func TestIsPodReady(t *testing.T) {
	t.Parallel()

	labels := make(map[string]string)
	labels["app"] = "ians-app"

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "web",
			Labels:    labels,
			Namespace: "jx-testing",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "web",
					Image: "nginx:1.12",
					Ports: []v1.ContainerPort{
						{
							Name:          "http",
							Protocol:      v1.ProtocolTCP,
							ContainerPort: 80,
						},
					},
				},
			},
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
			Conditions: []v1.PodCondition{
				{
					Type:   v1.PodReady,
					Status: v1.ConditionTrue,
				},
			},
		},
	}

	res := pods.IsPodReady(pod)
	assert.Equal(t, true, res)
}

func TestIsPodReadyFailures(t *testing.T) {
	t.Parallel()

	labels := make(map[string]string)
	labels["app"] = "ians-app"

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "web",
			Labels:    labels,
			Namespace: "jx-testing",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "web",
					Image: "nginx:1.12",
					Ports: []v1.ContainerPort{
						{
							Name:          "http",
							Protocol:      v1.ProtocolTCP,
							ContainerPort: 80,
						},
					},
				},
			},
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
			Conditions: []v1.PodCondition{
				{
					Type:   "Something else",
					Status: v1.ConditionTrue,
				},
			},
		},
	}

	res := pods.IsPodReady(pod)
	assert.Equal(t, false, res)

	pod = &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "web",
			Labels:    labels,
			Namespace: "jx-testing",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "web",
					Image: "nginx:1.12",
					Ports: []v1.ContainerPort{
						{
							Name:          "http",
							Protocol:      v1.ProtocolTCP,
							ContainerPort: 80,
						},
					},
				},
			},
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
			Conditions: []v1.PodCondition{
				{
					Type:   v1.PodReady,
					Status: "Something else",
				},
			},
		},
	}

	res = pods.IsPodReady(pod)
	assert.Equal(t, false, res)
}
