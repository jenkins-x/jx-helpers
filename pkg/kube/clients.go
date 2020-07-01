package kube

import (
	"github.com/jenkins-x/jx-kube-client/pkg/kubeclient"
	"github.com/pkg/errors"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
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
