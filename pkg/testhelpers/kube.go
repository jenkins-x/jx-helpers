package testhelpers

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// AssertLabel asserts the object has the given label value
func AssertLabel(t *testing.T, label string, expected string, objectMeta metav1.ObjectMeta, kindMessage string) {
	message := ObjectNameMessage(objectMeta, kindMessage)
	labels := objectMeta.Labels
	require.NotNil(t, labels, "no labels for %s", message)
	value := labels[label]
	assert.Equal(t, expected, value, "label %s for %s", label, message)
	t.Logf("%s has label %s=%s", message, label, value)
}

// AssertAnnotation asserts the object has the given annotation value
func AssertAnnotation(t *testing.T, annotation string, expected string, objectMeta metav1.ObjectMeta, kindMessage string) {
	message := ObjectNameMessage(objectMeta, kindMessage)
	ann := objectMeta.Annotations
	require.NotNil(t, ann, "no annotations for %s", message)
	value := ann[annotation]
	assert.Equal(t, expected, value, "annotation %s for %s", annotation, message)
	t.Logf("%s has annotation %s=%s", message, annotation, value)
}

// ObjectNameMessage returns an object name message used in the tests
func ObjectNameMessage(objectMeta metav1.ObjectMeta, kindMessage string) string {
	return fmt.Sprintf("%s for name %s", kindMessage, objectMeta.Name)
}

// RequireConfigMapExists requires that the given configMap exists
func RequireConfigMapExists(t *testing.T, client kubernetes.Interface, ns, name string) (*corev1.ConfigMap, string) {
	message := fmt.Sprintf("ConfigMap %s in namespace %s", name, ns)
	configMap, err := client.CoreV1().ConfigMaps(ns).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil && apierrors.IsNotFound(err) {
		require.Fail(t, "missing configMap", "no ConfigMap %s found in namespace %s", name, ns)
		return configMap, message
	}
	require.NoError(t, err, "failed to find ConfigMap %s in namespace %s", name, ns)
	return configMap, message
}

// RequireSecretExists requires that the given secret exists
func RequireSecretExists(t *testing.T, client kubernetes.Interface, ns, name string) (*corev1.Secret, string) {
	message := fmt.Sprintf("Secret %s in namespace %s", name, ns)
	secret, err := client.CoreV1().Secrets(ns).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil && apierrors.IsNotFound(err) {
		require.Fail(t, "missing secret", "no Secret %s found in namespace %s", name, ns)
		return secret, message
	}
	require.NoError(t, err, "failed to find Secret %s in namespace %s", name, ns)
	return secret, message
}

// AssertSecretEntryEquals asserts the Secret resource has the given value
func AssertSecretEntryEquals(t *testing.T, secret *corev1.Secret, key string, expected string, kindMessage string) string {
	require.NotNil(t, secret, "Secret is nil for %s", kindMessage)
	name := secret.Name
	require.NotEmpty(t, secret.Data, "Data is empty in Secret %s for %s", name, kindMessage)

	value := secret.Data[key]
	require.NotNil(t, value, "Secret %s does not have key %s for %s", name, key, kindMessage)
	stringValue := string(value)
	assert.Equal(t, expected, stringValue, "Secret %s key %s for %s", name, key, kindMessage)
	t.Logf("Secret %s has key %s=%s for %s", name, key, value, kindMessage)
	return stringValue
}

// AssertConfigMapEntryEquals asserts the ConfigMap resource has the given data value
func AssertConfigMapEntryEquals(t *testing.T, resource *corev1.ConfigMap, key string, expected string, kindMessage string) string {
	require.NotNil(t, resource, "ConfigMap is nil for %s", kindMessage)
	name := resource.Name
	require.NotEmpty(t, resource.Data, "Data is empty in ConfigMap %s for %s", name, kindMessage)

	value := resource.Data[key]
	assert.Equal(t, expected, value, "ConfigMap %s key %s for %s", name, key, kindMessage)
	t.Logf("ConfigMap %s has key %s=%s for %s", name, key, value, kindMessage)
	return value
}

// AssertConfigMapData asserts the ConfigMap resource has the given data value
func AssertConfigMapHasEntry(t *testing.T, resource *corev1.ConfigMap, key string, kindMessage string) string {
	require.NotNil(t, resource, "ConfigMap is nil for %s", kindMessage)
	name := resource.Name
	require.NotEmpty(t, resource.Data, "Data is empty in ConfigMap %s for %s", name, kindMessage)

	value := resource.Data[key]
	assert.NotEmpty(t, value, "ConfigMap %s key %s for %s", name, key, kindMessage)
	t.Logf("ConfigMap %s has key %s=%s for %s", name, key, value, kindMessage)
	return value
}
