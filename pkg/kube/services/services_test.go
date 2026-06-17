//go:build unit
// +build unit

package services_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/jenkins-x/jx-helpers/v3/pkg/kube"
	"github.com/jenkins-x/jx-helpers/v3/pkg/kube/services"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	nv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	fakedyn "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"
)

// gatewayGVR is the real resource for a Gateway API Gateway. Tests insert Gateways
// under this GVR explicitly because the fake dynamic client's kind->resource heuristic
// mispluralises "Gateway" to "gatewaies".
var gatewayGVR = schema.GroupVersionResource{
	Group:    "gateway.networking.k8s.io",
	Version:  "v1",
	Resource: "gateways",
}

func TestMain(m *testing.M) {
	// warm up jx-logging before parallel tests fan out to prevent a global init race
	log.Logger()
	os.Exit(m.Run())
}

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

func TestIngressURL(t *testing.T) {
	t.Parallel()

	type testData struct {
		Name        string
		ExpectedURL string
		Ingress     *nv1.Ingress
	}

	testCases := []testData{
		{
			Name:        "http-LoadBalancer",
			ExpectedURL: "http://hook-jx.1.2.3.4.nip.io",
			Ingress: &nv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name: "hook",
				},
				Spec: nv1.IngressSpec{
					Rules: []nv1.IngressRule{
						{
							Host: "hook-jx.1.2.3.4.nip.io",
							IngressRuleValue: nv1.IngressRuleValue{
								HTTP: &nv1.HTTPIngressRuleValue{
									Paths: []nv1.HTTPIngressPath{
										{
											Path: "",
											Backend: nv1.IngressBackend{
												Service: &nv1.IngressServiceBackend{
													Name: "hook",
													Port: nv1.ServiceBackendPort{
														Number: 80,
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
		},
		{
			Name:        "https-LoadBalancer",
			ExpectedURL: "https://hook-jx.1.2.3.4.nip.io",
			Ingress: &nv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name: "hook",
				},
				Spec: nv1.IngressSpec{
					Rules: []nv1.IngressRule{
						{
							Host: "hook-jx.1.2.3.4.nip.io",
							IngressRuleValue: nv1.IngressRuleValue{
								HTTP: &nv1.HTTPIngressRuleValue{
									Paths: []nv1.HTTPIngressPath{
										{
											Path: "",
											Backend: nv1.IngressBackend{
												Service: &nv1.IngressServiceBackend{
													Name: "hook",
													Port: nv1.ServiceBackendPort{
														Number: 80,
													},
												},
											},
										},
									},
								},
							},
						},
					},
					TLS: []nv1.IngressTLS{
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
			Ingress: &nv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name: "hook",
					Annotations: map[string]string{
						kube.AnnotationHost: "1.2.3.4:4567",
					},
				},
				Spec: nv1.IngressSpec{
					Rules: []nv1.IngressRule{
						{
							IngressRuleValue: nv1.IngressRuleValue{
								HTTP: &nv1.HTTPIngressRuleValue{
									Paths: []nv1.HTTPIngressPath{
										{
											Path: "/jx/hook",
											Backend: nv1.IngressBackend{
												Service: &nv1.IngressServiceBackend{
													Name: "hook",
													Port: nv1.ServiceBackendPort{
														Number: 80,
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
		},
	}

	for _, tc := range testCases {
		actual := services.IngressURL(tc.Ingress)
		assert.Equal(t, tc.ExpectedURL, actual, "IngressURL for %s", tc.Name)
		t.Logf("test %s generated URL from Ingress %s", tc.Name, actual)
	}
}

func TestGetURLFromVirtualService(t *testing.T) {
	t.Parallel()
	var namespace = "default"
	var name = "jenkins"
	// case when there is no virtual service
	scheme := runtime.NewScheme()
	dynamicClient := fakedyn.NewSimpleDynamicClient(scheme)
	url, err := services.FindURLFromVSIstio(dynamicClient, namespace, name)
	assert.Equal(t, "", url)
	assert.NoError(t, err)
	// case with a virtual service in the given namespace
	virtualService := &unstructured.Unstructured{}
	virtualService.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "networking.istio.io/v1alpha3",
		"kind":       "VirtualService",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
		"spec": map[string]interface{}{
			"hosts": []interface{}{"testing.jenkins-x.in"},
		},
	})
	dynamicClient = fakedyn.NewSimpleDynamicClient(scheme, virtualService)
	url, err = services.FindURLFromVSIstio(dynamicClient, namespace, name)
	if assert.NoError(t, err) {
		assert.Equal(t, "http://testing.jenkins-x.in", url)
	}
	// case with a virtual service whose spec.hosts is present but empty: must not panic and yields no URL
	emptyHostsVS := &unstructured.Unstructured{}
	emptyHostsVS.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "networking.istio.io/v1alpha3",
		"kind":       "VirtualService",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
		"spec": map[string]interface{}{
			"hosts": []interface{}{},
		},
	})
	dynamicClient = fakedyn.NewSimpleDynamicClient(scheme, emptyHostsVS)
	url, err = services.FindURLFromVSIstio(dynamicClient, namespace, name)
	assert.Error(t, err)
	assert.Equal(t, "", url)
	// istio is optional: forbidden/unauthorized GETs are swallowed to ("", nil) rather than surfaced
	for _, swallowed := range []struct {
		name string
		err  error
	}{
		{"forbidden", apierrors.NewForbidden(schema.GroupResource{Resource: "virtualservices"}, name, fmt.Errorf("nope"))},
		{"unauthorized", apierrors.NewUnauthorized("nope")},
	} {
		errClient := fakedyn.NewSimpleDynamicClient(scheme)
		errClient.PrependReactor("get", "virtualservices", func(clienttesting.Action) (bool, runtime.Object, error) {
			return true, nil, swallowed.err
		})
		url, err = services.FindURLFromVSIstio(errClient, namespace, name)
		assert.NoError(t, err, swallowed.name)
		assert.Equal(t, "", url, swallowed.name)
	}
}

func TestFindURLFromHTTPRoute(t *testing.T) {
	t.Parallel()
	const namespace = "jx"
	const name = "myapp"

	newHTTPRoute := func(hostname, gatewayName, sectionName string) *unstructured.Unstructured {
		parentRef := map[string]interface{}{"name": gatewayName}
		if sectionName != "" {
			parentRef["sectionName"] = sectionName
		}
		u := &unstructured.Unstructured{}
		u.SetUnstructuredContent(map[string]interface{}{
			"apiVersion": "gateway.networking.k8s.io/v1",
			"kind":       "HTTPRoute",
			"metadata":   map[string]interface{}{"name": name, "namespace": namespace},
			"spec": map[string]interface{}{
				"parentRefs": []interface{}{parentRef},
				"hostnames":  []interface{}{hostname},
			},
		})
		return u
	}

	newGateway := func(gatewayName string, listeners ...map[string]interface{}) *unstructured.Unstructured {
		ls := make([]interface{}, 0, len(listeners))
		for _, l := range listeners {
			ls = append(ls, l)
		}
		u := &unstructured.Unstructured{}
		u.SetUnstructuredContent(map[string]interface{}{
			"apiVersion": "gateway.networking.k8s.io/v1",
			"kind":       "Gateway",
			"metadata":   map[string]interface{}{"name": gatewayName, "namespace": namespace},
			"spec":       map[string]interface{}{"listeners": ls},
		})
		return u
	}

	httpsListener := map[string]interface{}{"name": "https", "protocol": "HTTPS", "port": int64(443)}
	httpListener := map[string]interface{}{"name": "http", "protocol": "HTTP", "port": int64(80)}

	// newClient seeds a fake dynamic client. The HTTPRoute is added normally, but the Gateway
	// must be inserted under its real "gateways" resource: the fake client's UnsafeGuessKindToResource
	// heuristic mispluralises kind "Gateway" to "gatewaies", so a plain Add would not be GET-able.
	newClient := func(httpRoute, gateway *unstructured.Unstructured) dynamic.Interface {
		c := fakedyn.NewSimpleDynamicClient(runtime.NewScheme(), httpRoute)
		if gateway != nil {
			err := c.Tracker().Create(gatewayGVR, gateway, gateway.GetNamespace())
			assert.NoError(t, err)
		}
		return c
	}

	// newRouteClient seeds a fake client holding a single HTTPRoute with the given (possibly malformed) spec.
	newRouteClient := func(spec map[string]interface{}) dynamic.Interface {
		u := &unstructured.Unstructured{}
		u.SetUnstructuredContent(map[string]interface{}{
			"apiVersion": "gateway.networking.k8s.io/v1",
			"kind":       "HTTPRoute",
			"metadata":   map[string]interface{}{"name": name, "namespace": namespace},
			"spec":       spec,
		})
		return fakedyn.NewSimpleDynamicClient(runtime.NewScheme(), u)
	}

	// errClient returns a client whose GET on httproutes fails with the given error.
	errClient := func(getErr error) dynamic.Interface {
		c := fakedyn.NewSimpleDynamicClient(runtime.NewScheme())
		c.PrependReactor("get", "httproutes", func(clienttesting.Action) (bool, runtime.Object, error) {
			return true, nil, getErr
		})
		return c
	}
	groupResource := schema.GroupResource{Group: "gateway.networking.k8s.io", Resource: "httproutes"}

	testCases := []struct {
		name        string
		client      dynamic.Interface
		expectedURL string
		expectErr   bool
		errContains string
	}{
		{
			name:        "no HTTPRoute present is swallowed",
			client:      fakedyn.NewSimpleDynamicClient(runtime.NewScheme()),
			expectedURL: "",
		},
		{
			name:        "pinned HTTPS listener -> https",
			client:      newClient(newHTTPRoute("myapp.example.com", "gw", "https"), newGateway("gw", httpListener, httpsListener)),
			expectedURL: "https://myapp.example.com",
		},
		{
			name:        "first listener HTTP, no sectionName -> http",
			client:      newClient(newHTTPRoute("myapp.example.com", "gw", ""), newGateway("gw", httpListener, httpsListener)),
			expectedURL: "http://myapp.example.com",
		},
		{
			name:        "parent Gateway absent -> best-effort http",
			client:      newClient(newHTTPRoute("myapp.example.com", "missing-gw", "https"), nil),
			expectedURL: "http://myapp.example.com",
		},
		{
			// malformed content surfaces an error carrying the HTTPRoute identity, not best-effort http
			name:        "no parentRefs -> error with identity",
			client:      newRouteClient(map[string]interface{}{"hostnames": []interface{}{"myapp.example.com"}}),
			expectErr:   true,
			errContains: namespace + "/" + name,
		},
		{
			name:      "no hostnames -> error",
			client:    newRouteClient(map[string]interface{}{"parentRefs": []interface{}{map[string]interface{}{"name": "gw"}}}),
			expectErr: true,
		},
		{
			name:        "Forbidden is swallowed",
			client:      errClient(apierrors.NewForbidden(groupResource, name, fmt.Errorf("no access"))),
			expectedURL: "",
		},
		{
			name:        "Unauthorized is swallowed",
			client:      errClient(apierrors.NewUnauthorized("no auth")),
			expectedURL: "",
		},
		{
			name:      "internal error propagates",
			client:    errClient(apierrors.NewInternalError(fmt.Errorf("boom"))),
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			url, err := services.FindURLFromHTTPRoute(tc.client, namespace, name)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if tc.errContains != "" {
				assert.ErrorContains(t, err, tc.errContains)
			}
			assert.Equal(t, tc.expectedURL, url)
		})
	}
}

func TestFindURLFromService(t *testing.T) {
	t.Parallel()
	const namespace = "jx"

	// service does not exist -> empty url, no error
	client := fake.NewSimpleClientset()
	url, err := services.FindURLFromService(client, namespace, "missing")
	assert.NoError(t, err)
	assert.Equal(t, "", url)

	// service with the expose url annotation -> annotation value
	client = fake.NewSimpleClientset(&v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "exposed",
			Namespace:   namespace,
			Annotations: map[string]string{services.ExposeURLAnnotation: "https://exposed.example.com"},
		},
	})
	url, err = services.FindURLFromService(client, namespace, "exposed")
	assert.NoError(t, err)
	assert.Equal(t, "https://exposed.example.com", url)

	// LoadBalancer service without annotation -> GetServiceURL fallback to the ingress IP
	client = fake.NewSimpleClientset(&v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "lb",
			Namespace: namespace,
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeLoadBalancer,
		},
		Status: v1.ServiceStatus{
			LoadBalancer: v1.LoadBalancerStatus{
				Ingress: []v1.LoadBalancerIngress{{IP: "1.2.3.4"}},
			},
		},
	})
	url, err = services.FindURLFromService(client, namespace, "lb")
	assert.NoError(t, err)
	assert.Equal(t, "http://1.2.3.4/", url)
}

func TestFindURLFromIngress(t *testing.T) {
	t.Parallel()
	const namespace = "jx"

	// ingress does not exist -> empty url, no error
	client := fake.NewSimpleClientset()
	url, err := services.FindURLFromIngress(client, namespace, "missing")
	assert.NoError(t, err)
	assert.Equal(t, "", url)

	// ingress present -> IngressURL result
	client = fake.NewSimpleClientset(&nv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hook",
			Namespace: namespace,
		},
		Spec: nv1.IngressSpec{
			Rules: []nv1.IngressRule{{Host: "hook-jx.1.2.3.4.nip.io"}},
		},
	})
	url, err = services.FindURLFromIngress(client, namespace, "hook")
	assert.NoError(t, err)
	assert.Equal(t, "http://hook-jx.1.2.3.4.nip.io", url)
}

// TestFindServiceURLWithDynamicClient checks the orchestration of URL searching across all resource types
func TestFindServiceURLWithDynamicClient(t *testing.T) {
	t.Parallel()
	const namespace = "jx"
	const name = "myapp"

	vsWithHost := &unstructured.Unstructured{}
	vsWithHost.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "networking.istio.io/v1alpha3",
		"kind":       "VirtualService",
		"metadata":   map[string]interface{}{"name": name, "namespace": namespace},
		"spec":       map[string]interface{}{"hosts": []interface{}{"myapp.example.com"}},
	})

	httpRoute := &unstructured.Unstructured{}
	httpRoute.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "gateway.networking.k8s.io/v1",
		"kind":       "HTTPRoute",
		"metadata":   map[string]interface{}{"name": name, "namespace": namespace},
		"spec": map[string]interface{}{
			"parentRefs": []interface{}{map[string]interface{}{"name": "gw", "sectionName": "https"}},
			"hostnames":  []interface{}{"route.example.com"},
		},
	})
	gateway := &unstructured.Unstructured{}
	gateway.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "gateway.networking.k8s.io/v1",
		"kind":       "Gateway",
		"metadata":   map[string]interface{}{"name": "gw", "namespace": namespace},
		"spec": map[string]interface{}{
			"listeners": []interface{}{map[string]interface{}{"name": "https", "protocol": "HTTPS", "port": int64(443)}},
		},
	})
	// HTTPRoute is added normally; the Gateway must be inserted under its real "gateways" resource
	httpRouteClient := fakedyn.NewSimpleDynamicClient(runtime.NewScheme(), httpRoute)
	if err := httpRouteClient.Tracker().Create(gatewayGVR, gateway, namespace); err != nil {
		t.Fatal(err)
	}

	// a malformed HTTPRoute (valid hostname, but no parentRefs) makes FindURLFromHTTPRoute return
	// an error; the orchestrator should swallow it - like Istio - and keep falling through
	malformedRoute := &unstructured.Unstructured{}
	malformedRoute.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "gateway.networking.k8s.io/v1",
		"kind":       "HTTPRoute",
		"metadata":   map[string]interface{}{"name": name, "namespace": namespace},
		"spec":       map[string]interface{}{"hostnames": []interface{}{"route.example.com"}},
	})
	malformedRouteClient := fakedyn.NewSimpleDynamicClient(runtime.NewScheme(), malformedRoute)

	// clientset whose service GET fails with a non-NotFound error
	errClient := fake.NewSimpleClientset()
	errClient.PrependReactor("get", "services", func(clienttesting.Action) (bool, runtime.Object, error) {
		return true, nil, apierrors.NewServiceUnavailable("boom")
	})

	// clientset whose ingress GET fails with a non-NotFound error (e.g. forbidden),
	// while the service is simply absent - discovery should still fall through to istio
	ingErrClient := fake.NewSimpleClientset()
	ingErrClient.PrependReactor("get", "ingresses", func(clienttesting.Action) (bool, runtime.Object, error) {
		return true, nil, apierrors.NewForbidden(schema.GroupResource{Resource: "ingresses"}, name, fmt.Errorf("nope"))
	})

	testCases := []struct {
		name          string
		client        *fake.Clientset
		dynamicClient dynamic.Interface
		expectedURL   string
		expectErr     bool
	}{
		{
			name: "service wins (no fall-through)",
			client: fake.NewSimpleClientset(&v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:        name,
					Namespace:   namespace,
					Annotations: map[string]string{services.ExposeURLAnnotation: "https://svc.example.com"},
				},
			}),
			dynamicClient: fakedyn.NewSimpleDynamicClient(runtime.NewScheme()),
			expectedURL:   "https://svc.example.com",
		},
		{
			name: "falls through to ingress",
			client: fake.NewSimpleClientset(&nv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
				Spec:       nv1.IngressSpec{Rules: []nv1.IngressRule{{Host: "ing.example.com"}}},
			}),
			dynamicClient: fakedyn.NewSimpleDynamicClient(runtime.NewScheme()),
			expectedURL:   "http://ing.example.com",
		},
		{
			name:          "falls through to HTTPRoute",
			client:        fake.NewSimpleClientset(),
			dynamicClient: httpRouteClient,
			expectedURL:   "https://route.example.com",
		},
		{
			name:          "falls through to istio",
			client:        fake.NewSimpleClientset(),
			dynamicClient: fakedyn.NewSimpleDynamicClient(runtime.NewScheme(), vsWithHost),
			expectedURL:   "http://myapp.example.com",
		},
		{
			// malformed HTTPRoute error is swallowed (no expectErr) and the chain falls through
			name:          "malformed HTTPRoute is swallowed",
			client:        fake.NewSimpleClientset(),
			dynamicClient: malformedRouteClient,
			expectedURL:   "",
		},
		{
			name:          "nothing found",
			client:        fake.NewSimpleClientset(),
			dynamicClient: fakedyn.NewSimpleDynamicClient(runtime.NewScheme()),
			expectedURL:   "",
		},
		{
			name:          "service error propagates",
			client:        errClient,
			dynamicClient: fakedyn.NewSimpleDynamicClient(runtime.NewScheme()),
			expectErr:     true,
		},
		{
			name:          "ingress error falls through to istio",
			client:        ingErrClient,
			dynamicClient: fakedyn.NewSimpleDynamicClient(runtime.NewScheme(), vsWithHost),
			expectedURL:   "http://myapp.example.com",
		},
		{
			name:          "ingress error surfaces when nothing else found",
			client:        ingErrClient,
			dynamicClient: fakedyn.NewSimpleDynamicClient(runtime.NewScheme()),
			expectErr:     true,
			expectedURL:   "",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			url, err := services.FindServiceURLWithDynamicClient(tc.client, namespace, name, tc.dynamicClient)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectedURL, url)
		})
	}
}
