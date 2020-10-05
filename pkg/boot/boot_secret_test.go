package boot_test

import (
	"testing"

	"github.com/jenkins-x/jx-helpers/v3/pkg/boot"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestBootSecret(t *testing.T) {
	expectedURL := "https://github.com/myorg/myrepo.git"
	expectedUsername := "myuser"
	expectedPassword := "mypwd"

	for _, ns := range []string{boot.SecretName, boot.GitOperatorNamespace, "random"} {
		secretNS := ns
		if secretNS == "random" {
			secretNS = boot.GitOperatorNamespace
		}
		kubeClient := fake.NewSimpleClientset(
			&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      boot.SecretName,
					Namespace: secretNS,
				},
				Data: map[string][]byte{
					"url":      []byte(expectedURL),
					"username": []byte(expectedUsername),
					"password": []byte(expectedPassword),
				},
			},
		)

		bootSecret, err := boot.LoadBootSecret(kubeClient, ns, boot.GitOperatorNamespace, boot.SecretName, "")
		require.NoError(t, err, "failed to load boot secret in ns %s with secret namespace %s", ns, secretNS)
		require.NotNil(t, bootSecret, "no BootSecret in ns %s with secret namespace %s", ns, secretNS)

		assert.Equal(t, expectedURL, bootSecret.URL, "bootSecret.URL for ns %s with secret namespace %s", ns, secretNS)
		assert.Equal(t, expectedUsername, bootSecret.Username, "bootSecret.Username for ns %s with secret namespace %s", ns, secretNS)
		assert.Equal(t, expectedPassword, bootSecret.Password, "bootSecret.Password for ns %s with secret namespace %s", ns, secretNS)

		t.Logf("loaded boot secret when secret namespace is %s: URL: %s user: %s pwd: %s\n", ns, bootSecret.URL, bootSecret.Username, bootSecret.Password)
	}
}
