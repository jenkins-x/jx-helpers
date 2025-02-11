package jxclient

import (
	"fmt"
	"os"

	"github.com/jenkins-x/jx-api/v4/pkg/client/clientset/versioned"
	fakejx "github.com/jenkins-x/jx-api/v4/pkg/client/clientset/versioned/fake"
	"github.com/jenkins-x/jx-helpers/v3/pkg/kube"
	"github.com/jenkins-x/jx-helpers/v3/pkg/kube/jxenv"
	"github.com/jenkins-x/jx-kube-client/v3/pkg/kubeclient"
)

// LazyCreateJXClient lazy creates the jx client if its not defined
func LazyCreateJXClient(client versioned.Interface) (versioned.Interface, error) {
	if client != nil {
		return client, nil
	}
	if kube.IsNoKubernetes() {
		return noKubernetesFakeJXClient()
	}
	f := kubeclient.NewFactory()
	cfg, err := f.CreateKubeConfig()
	if err != nil {
		return client, fmt.Errorf("failed to get kubernetes config: %w", err)
	}
	client, err = versioned.NewForConfig(cfg)
	if err != nil {
		return client, fmt.Errorf("error building jx clientset: %w", err)
	}
	return client, nil
}

// LazyCreateJXClientAndNamespace lazy creates the jx client and/or the current namespace if not already defined
func LazyCreateJXClientAndNamespace(client versioned.Interface, ns string) (versioned.Interface, string, error) {
	if client != nil && ns != "" {
		return client, ns, nil
	}
	if kube.IsNoKubernetes() {
		if client == nil {
			var err error
			client, err = noKubernetesFakeJXClient()
			if err != nil {
				return client, ns, fmt.Errorf("failed to : %w", err)
			}
		}
		if ns == "" {
			ns = "default"
		}
		return client, ns, nil
	}
	if client == nil {
		f := kubeclient.NewFactory()
		cfg, err := f.CreateKubeConfig()
		if err != nil {
			return client, ns, fmt.Errorf("failed to get kubernetes config: %w", err)
		}
		client, err = versioned.NewForConfig(cfg)
		if err != nil {
			return client, ns, fmt.Errorf("error building jx clientset: %w", err)
		}
	}
	if ns == "" {
		var err error
		ns, err = kubeclient.CurrentNamespace()
		if err != nil {
			return client, ns, fmt.Errorf("failed to get current kubernetes namespace: %w", err)
		}
	}
	return client, ns, nil
}

func noKubernetesFakeJXClient() (versioned.Interface, error) {
	ns := "jx"
	gitURL := os.Getenv("JX_ENVIRONMENT_GIT_URL")
	if gitURL == "" {
		gitURL = "https://github.com/jx3-gitops-repositories/jx3-github.git"
	}
	devEnv := jxenv.CreateDefaultDevEnvironment(ns)
	devEnv.Namespace = ns
	devEnv.Spec.Source.URL = gitURL

	defaultNS := "default"
	devEnvDefault := jxenv.CreateDefaultDevEnvironment(defaultNS)
	devEnvDefault.Namespace = defaultNS
	devEnvDefault.Spec.Source.URL = gitURL
	return fakejx.NewSimpleClientset(devEnv, devEnvDefault), nil
}
