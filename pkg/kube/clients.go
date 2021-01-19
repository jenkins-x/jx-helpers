package kube

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"strings"

	"github.com/jenkins-x/jx-kube-client/v3/pkg/kubeclient"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	fakedyn "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

// LazyCreateDynamicClient lazily creates the dynamic client if its not defined
func LazyCreateDynamicClient(client dynamic.Interface) (dynamic.Interface, error) {
	if client != nil {
		return client, nil
	}
	if IsNoKubernetes() {
		scheme := runtime.NewScheme()
		return fakedyn.NewSimpleDynamicClient(scheme), nil
	}
	f := kubeclient.NewFactory()
	cfg, err := f.CreateKubeConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get kubernetes config")
	}
	client, err = dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "error building dynamic clientset")
	}
	return client, nil
}

// LazyCreateKubeClient lazy creates the kube client if its not defined
func LazyCreateKubeClient(client kubernetes.Interface) (kubernetes.Interface, error) {
	if client != nil {
		return client, nil
	}
	if IsNoKubernetes() {
		return NewFakeKubernetesClient("default"), nil
	}
	f := kubeclient.NewFactory()
	cfg, err := f.CreateKubeConfig()
	if err != nil {
		return client, errors.Wrap(err, "failed to get kubernetes config")
	}
	client, err = kubernetes.NewForConfig(cfg)
	if err != nil {
		return client, errors.Wrap(err, "error building kubernetes clientset")
	}
	return client, nil
}

// LazyCreateKubeClientAndNamespace lazy creates the kube client and/or the current namespace if not already defined
func LazyCreateKubeClientAndNamespace(client kubernetes.Interface, ns string) (kubernetes.Interface, string, error) {
	if client != nil && ns != "" {
		return client, ns, nil
	}
	if IsNoKubernetes() {
		if ns == "" {
			ns = "default"
		}
		if client == nil {
			client = NewFakeKubernetesClient(ns)
		}
		return client, ns, nil
	}
	if client == nil {
		f := kubeclient.NewFactory()
		cfg, err := f.CreateKubeConfig()
		if err != nil {
			return client, ns, errors.Wrap(err, "failed to get kubernetes config")
		}
		client, err = kubernetes.NewForConfig(cfg)
		if err != nil {
			return client, ns, errors.Wrap(err, "error building kubernetes clientset")
		}
	}
	if ns == "" {
		var err error
		ns, err = kubeclient.CurrentNamespace()
		if err != nil {
			return client, ns, errors.Wrap(err, "failed to get current kubernetes namespace")
		}
	}
	return client, ns, nil
}

// IsInCluster tells if we are running incluster
func IsInCluster() bool {
	_, err := rest.InClusterConfig()
	return err == nil
}

// IsNoKubernetes returns true if we are inside a GitHub Action or not using kubernetes
func IsNoKubernetes() bool {
	// disable k8s by default if inside a github action
	if strings.ToLower(os.Getenv("GITHUB_ACTIONS")) == "true" {
		return strings.ToLower(os.Getenv("JX_KUBERNETES")) != "true"
	}
	return strings.ToLower(os.Getenv("JX_NO_KUBERNETES")) == "true"
}

// NewFakeKubernetesClient creates a fake k8s client if we have disabled kubernetes
func NewFakeKubernetesClient(ns string) *fake.Clientset {
	return fake.NewSimpleClientset(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: ns,
		},
	})
}
