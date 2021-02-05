// +build unit

package services_test

import (
	"testing"

	"github.com/jenkins-x/jx-helpers/v3/pkg/kube"
	"github.com/jenkins-x/jx-helpers/v3/pkg/kube/services"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestExtractServiceSchemePortDefault(t *testing.T) {
	t.Parallel()
	s := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-spring-boot-demo2",
			Namespace: "default-staging",
			Labels: map[string]string{
				"chart": "preview-0.0.0-SNAPSHOT-PR-29-28",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:     "http",
					Protocol: "TCP",
					Port:     80,
				},
			},
		},
	}
	schema, port, _ := services.ExtractServiceSchemePort(s)
	assert.Equal(t, "http", schema)
	assert.Equal(t, "80", port)
}

func TestExtractServiceSchemePortHttps(t *testing.T) {
	t.Parallel()
	s := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-spring-boot-demo2",
			Namespace: "default-staging",
			Labels: map[string]string{
				"chart": "preview-0.0.0-SNAPSHOT-PR-29-28",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:     "https",
					Protocol: "TCP",
					Port:     443,
				},
			},
		},
	}
	schema, port, _ := services.ExtractServiceSchemePort(s)
	assert.Equal(t, "https", schema)
	assert.Equal(t, "443", port)
}

func TestExtractServiceSchemePortHttpsFirst(t *testing.T) {
	t.Parallel()
	s := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-spring-boot-demo2",
			Namespace: "default-staging",
			Labels: map[string]string{
				"chart": "preview-0.0.0-SNAPSHOT-PR-29-28",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:     "http",
					Protocol: "TCP",
					Port:     80,
				},
				{
					Name:     "https",
					Protocol: "TCP",
					Port:     443,
				},
			},
		},
	}
	schema, port, _ := services.ExtractServiceSchemePort(s)
	assert.Equal(t, "https", schema)
	assert.Equal(t, "443", port)
}

func TestExtractServiceSchemePortHttpsOdd(t *testing.T) {
	t.Parallel()
	s := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-spring-boot-demo2",
			Namespace: "default-staging",
			Labels: map[string]string{
				"chart": "preview-0.0.0-SNAPSHOT-PR-29-28",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:     "brian",
					Protocol: "TCP",
					Port:     443,
				},
			},
		},
	}
	schema, port, _ := services.ExtractServiceSchemePort(s)
	assert.Equal(t, "https", schema)
	assert.Equal(t, "443", port)
}

func TestExtractServiceSchemePortHttpsNamed(t *testing.T) {
	t.Parallel()
	s := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-spring-boot-demo2",
			Namespace: "default-staging",
			Labels: map[string]string{
				"chart": "preview-0.0.0-SNAPSHOT-PR-29-28",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:     "dave",
					Protocol: "UDP",
					Port:     800,
				},
				{
					Name:     "brian",
					Protocol: "TCP",
					Port:     444,
				},
				{
					Name:     "https",
					Protocol: "TCP",
					Port:     443,
				},
			},
		},
	}
	schema, port, _ := services.ExtractServiceSchemePort(s)
	assert.Equal(t, "https", schema)
	assert.Equal(t, "443", port)
}

func TestExtractServiceSchemePortHttpNamed(t *testing.T) {
	t.Parallel()
	s := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-spring-boot-demo2",
			Namespace: "default-staging",
			Labels: map[string]string{
				"chart": "preview-0.0.0-SNAPSHOT-PR-29-28",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:     "dave",
					Protocol: "UDP",
					Port:     800,
				},
				{
					Name:     "brian",
					Protocol: "TCP",
					Port:     444,
				},
				{
					Name:     "http",
					Protocol: "TCP",
					Port:     8083,
				},
			},
		},
	}
	schema, port, _ := services.ExtractServiceSchemePort(s)
	assert.Equal(t, "http", schema)
	assert.Equal(t, "8083", port)
}

func TestExtractServiceSchemePortHttpNotNamed(t *testing.T) {
	t.Parallel()
	s := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-spring-boot-demo2",
			Namespace: "default-staging",
			Labels: map[string]string{
				"chart": "preview-0.0.0-SNAPSHOT-PR-29-28",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:     "http",
					Protocol: "TCP",
					Port:     8088,
				},
				{
					Name:     "alan",
					Protocol: "TCP",
					Port:     80,
				},
			},
		},
	}
	schema, port, _ := services.ExtractServiceSchemePort(s)
	assert.Equal(t, "http", schema)
	assert.Equal(t, "80", port)
}

func TestExtractServiceSchemePortNamedPrefHttps(t *testing.T) {
	t.Parallel()
	s := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-spring-boot-demo2",
			Namespace: "default-staging",
			Labels: map[string]string{
				"chart": "preview-0.0.0-SNAPSHOT-PR-29-28",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:     "ssh",
					Protocol: "UDP",
					Port:     22,
				},
				{
					Name:     "hiddenhttp",
					Protocol: "TCP",
					Port:     8083,
				},
				{
					Name:     "sctp-tunneling",
					Protocol: "TCP",
					Port:     9899,
				},
				{
					Name:     "https",
					Protocol: "TCP",
					Port:     8443,
				},
			},
		},
	}
	schema, port, _ := services.ExtractServiceSchemePort(s)
	assert.Equal(t, "https", schema)
	assert.Equal(t, "8443", port)
}

func TestExtractServiceSchemePortInconclusive(t *testing.T) {
	t.Parallel()
	s := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-spring-boot-demo2",
			Namespace: "default-staging",
			Labels: map[string]string{
				"chart": "preview-0.0.0-SNAPSHOT-PR-29-28",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:     "ssh",
					Protocol: "UDP",
					Port:     22,
				},
				{
					Name:     "hiddenhttp",
					Protocol: "TCP",
					Port:     8083,
				},
				{
					Name:     "sctp-tunneling",
					Protocol: "TCP",
					Port:     9899,
				},
			},
		},
	}
	schema, port, _ := services.ExtractServiceSchemePort(s)
	assert.Equal(t, "", schema)
	assert.Equal(t, "", port)
}

func TestFindURLFromIngress(t *testing.T) {
	t.Parallel()

	type testData struct {
		Name        string
		ExpectedURL string
		Ingress     *v1beta1.Ingress
	}

	testCases := []testData{
		{
			Name:        "http-LoadBalancer",
			ExpectedURL: "http://hook-jx.1.2.3.4.nip.io",
			Ingress: &v1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name: "hook",
				},
				Spec: v1beta1.IngressSpec{
					Rules: []v1beta1.IngressRule{
						{
							Host: "hook-jx.1.2.3.4.nip.io",
							IngressRuleValue: v1beta1.IngressRuleValue{
								HTTP: &v1beta1.HTTPIngressRuleValue{
									Paths: []v1beta1.HTTPIngressPath{
										{
											Path: "",
											Backend: v1beta1.IngressBackend{
												ServiceName: "hook",
												ServicePort: intstr.IntOrString{
													IntVal: 80,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			Name:        "https-LoadBalancer",
			ExpectedURL: "https://hook-jx.1.2.3.4.nip.io",
			Ingress: &v1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name: "hook",
				},
				Spec: v1beta1.IngressSpec{
					Rules: []v1beta1.IngressRule{
						{
							Host: "hook-jx.1.2.3.4.nip.io",
							IngressRuleValue: v1beta1.IngressRuleValue{
								HTTP: &v1beta1.HTTPIngressRuleValue{
									Paths: []v1beta1.HTTPIngressPath{
										{
											Path: "",
											Backend: v1beta1.IngressBackend{
												ServiceName: "hook",
												ServicePort: intstr.IntOrString{
													IntVal: 80,
												},
											},
										},
									},
								},
							},
						},
					},
					TLS: []v1beta1.IngressTLS{
						{
							Hosts:      []string{"hook-jx.1.2.3.4.nip.io"},
							SecretName: "",
						},
					},
				},
			},
		},
		{
			Name:        "http-NodePort",
			ExpectedURL: "http://1.2.3.4:4567/jx/hook",
			Ingress: &v1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name: "hook",
					Annotations: map[string]string{
						kube.AnnotationHost: "1.2.3.4:4567",
					},
				},
				Spec: v1beta1.IngressSpec{
					Rules: []v1beta1.IngressRule{
						{
							IngressRuleValue: v1beta1.IngressRuleValue{
								HTTP: &v1beta1.HTTPIngressRuleValue{
									Paths: []v1beta1.HTTPIngressPath{
										{
											Path: "/jx/hook",
											Backend: v1beta1.IngressBackend{
												ServiceName: "hook",
												ServicePort: intstr.IntOrString{
													IntVal: 80,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		actual := services.IngressURL(tc.Ingress)
		assert.Equal(t, tc.ExpectedURL, actual, "IngressURL for %s", tc.Name)
		t.Logf("test %s generated URL from Ingress %s", tc.Name, actual)
	}
}
