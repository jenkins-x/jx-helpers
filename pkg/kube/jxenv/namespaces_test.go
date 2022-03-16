//go:build unit
// +build unit

package jxenv_test

import (
	"testing"

	jenkinsio_v1 "github.com/jenkins-x/jx-api/v4/pkg/apis/jenkins.io/v1"
	"github.com/jenkins-x/jx-api/v4/pkg/client/clientset/versioned/fake"
	"github.com/jenkins-x/jx-helpers/v3/pkg/kube"
	"github.com/jenkins-x/jx-helpers/v3/pkg/kube/jxenv"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestEnsureDevEnvironmentSetup(t *testing.T) {
	t.Parallel()

	versiondInterface := fake.NewSimpleClientset()

	envFixture := &jenkinsio_v1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name: kube.LabelValueDevEnvironment,
		},
		Spec: jenkinsio_v1.EnvironmentSpec{
			Namespace:         "jx-testing",
			Label:             "Development",
			PromotionStrategy: jenkinsio_v1.PromotionStrategyTypeNever,
			Kind:              jenkinsio_v1.EnvironmentKindTypeDevelopment,
			TeamSettings: jenkinsio_v1.TeamSettings{
				AppsRepository: kube.DefaultChartMuseumURL,
			},
		},
	}

	env, err := jxenv.EnsureDevEnvironmentSetup(versiondInterface, "jx-testing")

	assert.NoError(t, err, "Should not error")
	assert.Equal(t, envFixture.ObjectMeta.Name, env.ObjectMeta.Name)
	assert.Equal(t, envFixture.Spec.Namespace, env.Spec.Namespace)
	assert.Equal(t, envFixture.Spec.Label, env.Spec.Label)
	assert.Equal(t, jenkinsio_v1.PromotionStrategyType("Never"), env.Spec.PromotionStrategy)
	assert.Equal(t, jenkinsio_v1.EnvironmentKindType("Development"), env.Spec.Kind)
	assert.Equal(t, jenkinsio_v1.PromotionEngineType("Prow"), env.Spec.TeamSettings.PromotionEngine)
	assert.Equal(t, envFixture.Spec.TeamSettings.AppsRepository, env.Spec.TeamSettings.AppsRepository)
}
