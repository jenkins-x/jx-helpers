package boot

import (
	"context"

	corev1 "k8s.io/api/core/v1"

	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// BootSecret loads the boot secret
type BootSecret struct {
	// URL the git URL to poll for git operator
	URL string
	// GitProviderURL the git provider URL such as: https://github.com
	GitProviderURL string
	// Username the git user name to clone git
	Username string
	// Password the git password/token to clone git
	Password string
}

// LoadBootSecret loads the boot secret from the current namespace
func LoadBootSecret(kubeClient kubernetes.Interface, ns, operatorNamespace, secretName, defaultUserName string) (*BootSecret, error) {
	secret, err := getBootSecret(kubeClient, ns, operatorNamespace, secretName)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find boot secret")
	}

	answer := &BootSecret{}
	data := secret.Data
	if data != nil {
		if secret.Annotations != nil {
			answer.GitProviderURL = secret.Annotations["tekton.dev/git-0"]
		}
		answer.URL = string(data["url"])
		if answer.URL == "" {
			log.Logger().Debugf("secret %s in namespace %s does not have a url entry", secretName, ns)
		}
		answer.Username = string(data["username"])
		if answer.Username == "" {
			answer.Username = defaultUserName
		}
		answer.Password = string(data["password"])
	}
	return answer, nil
}

func getBootSecret(kubeClient kubernetes.Interface, ns string, operatorNamespace string, secretName string) (*corev1.Secret, error) {
	secret, err := kubeClient.CoreV1().Secrets(ns).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		// lets try either the namespace: jx-git-operator or jx whichever is different
		if operatorNamespace == ns {
			operatorNamespace = "jx"
		}
		secret, err = kubeClient.CoreV1().Secrets(operatorNamespace).Get(context.TODO(), secretName, metav1.GetOptions{})
	}
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find secret %s in namespace %s or %s", secretName, ns, operatorNamespace)
	}
	return secret, nil
}
