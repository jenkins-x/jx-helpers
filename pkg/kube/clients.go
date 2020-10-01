package kube

import (
	"github.com/jenkins-x/jx-kube-client/v3/pkg/kubeclient"
	"github.com/pkg/errors"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// LazyCreateDynamicClient lazily creates the dynamic client if its not defined
func LazyCreateDynamicClient(client dynamic.Interface) (dynamic.Interface, error) {
	if client != nil {
		return client, nil
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
	f := kubeclient.NewFactory()
	if client == nil {
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
