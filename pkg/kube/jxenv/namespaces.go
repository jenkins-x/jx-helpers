package jxenv

import (
	"context"
	"fmt"

	v1 "github.com/jenkins-x/jx-api/v4/pkg/apis/jenkins.io/v1"
	"github.com/jenkins-x/jx-api/v4/pkg/client/clientset/versioned"
	"github.com/jenkins-x/jx-helpers/v3/pkg/kube"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// EnsureDevNamespaceCreatedWithoutEnvironment ensures that there is a development namespace created
func EnsureDevNamespaceCreatedWithoutEnvironment(kubeClient kubernetes.Interface, ns string) error {
	// lets annotate the team namespace as being the developer environment
	labels := map[string]string{
		kube.LabelTeam:        ns,
		kube.LabelEnvironment: kube.LabelValueDevEnvironment,
	}
	annotations := map[string]string{}
	// lets check that the current namespace is marked as the dev environment
	err := EnsureNamespaceCreated(kubeClient, ns, labels, annotations)
	return err
}

// EnsureDevEnvironmentSetup ensures that the Environment is created in the given namespace
func EnsureDevEnvironmentSetup(jxClient versioned.Interface, ns string) (*v1.Environment, error) {
	// lets ensure there is a dev Environment setup so that we can easily switch between all the environments
	env, err := jxClient.JenkinsV1().Environments(ns).Get(context.TODO(), kube.LabelValueDevEnvironment, metav1.GetOptions{})
	if err != nil {
		// lets create a dev environment
		env = CreateDefaultDevEnvironment(ns)
		env, err = jxClient.JenkinsV1().Environments(ns).Create(context.TODO(), env, metav1.CreateOptions{})
		if err != nil {
			return nil, err
		}
	}
	return env, nil
}

// CreateDefaultDevEnvironment creates a default development environment
func CreateDefaultDevEnvironment(ns string) *v1.Environment {
	return &v1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:   kube.LabelValueDevEnvironment,
			Labels: map[string]string{kube.LabelTeam: ns, kube.LabelEnvironment: kube.LabelValueDevEnvironment},
		},
		Spec: v1.EnvironmentSpec{
			Namespace:         ns,
			Label:             "Development",
			PromotionStrategy: v1.PromotionStrategyTypeNever,
			Kind:              v1.EnvironmentKindTypeDevelopment,
			WebHookEngine:     v1.WebHookEngineLighthouse,
		},
	}
}

// GetEnrichedDevEnvironment lazily creates the dev namespace if it does not already exist and
// auto-detects the webhook engine if its not specified
func GetEnrichedDevEnvironment(kubeClient kubernetes.Interface, jxClient versioned.Interface, ns string) (*v1.Environment, error) {
	env, err := EnsureDevEnvironmentSetup(jxClient, ns)
	if err != nil {
		return env, err
	}
	if env.Spec.WebHookEngine == "none" {
		env.Spec.WebHookEngine = v1.WebHookEngineLighthouse
	}
	return env, nil
}

// Ensure that the namespace exists for the given name
func EnsureNamespaceCreated(kubeClient kubernetes.Interface, name string, labels map[string]string, annotations map[string]string) error {
	n, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), name, metav1.GetOptions{})
	if err == nil {
		// lets check if we have the labels setup
		if n.Annotations == nil {
			n.Annotations = map[string]string{}
		}
		if n.Labels == nil {
			n.Labels = map[string]string{}
		}
		changed := false
		if labels != nil {
			for k, v := range labels {
				if n.Labels[k] != v {
					n.Labels[k] = v
					changed = true
				}
			}
		}
		if annotations != nil {
			for k, v := range annotations {
				if n.Annotations[k] != v {
					n.Annotations[k] = v
					changed = true
				}
			}
		}
		if changed {
			_, err = kubeClient.CoreV1().Namespaces().Update(context.TODO(), n, metav1.UpdateOptions{})
			if err != nil {
				return fmt.Errorf("Failed to label Namespace %s %s", name, err)
			}
		}
		return nil
	}

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Labels:      labels,
			Annotations: annotations,
		},
	}
	_, err = kubeClient.CoreV1().Namespaces().Create(context.TODO(), namespace, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("Failed to create Namespace %s %s", name, err)
	} else {
		log.Logger().Infof("Namespace %s created ", name)
	}
	return err
}
