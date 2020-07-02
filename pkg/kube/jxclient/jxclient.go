package jxclient

import (
	"github.com/jenkins-x/jx-api/pkg/client/clientset/versioned"
	"github.com/jenkins-x/jx-kube-client/pkg/kubeclient"
	"github.com/pkg/errors"
)

// LazyCreateJXClient lazy creates the jx client if its not defined
func LazyCreateJXClient(client versioned.Interface) (versioned.Interface, error) {
	if client != nil {
		return client, nil
	}
	f := kubeclient.NewFactory()
	cfg, err := f.CreateKubeConfig()
	if err != nil {
		return client, errors.Wrap(err, "failed to get kubernetes config")
	}
	client, err = versioned.NewForConfig(cfg)
	if err != nil {
		return client, errors.Wrap(err, "error building jx clientset")
	}
	return client, nil
}

// LazyCreateJXClientAndNamespace lazy creates the jx client and/or the current namespace if not already defined
func LazyCreateJXClientAndNamespace(client versioned.Interface, ns string) (versioned.Interface, string, error) {
	if client != nil && ns != "" {
		return client, ns, nil
	}
	f := kubeclient.NewFactory()
	cfg, err := f.CreateKubeConfig()
	if err != nil {
		return client, ns, errors.Wrap(err, "failed to get kubernetes config")
	}
	client, err = versioned.NewForConfig(cfg)
	if err != nil {
		return client, ns, errors.Wrap(err, "error building jx clientset")
	}
	if ns == "" {
		ns, err = kubeclient.CurrentNamespace()
		if err != nil {
			return client, ns, errors.Wrap(err, "failed to get current kubernetes namespace")
		}
	}
	return client, ns, nil
}
