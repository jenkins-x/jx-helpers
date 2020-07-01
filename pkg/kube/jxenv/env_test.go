// +build unit

package jxenv_test

import (
	"testing"

	jenkinsio_v1 "github.com/jenkins-x/jx-api/pkg/apis/jenkins.io/v1"
	"github.com/jenkins-x/jx-helpers/pkg/kube"
	"github.com/jenkins-x/jx-helpers/pkg/kube/jxenv"
	"github.com/stretchr/testify/assert"
	"gopkg.in/AlecAivazis/survey.v1/core"
	k8sv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_mocks "k8s.io/client-go/kubernetes/fake"
)

func init() {
	// disable color output for all prompts to simplify testing
	core.DisableColor = true
}

func TestSortEnvironments(t *testing.T) {
	t.Parallel()
	environments := []jenkinsio_v1.Environment{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "c",
			},
			Spec: jenkinsio_v1.EnvironmentSpec{
				Order: 100,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "z",
			},
			Spec: jenkinsio_v1.EnvironmentSpec{
				Order: 5,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "d",
			},
			Spec: jenkinsio_v1.EnvironmentSpec{
				Order: 100,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "a",
			},
			Spec: jenkinsio_v1.EnvironmentSpec{
				Order: 150,
			},
		},
	}

	jxenv.SortEnvironments(environments)

	assert.Equal(t, "z", environments[0].Name, "Environment 0")
	assert.Equal(t, "c", environments[1].Name, "Environment 1")
	assert.Equal(t, "d", environments[2].Name, "Environment 2")
	assert.Equal(t, "a", environments[3].Name, "Environment 3")
}

func TestSortEnvironments2(t *testing.T) {
	t.Parallel()
	environments := []jenkinsio_v1.Environment{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "dev",
			},
			Spec: jenkinsio_v1.EnvironmentSpec{
				Order: 0,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "prod",
			},
			Spec: jenkinsio_v1.EnvironmentSpec{
				Order: 200,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "staging",
			},
			Spec: jenkinsio_v1.EnvironmentSpec{
				Order: 100,
			},
		},
	}

	jxenv.SortEnvironments(environments)

	assert.Equal(t, "dev", environments[0].Name, "Environment 0")
	assert.Equal(t, "staging", environments[1].Name, "Environment 1")
	assert.Equal(t, "prod", environments[2].Name, "Environment 2")
}

func TestGetDevNamespace(t *testing.T) {
	namespace := &k8sv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "jx-testing",
			Namespace: "jx-testing",
		},
	}
	kubernetesInterface := kube_mocks.NewSimpleClientset(namespace)
	testNS := "jx-testing"
	testEnv := ""

	ns, env, err := jxenv.GetDevNamespace(kubernetesInterface, testNS)

	assert.NoError(t, err, "Should not error")
	assert.Equal(t, testNS, ns)
	assert.Equal(t, testEnv, env)
}

func TestGetPreviewEnvironmentReleaseName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		env                 *jenkinsio_v1.Environment
		expectedReleaseName string
	}{
		{
			env:                 nil,
			expectedReleaseName: "",
		},
		{
			env:                 &jenkinsio_v1.Environment{},
			expectedReleaseName: "",
		},
		{
			env:                 jxenv.NewPreviewEnvironment("test"),
			expectedReleaseName: "",
		},
		{
			env: func() *jenkinsio_v1.Environment {
				env := jxenv.NewPreviewEnvironment("test")
				if env.Annotations == nil {
					env.Annotations = map[string]string{}
				}
				env.Annotations[kube.AnnotationReleaseName] = "release-name"
				return env
			}(),
			expectedReleaseName: "release-name",
		},
	}

	for i, test := range tests {
		releaseName := jxenv.GetPreviewEnvironmentReleaseName(test.env)
		if releaseName != test.expectedReleaseName {
			t.Errorf("[%d] Expected release name %s but got %s", i, test.expectedReleaseName, releaseName)
		}
	}
}
