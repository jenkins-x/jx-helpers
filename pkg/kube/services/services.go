package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/jenkins-x/jx-helpers/v3/pkg/kube"
	"github.com/jenkins-x/jx-helpers/v3/pkg/stringhelpers"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"sort"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"
	nv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

const ExposeURLAnnotation = "fabric8.io/exposeUrl"

type ServiceURL struct {
	Name string
	URL  string
}

func GetServices(client kubernetes.Interface, ns string) (map[string]*v1.Service, error) {
	answer := map[string]*v1.Service{}
	list, err := client.CoreV1().Services(ns).List(context.TODO(), meta_v1.ListOptions{})
	if err != nil {
		return answer, fmt.Errorf("failed to load Services %s", err)
	}
	for _, r := range list.Items {
		name := r.Name
		c := r
		answer[name] = &c
	}
	return answer, nil
}

// GetServicesByName returns a list of Service objects from a list of service names
func GetServicesByName(client kubernetes.Interface, ns string, services []string) ([]*v1.Service, error) {
	answer := make([]*v1.Service, 0)
	svcList, err := client.CoreV1().Services(ns).List(context.TODO(), meta_v1.ListOptions{})
	if err != nil {
		return answer, fmt.Errorf("listing the services in namespace %q: %w", ns, err)
	}
	for _, s := range svcList.Items {
		i := stringhelpers.StringArrayIndex(services, s.GetName())
		if i > 0 {
			c := s
			answer = append(answer, &c)
		}
	}
	return answer, nil
}

func GetServiceNames(client kubernetes.Interface, ns string, filter string) ([]string, error) {
	var names []string
	list, err := client.CoreV1().Services(ns).List(context.TODO(), meta_v1.ListOptions{})
	if err != nil {
		return names, fmt.Errorf("failed to load Services %s", err)
	}
	for _, r := range list.Items {
		name := r.Name
		if filter == "" || strings.Contains(name, filter) {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names, nil
}

func GetServiceURLFromMap(services map[string]*v1.Service, name string) string {
	return GetServiceURL(services[name])
}

func getIstioVirtualService(dynamicClient dynamic.Interface, namespace, name string) (*unstructured.Unstructured, error) {
	dynamicClient, err := kube.LazyCreateDynamicClient(dynamicClient)
	if err != nil {
		return nil, err
	}
	//  Create a GVR which represents an Istio Virtual Service.
	virtualServiceGVR := schema.GroupVersionResource{
		Group:    "networking.istio.io",
		Version:  "v1alpha3",
		Resource: "virtualservices",
	}
	virtualService, err := dynamicClient.Resource(virtualServiceGVR).Namespace(namespace).Get(context.TODO(), name, meta_v1.GetOptions{})
	return virtualService, err
}

// getURLFromVirtualService finds url from virtual services istio
func getURLFromVirtualService(virtualService *unstructured.Unstructured) (string, error) {
	vs := virtualService.Object
	if spec, ok := vs["spec"].(map[string]interface{}); ok {
		if hosts, ok := spec["hosts"].([]interface{}); ok && len(hosts) > 0 {
			if host := fmt.Sprintf("%v", hosts[0]); host != "" {
				return "http://" + host, nil
			}
		}
	}
	return "", errors.New("no URL found in the Istio virtual service")
}

// FindURLFromVSIstio finds the host from istio virtual service
func FindURLFromVSIstio(dynamicClient dynamic.Interface, namespace, name string) (string, error) {
	log.Logger().Debugf("finding url from VS Istio %s in namespace %s", name, namespace)
	virtualService, err := getIstioVirtualService(dynamicClient, namespace, name)
	if err != nil {
		switch {
		// istio is optional: missing resource, CRD or lack of RBAC should not log an error
		case apierrors.IsNotFound(err), apierrors.IsForbidden(err), apierrors.IsUnauthorized(err):
			log.Logger().Debugf("Istio VS %s not reachable in namespace %s", name, namespace)
			return "", nil
		default:
			return "", fmt.Errorf("finding the Istio VS %s in namespace %s: %w", name, namespace, err)
		}
	}
	log.Logger().Debugf("attempting to find via istio virtual services")
	return getURLFromVirtualService(virtualService)
}

// getHTTPRoute gets the named HTTPRoute. The dynamic client must already be realized.
func getHTTPRoute(dynamicClient dynamic.Interface, namespace, name string) (*unstructured.Unstructured, error) {
	httpRouteGVR := schema.GroupVersionResource{
		Group:    "gateway.networking.k8s.io",
		Version:  "v1",
		Resource: "httproutes",
	}
	return dynamicClient.Resource(httpRouteGVR).Namespace(namespace).Get(context.TODO(), name, meta_v1.GetOptions{})
}

// getGateway gets the named Gateway. The dynamic client must already be realized.
func getGateway(dynamicClient dynamic.Interface, namespace, name string) (*unstructured.Unstructured, error) {
	gatewayGVR := schema.GroupVersionResource{
		Group:    "gateway.networking.k8s.io",
		Version:  "v1",
		Resource: "gateways",
	}
	return dynamicClient.Resource(gatewayGVR).Namespace(namespace).Get(context.TODO(), name, meta_v1.GetOptions{})
}

// parseGatewayParentRef extracts the parent Gateway reference from the HTTPRoute spec.
func parseGatewayParentRef(httpRouteNamespace string, spec map[string]interface{}) (namespace, name, sectionName string, err error) {
	parentRefs, ok := spec["parentRefs"].([]interface{})
	if !ok || len(parentRefs) == 0 {
		return "", "", "", errors.New("missing spec.parentRefs in the HTTPRoute")
	}
	parentRef, ok := parentRefs[0].(map[string]interface{})
	if !ok {
		return "", "", "", errors.New("invalid spec.parentRefs[0] in the HTTPRoute")
	}
	name, _ = parentRef["name"].(string)
	if name == "" {
		return "", "", "", errors.New("missing spec.parentRefs[0].name in the HTTPRoute")
	}
	namespace = httpRouteNamespace
	if ns, ok := parentRef["namespace"].(string); ok && ns != "" {
		namespace = ns
	}
	sectionName, _ = parentRef["sectionName"].(string)
	return namespace, name, sectionName, nil
}

// getGatewayListenerProtocol returns the protocol (HTTP/HTTPS) of the parent Gateway's listener.
func getGatewayListenerProtocol(dynamicClient dynamic.Interface, namespace, name, sectionName string) (string, error) {
	gateway, err := getGateway(dynamicClient, namespace, name)
	if err != nil {
		return "", err
	}
	return getListenerProtocol(gateway, sectionName)
}

// getListenerProtocol returns the protocol of the Gateway listener named by sectionName, or the first listener when empty.
func getListenerProtocol(gateway *unstructured.Unstructured, sectionName string) (string, error) {
	spec, ok := gateway.Object["spec"].(map[string]interface{})
	if !ok {
		return "", errors.New("no spec found in the Gateway")
	}
	listeners, ok := spec["listeners"].([]interface{})
	if !ok || len(listeners) == 0 {
		return "", errors.New("no listeners found in the Gateway")
	}
	for _, l := range listeners {
		listener, ok := l.(map[string]interface{})
		if !ok {
			continue
		}
		// when a sectionName is given, only the listener with the matching name is relevant
		if sectionName != "" {
			if name, _ := listener["name"].(string); name != sectionName {
				continue
			}
		}
		if protocol, ok := listener["protocol"].(string); ok && protocol != "" {
			return protocol, nil
		}
		// no sectionName: only the first listener is considered
		if sectionName == "" {
			break
		}
	}
	return "", errors.New("no matching listener protocol found in the Gateway")
}

// getURLFromHTTPRoute builds the URL from the HTTPRoute's hostname and parent Gateway scheme.
func getURLFromHTTPRoute(dynamicClient dynamic.Interface, httpRoute *unstructured.Unstructured) (string, error) {
	spec, ok := httpRoute.Object["spec"].(map[string]interface{})
	if !ok {
		return "", errors.New("missing spec in the HTTPRoute")
	}
	hostnames, ok := spec["hostnames"].([]interface{})
	if !ok || len(hostnames) == 0 {
		return "", errors.New("missing spec.hostnames in the HTTPRoute")
	}
	hostname, ok := hostnames[0].(string)
	if !ok || hostname == "" {
		return "", errors.New("invalid spec.hostnames[0] in the HTTPRoute")
	}

	// a malformed parentRef means the HTTPRoute is invalid - surface it
	gatewayNamespace, gatewayName, sectionName, err := parseGatewayParentRef(httpRoute.GetNamespace(), spec)
	if err != nil {
		return "", err
	}

	// derive the scheme from the parent Gateway listener's protocol, defaulting to http
	scheme := "http"
	protocol, err := getGatewayListenerProtocol(dynamicClient, gatewayNamespace, gatewayName, sectionName)
	if err != nil {
		log.Logger().Debugf("unable to determine protocol from parent Gateway listener, defaulting to http: %s", err)
	} else if strings.EqualFold(protocol, "HTTPS") {
		scheme = "https"
	}
	return scheme + "://" + hostname, nil
}

// FindURLFromHTTPRoute finds the URL from the HTTPRoute resource, realizing a dynamic client if needed.
func FindURLFromHTTPRoute(dynamicClient dynamic.Interface, namespace, name string) (string, error) {
	dynamicClient, err := kube.LazyCreateDynamicClient(dynamicClient)
	if err != nil {
		return "", err
	}
	return findURLFromHTTPRouteRealized(dynamicClient, namespace, name)
}

// findURLFromHTTPRouteRealized finds the HTTPRoute URL using an already-realized dynamic client.
func findURLFromHTTPRouteRealized(dynamicClient dynamic.Interface, namespace, name string) (string, error) {
	log.Logger().Debugf("finding url from HTTP route %s in namespace %s", name, namespace)
	httpRoute, err := getHTTPRoute(dynamicClient, namespace, name)
	if err != nil {
		switch {
		// HTTPRoute is optional: missing resource, CRD or lack of RBAC should not surface an error
		case apierrors.IsNotFound(err), apierrors.IsForbidden(err), apierrors.IsUnauthorized(err):
			log.Logger().Debugf("HTTPRoute %s not reachable in namespace %s", name, namespace)
			return "", nil
		default:
			return "", fmt.Errorf("finding the HTTP route %s in namespace %s: %w", name, namespace, err)
		}
	}
	log.Logger().Debugf("attempting to find URL via HTTPRoute")
	url, err := getURLFromHTTPRoute(dynamicClient, httpRoute)
	if err != nil {
		return "", fmt.Errorf("parsing HTTPRoute %s/%s: %w", namespace, name, err)
	}
	return url, nil
}

// FindURLFromIngress finds the URL from the Ingress resource using the kubernetes client
func FindURLFromIngress(client kubernetes.Interface, namespace string, name string) (string, error) {
	log.Logger().Debugf("finding Ingress url for %s in namespace %s", name, namespace)
	ingress, err := client.NetworkingV1().Ingresses(namespace).Get(context.TODO(), name, meta_v1.GetOptions{})
	if err != nil {
		switch {
		// ingress was not found but no other errors occurred. Log and return nil err
		case apierrors.IsNotFound(err):
			log.Logger().Debugf("Ingress %s not found in namespace %s", name, namespace)
			return "", nil
		default:
			return "", fmt.Errorf("finding the ingress %s in namespace %s: %w", name, namespace, err)
		}
	}
	return IngressURL(ingress), nil
}

// FindURLFromService finds the URL from the service resource using the kubernetes client
func FindURLFromService(client kubernetes.Interface, namespace string, name string) (string, error) {
	log.Logger().Debugf("finding service url for %s in namespace %s", name, namespace)
	service, err := client.CoreV1().Services(namespace).Get(context.TODO(), name, meta_v1.GetOptions{})
	if err != nil {
		switch {
		// service was not found but no other errors occurred. Log and return nil err
		case apierrors.IsNotFound(err):
			log.Logger().Debugf("service %s not found in namespace %s", name, namespace)
			return "", nil
		default:
			return "", fmt.Errorf("finding the service %s in namespace %s: %w", name, namespace, err)
		}
	}
	return GetServiceURL(service), nil
}

func FindServiceURLWithDynamicClient(client kubernetes.Interface, namespace string, name string, dynamicClient dynamic.Interface) (string, error) {
	url, err := FindURLFromService(client, namespace, name)
	if err != nil {
		return "", err
	}
	if url != "" {
		log.Logger().Debugf("found the service url %s", url)
		return url, nil
	}
	log.Logger().Debugf("couldn't find url via service, attempting to look up via ingress")

	// let's try finding the service via Ingress
	// hold onto genuine lookup failures and attempt further discoveries before surfacing them
	url, err = FindURLFromIngress(client, namespace, name)
	if err != nil {
		log.Logger().Debugf("unable to find url via ingress for %s in namespace %s - err %s", name, namespace, err)
	}
	if url != "" {
		log.Logger().Debugf("found ingress url %s", url)
		return url, nil
	}

	log.Logger().Debugf("couldn't find url via ingress, attempting to look up via HTTPRoute")

	// realize the dynamic client once and share it across the HTTPRoute and Istio lookups. these
	// are skipped entirely when a client can't be built, so the realized helpers never get a nil client.
	dynamicClient, dcErr := kube.LazyCreateDynamicClient(dynamicClient)
	if dcErr != nil {
		log.Logger().Debugf("unable to create dynamic client for %s in namespace %s, skipping HTTPRoute and istio lookups - err %s", name, namespace, dcErr)
	} else {
		// let's try finding the URL via HTTPRoute
		// use a dedicated error variable so an optional HTTPRoute failure doesn't clobber a genuine ingress error
		url, hrErr := findURLFromHTTPRouteRealized(dynamicClient, namespace, name)
		if hrErr != nil {
			log.Logger().Debugf("unable to find url via HTTPRoute for %s in namespace %s - err %s", name, namespace, hrErr)
		}
		if url != "" {
			log.Logger().Debugf("found HTTPRoute url %s", url)
			return url, nil
		}

		log.Logger().Debugf("couldn't find url via HTTPRoute, attempting to look up via istio virtual services")

		// let's try finding the service via Istio Virtual Services
		url, vsErr := FindURLFromVSIstio(dynamicClient, namespace, name)
		// log errors from Istio but don't propagate them as Istio is an optional dependency
		if vsErr != nil {
			log.Logger().Debugf("unable to find url via istio virtual service for %s in namespace %s - err %s", name, namespace, vsErr)
		}
		if url != "" {
			log.Logger().Debugf("Found istio virtual service url %s", url)
			return url, nil
		}
	}

	// nothing found anywhere - if a lookup action errored, surface it here
	if err != nil {
		return "", err
	}

	log.Logger().Debugf("couldn't find service url, exiting")
	return "", nil
}

func FindServiceURL(client kubernetes.Interface, namespace string, name string) (string, error) {
	return FindServiceURLWithDynamicClient(client, namespace, name, nil)
}

func FindIngressURL(client kubernetes.Interface, namespace string, name string) (string, error) {
	log.Logger().Debugf("Finding ingress url for %s in namespace %s", name, namespace)
	// lets try find the service via Ingress
	ing, err := client.NetworkingV1().Ingresses(namespace).Get(context.TODO(), name, meta_v1.GetOptions{})
	if err != nil {
		log.Logger().Debugf("Error finding ingress for %s in namespace %s - err %s", name, namespace, err)
		return "", nil
	}

	url := IngressURL(ing)
	if url == "" {
		log.Logger().Debugf("Unable to find url via ingress for %s in namespace %s", name, namespace)
	}
	return url, nil
}

// IngressURL returns the URL for the ingres
func IngressURL(ing *nv1.Ingress) string {
	if ing == nil {
		log.Logger().Debug("Ingress is nil, returning empty string for url")
		return ""
	}
	if len(ing.Spec.Rules) == 0 {
		log.Logger().Debug("Ingress spec has no rules, returning empty string for url")
		return ""
	}

	rule := ing.Spec.Rules[0]
	for _, tls := range ing.Spec.TLS {
		for _, h := range tls.Hosts {
			if h != "" {
				url := "https://" + h
				log.Logger().Debugf("found service url %s", url)
				return url
			}
		}
	}
	ann := ing.Annotations
	hostname := rule.Host
	if hostname == "" && ann != nil {
		hostname = ann[kube.AnnotationHost]
	}
	if hostname == "" {
		log.Logger().Debug("Could not find hostname from rule or ingress annotation")
		return ""
	}

	url := "http://" + hostname
	if rule.HTTP != nil {
		if len(rule.HTTP.Paths) > 0 {
			p := rule.HTTP.Paths[0].Path
			if p != "" {
				url += p
			}
		}
	}
	log.Logger().Debugf("found service url %s", url)
	return url
}

// IngressHost returns the host for the ingres
func IngressHost(ing *nv1.Ingress) string {
	if ing != nil {
		if len(ing.Spec.Rules) > 0 {
			rule := ing.Spec.Rules[0]
			hostname := rule.Host
			for _, tls := range ing.Spec.TLS {
				for _, h := range tls.Hosts {
					if h != "" {
						return h
					}
				}
			}
			if hostname != "" {
				return hostname
			}
		}
	}
	return ""
}

// IngressProtocol returns the scheme (https / http) for the Ingress
func IngressProtocol(ing *nv1.Ingress) string {
	if ing != nil && len(ing.Spec.TLS) == 0 {
		return "http"
	}
	return "https"
}

func FindServiceHostname(client kubernetes.Interface, namespace string, name string) (string, error) {
	// lets try find the service via Ingress
	ing, err := client.NetworkingV1().Ingresses(namespace).Get(context.TODO(), name, meta_v1.GetOptions{})
	if ing != nil && err == nil {
		if len(ing.Spec.Rules) > 0 {
			rule := ing.Spec.Rules[0]
			hostname := rule.Host
			for _, tls := range ing.Spec.TLS {
				for _, h := range tls.Hosts {
					if h != "" {
						return h, nil
					}
				}
			}
			if hostname != "" {
				return hostname, nil
			}
		}
	}
	return "", nil
}

// FindService looks up a service by name across all namespaces
func FindService(client kubernetes.Interface, name string) (*v1.Service, error) {
	nsl, err := client.CoreV1().Namespaces().List(context.TODO(), meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, ns := range nsl.Items {
		svc, err := client.CoreV1().Services(ns.GetName()).Get(context.TODO(), name, meta_v1.GetOptions{})
		if err == nil {
			return svc, nil
		}
	}
	return nil, errors.New("Service not found!")
}

// GetServiceURL returns the
func GetServiceURL(svc *v1.Service) string {
	if svc == nil {
		return ""
	}
	url := ""
	if svc.Annotations != nil {
		url = svc.Annotations[ExposeURLAnnotation]
	}
	if url == "" {
		scheme := "http"
		if svc.Spec.Ports != nil {
			for _, port := range svc.Spec.Ports {
				if port.Port == 443 {
					scheme = "https"
					break
				}
			}
		}

		// lets check if its a LoadBalancer
		if svc.Spec.Type == v1.ServiceTypeLoadBalancer {
			for _, ing := range svc.Status.LoadBalancer.Ingress {
				if ing.IP != "" {
					return scheme + "://" + ing.IP + "/"
				}
				if ing.Hostname != "" {
					return scheme + "://" + ing.Hostname + "/"
				}
			}
		}
	}
	return url
}

// FindServiceSchemePort parses the service definition and interprets http scheme in the absence of an external ingress
func FindServiceSchemePort(client kubernetes.Interface, namespace string, name string) (string, string, error) {
	svc, err := client.CoreV1().Services(namespace).Get(context.TODO(), name, meta_v1.GetOptions{})
	if err != nil {
		return "", "", fmt.Errorf("failed to find service %s in namespace %s: %w", name, namespace, err)
	}
	return ExtractServiceSchemePort(svc)
}

func GetServiceURLFromName(c kubernetes.Interface, name, ns string) (string, error) {
	return FindServiceURL(c, ns, name)
}

func FindServiceURLs(client kubernetes.Interface, namespace string) ([]ServiceURL, error) {
	options := meta_v1.ListOptions{}
	var urls []ServiceURL
	svcs, err := client.CoreV1().Services(namespace).List(context.TODO(), options)
	if err != nil {
		return urls, err
	}
	for _, svc := range svcs.Items {
		url := GetServiceURL(&svc)
		if url == "" {
			url, _ = FindServiceURL(client, namespace, svc.Name)
		}
		if len(url) > 0 {
			urls = append(urls, ServiceURL{
				Name: svc.Name,
				URL:  url,
			})
		}
	}
	return urls, nil
}

func HasExternalAddress(svc *v1.Service) bool {
	for _, v := range svc.Status.LoadBalancer.Ingress {
		if v.IP != "" || v.Hostname != "" {
			return true
		}
	}
	return false
}

func GetService(client kubernetes.Interface, currentNamespace, targetNamespace, serviceName string) error {
	svc := v1.Service{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      serviceName,
			Namespace: currentNamespace,
		},
		Spec: v1.ServiceSpec{
			Type:         v1.ServiceTypeExternalName,
			ExternalName: fmt.Sprintf("%s.%s.svc.cluster.local", serviceName, targetNamespace),
		},
	}
	_, err := client.CoreV1().Services(currentNamespace).Create(context.TODO(), &svc, meta_v1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

// ExtractServiceSchemePort is a utility function to interpret http scheme and port information from k8s service definitions
func ExtractServiceSchemePort(svc *v1.Service) (string, string, error) {
	scheme := ""
	port := ""

	found := false

	// Search in order of degrading priority
	for _, p := range svc.Spec.Ports {
		if p.Port == 443 { // Prefer 443/https if found
			scheme = "https"
			port = "443"
			found = true
			break
		}
	}

	if !found {
		for _, p := range svc.Spec.Ports {
			if p.Port == 80 { // Use 80/http if found
				scheme = "http"
				port = "80"
				found = true
			}
		}
	}

	if !found { // No conventional ports, so search for named https ports
		for _, p := range svc.Spec.Ports {
			if p.Protocol == "TCP" {
				if p.Name == "https" {
					scheme = "https"
					port = strconv.FormatInt(int64(p.Port), 10)
					found = true
					break
				}
			}
		}
	}

	if !found { // No conventional ports, so search for named http ports
		for _, p := range svc.Spec.Ports {
			if p.Name == "http" {
				scheme = "http"
				port = strconv.FormatInt(int64(p.Port), 10)
				found = true
				break
			}
		}
	}

	return scheme, port, nil
}
