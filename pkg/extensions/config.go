package extensions

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const ExtensionsConfigDefaultConfigMap = "jenkins-x-extensions"

func GetOrCreateExtensionsConfig(kubeClient kubernetes.Interface, ns string) (*corev1.ConfigMap, error) {
	extensionsConfig, err := kubeClient.CoreV1().ConfigMaps(ns).Get(context.TODO(), ExtensionsConfigDefaultConfigMap, metav1.GetOptions{})
	if err != nil {
		// ConfigMap doesn't exist, create it
		extensionsConfig, err = kubeClient.CoreV1().ConfigMaps(ns).Create(context.TODO(), &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: ExtensionsConfigDefaultConfigMap,
			},
		}, metav1.CreateOptions{})
		if err != nil {
			return nil, err
		}
	}
	if extensionsConfig.Data == nil {
		extensionsConfig.Data = make(map[string]string)
	}
	return extensionsConfig, nil
}
