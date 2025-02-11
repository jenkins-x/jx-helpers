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

// getUrlFromVirtualService finds url from virtual services istio
func getUrlFromVirtualService(virtualService *unstructured.Unstructured) (string, error) {
	vs := virtualService.Object
	if spec, ok := vs["spec"].(map[string]interface{}); ok {
		if hosts, ok := spec["hosts"].([]interface{}); ok {
			return "http://" + fmt.Sprintf("%v", hosts[0]), nil
		}
	}
	return "", errors.New("No url found in the virtual service")
}

// FindUrlFromVsIstio finds the host from istio virtual service
func FindUrlFromVsIstio(dynamicClient dynamic.Interface, namespace, name string) (string, error) {
	virtualService, err := getIstioVirtualService(dynamicClient, namespace, name)
	if err != nil {
		return "", nil
	}
	log.Logger().Debugf("Attempting to find via istio virtual services")
	return getUrlFromVirtualService(virtualService)
}

func FindServiceURLWithDynamicClient(client kubernetes.Interface, namespace string, name string, dynamicClient dynamic.Interface) (string, error) {
	log.Logger().Debugf("Finding service url for %s in namespace %s", name, namespace)
	svc, err := client.CoreV1().Services(namespace).Get(context.TODO(), name, meta_v1.GetOptions{})
	if err != nil && apierrors.IsNotFound(err) {
		err = nil
	}
	if err != nil {
		return "", fmt.Errorf("finding the service %s in namespace %s: %w", name, namespace, err)
	}
	answer := ""
	if svc != nil {
		answer = GetServiceURL(svc)
	}
	if answer != "" {
		log.Logger().Debugf("Found service url %s", answer)
		return answer, nil
	}

	log.Logger().Debugf("Couldn't find service url, attempting to look up via ingress")

	// lets try find the service via Ingress
	ing, err := client.NetworkingV1().Ingresses(namespace).Get(context.TODO(), name, meta_v1.GetOptions{})
	if err != nil && apierrors.IsNotFound(err) {
		err = nil
	}
	if err != nil {
		log.Logger().Debugf("Unable to finding ingress for %s in namespace %s - err %s", name, namespace, err)
		url, vs_err := FindUrlFromVsIstio(dynamicClient, namespace, name)
		if url != "" && vs_err == nil {
			return url, nil
		}
		if vs_err != nil {
			log.Logger().Debugf("Unable to finding istio for %s in namespace %s - err %s", name, namespace, vs_err)
		}
		return "", fmt.Errorf("getting ingress for service %q in namespace %s: %w", name, namespace, err)
	}
	url := ""

	url = IngressURL(ing)

	if url == "" {
		log.Logger().Debugf("Unable to find service url via ingress for %s in namespace %s", name, namespace)
		url, vs_err := FindUrlFromVsIstio(dynamicClient, namespace, name)
		if url != "" && vs_err == nil {
			return url, nil
		}
		if vs_err != nil {
			log.Logger().Debugf("Unable to finding istio for %s in namespace %s - err %s", name, namespace, vs_err)
		}
	}
	return url, nil
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
	url := ""
	if svc != nil && svc.Annotations != nil {
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
