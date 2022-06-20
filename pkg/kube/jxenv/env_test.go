//go:build unit
// +build unit

package jxenv_test

import (
	"testing"

	"github.com/AlecAivazis/survey/v2/core"
	jenkinsio_v1 "github.com/jenkins-x/jx-api/v4/pkg/apis/jenkins.io/v1"
	v1 "github.com/jenkins-x/jx-api/v4/pkg/apis/jenkins.io/v1"
	"github.com/jenkins-x/jx-helpers/v3/pkg/kube"
	"github.com/jenkins-x/jx-helpers/v3/pkg/kube/jxenv"
	"github.com/stretchr/testify/assert"
	k8sv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
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
			Name: "jx-testing",
		},
	}
	kubernetesInterface := fake.NewSimpleClientset(namespace)
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

func TestGetRepositoryGitURL(t *testing.T) {

	tests := []struct {
		name    string
		args    *v1.SourceRepository
		want    string
		wantErr bool
	}{
		{
			name: "simple", args: &v1.SourceRepository{
				Spec: v1.SourceRepositorySpec{
					Org:      "foo",
					Repo:     "bar",
					Provider: "github.com",
				},
			}, want: "github.com/foo/bar.git", wantErr: false},
		{
			name: "bb-simple", args: &v1.SourceRepository{
				Spec: v1.SourceRepositorySpec{
					Org:          "foo",
					Repo:         "bar",
					Provider:     "bitbucketserver.com",
					ProviderKind: "bitbucketserver",
				},
			}, want: "bitbucketserver.com/scm/foo/bar.git", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jxenv.GetRepositoryGitURL(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRepositoryGitURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetRepositoryGitURL() got = %v, want %v", got, tt.want)
			}
		})
	}
}
