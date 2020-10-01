// +build unit

package pods_test

import (
	"testing"

	"github.com/jenkins-x/jx-helpers/v3/pkg/kube/pods"
	"github.com/stretchr/testify/assert"
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
