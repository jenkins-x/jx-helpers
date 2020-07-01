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
		return client, errors.Wrap(err, "error building kubernetes clientset")
	}
	return client, nil
}
