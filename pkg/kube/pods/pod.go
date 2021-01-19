package pods

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/jenkins-x/jx-helpers/v3/pkg/kube/naming"
	"github.com/jenkins-x/jx-helpers/v3/pkg/termcolor"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	tools_watch "k8s.io/client-go/tools/watch"
)

// credit https://github.com/kubernetes/kubernetes/blob/8719b4a/pkg/api/v1/pod/util.go
// IsPodReady returns true if a pod is ready; false otherwise.
func IsPodReady(pod *v1.Pod) bool {
	phase := pod.Status.Phase
	if phase != v1.PodRunning || pod.DeletionTimestamp != nil {
		return false
	}
	return IsPodReadyConditionTrue(pod.Status)
}

// IsPodPending returns true if a pod is pending
func IsPodPending(pod *v1.Pod) bool {
	switch pod.Status.Phase {
	case v1.PodFailed, v1.PodSucceeded, v1.PodRunning:
		return false
	default:
		return true
	}
}

// IsPodCompleted returns true if a pod is completed (succeeded or failed); false otherwise.
func IsPodCompleted(pod *v1.Pod) bool {
	phase := pod.Status.Phase
	if phase == v1.PodSucceeded || phase == v1.PodFailed {
		return true
	}
	return false
}

// IsPodSucceeded returns true if a pod is succeeded
func IsPodSucceeded(pod *v1.Pod) bool {
	phase := pod.Status.Phase
	if phase == v1.PodSucceeded {
		return true
	}
	return false
}

// IsPodFailed returns true if a pod is failed
func IsPodFailed(pod *v1.Pod) bool {
	switch pod.Status.Phase {
	case v1.PodSucceeded, v1.PodPending, v1.PodReasonUnschedulable:
		return false
	case v1.PodFailed:
		return true
	case v1.PodRunning:
		for _, c := range pod.Status.ContainerStatuses {
			if c.State.Running != nil {
				continue
			}
			terminated := c.State.Terminated
			if terminated == nil && c.State.Waiting != nil {
				terminated = c.LastTerminationState.Terminated
			}
			if terminated != nil && terminated.ExitCode != 0 {
				return true
			}
		}
	}
	return false
}

// credit https://github.com/kubernetes/kubernetes/blob/8719b4a/pkg/api/v1/pod/util.go
// IsPodReady retruns true if a pod is ready; false otherwise.
func IsPodReadyConditionTrue(status v1.PodStatus) bool {
	condition := GetPodReadyCondition(status)
	return condition != nil && condition.Status == v1.ConditionTrue
}

func PodStatus(pod *v1.Pod) string {
	if pod.DeletionTimestamp != nil {
		return "Terminating"
	}
	phase := pod.Status.Phase
	if IsPodReady(pod) {
		return "Ready"
	}
	return string(phase)
}

// credit https://github.com/kubernetes/kubernetes/blob/8719b4a/pkg/api/v1/pod/util.go
// Extracts the pod ready condition from the given status and returns that.
// Returns nil if the condition is not present.
func GetPodReadyCondition(status v1.PodStatus) *v1.PodCondition {
	_, condition := GetPodCondition(&status, v1.PodReady)
	return condition
}

// credit https://github.com/kubernetes/kubernetes/blob/8719b4a/pkg/api/v1/pod/util.go
// GetPodCondition extracts the provided condition from the given status and returns that.
// Returns nil and -1 if the condition is not present, and the index of the located condition.
func GetPodCondition(status *v1.PodStatus, conditionType v1.PodConditionType) (int, *v1.PodCondition) {
	if status == nil {
		return -1, nil
	}
	for i := range status.Conditions {
		if status.Conditions[i].Type == conditionType {
			return i, &status.Conditions[i]
		}
	}
	return -1, nil
}

// GetCurrentPod returns the current pod the code is running in or nil if it cannot be deduced
func GetCurrentPod(kubeClient kubernetes.Interface, ns string) (*v1.Pod, error) {
	name := os.Getenv("HOSTNAME")
	if name == "" {
		return nil, nil
	}
	name = naming.ToValidName(name)
	return kubeClient.CoreV1().Pods(ns).Get(context.TODO(), name, metav1.GetOptions{})
}

func waitForPodSelector(client kubernetes.Interface, namespace string, options metav1.ListOptions,
	timeout time.Duration, condition func(event watch.Event) (bool, error)) error {
	w, err := client.CoreV1().Pods(namespace).Watch(context.TODO(), options)
	if err != nil {
		return err
	}
	defer w.Stop()

	ctx, _ := context.WithTimeout(context.Background(), timeout)
	_, err = tools_watch.UntilWithoutRetry(ctx, w, condition)

	if err == wait.ErrWaitTimeout {
		return fmt.Errorf("pod %s never became ready", ListOptionsString(options))
	}
	return err
}

// ListOptionsString returns a string summary of the list options
func ListOptionsString(options metav1.ListOptions) string {
	var values []string
	if options.FieldSelector != "" {
		values = append(values, options.FieldSelector)
	}
	if options.LabelSelector != "" {
		values = append(values, "selector: "+options.LabelSelector)
	}
	if options.ResourceVersion != "" {
		values = append(values, "resourceVersion: "+options.ResourceVersion)
	}
	return strings.Join(values, ", ")
}

// HasContainerStarted returns true if the given Container has started running
func HasContainerStarted(pod *v1.Pod, idx int) bool {
	if pod == nil {
		return false
	}
	_, statuses, _ := GetContainersWithStatusAndIsInit(pod)
	if idx >= len(statuses) {
		return false
	}
	ic := statuses[idx]
	if ic.State.Running != nil || ic.State.Terminated != nil {
		return true
	}
	return false
}

// WaitForPodNameToBeRunning waits for the pod with the given name to be running
func WaitForPodNameToBeRunning(client kubernetes.Interface, namespace string, name string, timeout time.Duration) error {
	condition := func(pod *corev1.Pod) (bool, error) {
		return pod.Status.Phase == corev1.PodRunning, nil
	}
	return WaitforPodNameCondition(client, namespace, name, timeout, condition)
}

// WaitForPodNameToBeReady waits for the pod with the given name to become ready
func WaitForPodNameToBeReady(client kubernetes.Interface, namespace string, name string, timeout time.Duration) error {
	condition := func(pod *corev1.Pod) (bool, error) {
		return IsPodReady(pod), nil
	}
	return WaitforPodNameCondition(client, namespace, name, timeout, condition)
}

// WaitforPodNameCondition waits for the given pod name to match the given condition function
func WaitforPodNameCondition(client kubernetes.Interface, namespace string, name string, timeout time.Duration, condition func(pod *corev1.Pod) (bool, error)) error {
	options := metav1.ListOptions{
		// TODO
		//FieldSelector: fields.OneTermEqualSelector(api.ObjectNameField, name).String(),
		FieldSelector: fields.OneTermEqualSelector("metadata.name", name).String(),
	}

	// lets check if its already ready
	pod, err := client.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil && apierrors.IsNotFound(err) {
		err = nil
	}
	if err != nil {
		return errors.Wrapf(err, "failed to get pod %s in namespace %s", name, namespace)
	}
	if pod != nil {
		flag, err := condition(pod)
		if err != nil {
			return errors.Wrap(err, "failed to check pod condition")
		}
		if flag {
			return nil
		}
	}
	return waitForPodSelector(client, namespace, options, timeout, func(event watch.Event) (bool, error) {
		pod := event.Object.(*v1.Pod)
		if pod == nil {
			return false, err
		}
		return condition(pod)
	})
}

// WaitForPodSelectorToBeReady waits for the pod to become ready using the given selector name
func WaitForPodSelectorToBeReady(client kubernetes.Interface, namespace string, selector string, timeout time.Duration) (*corev1.Pod, error) {
	// lets check if its already ready
	opts := metav1.ListOptions{
		LabelSelector: selector,
	}
	pod, err := GetReadyPodForSelector(client, namespace, selector)
	if err != nil {
		return pod, errors.Wrapf(err, "failed to ")
	}
	statusMap := map[string]string{}
	logFailed := false
	condition := func(event watch.Event) (bool, error) {
		pod := event.Object.(*v1.Pod)
		status := PodStatus(pod)
		if statusMap[pod.Name] != status && !IsPodCompleted(pod) && pod.DeletionTimestamp == nil {
			log.Logger().Infof("pod %s has status %s", termcolor.ColorInfo(pod.Name), termcolor.ColorInfo(status))
			statusMap[pod.Name] = status
		}
		if IsPodFailed(pod) {
			if !logFailed {
				logFailed = true
			}
			cmdLine := fmt.Sprintf("kubectl log -n %s %s", namespace, pod.Name)
			log.Logger().Info()
			log.Logger().Warnf("the git operator pod has failed but will restart")
			log.Logger().Infof("to view the log of the failed git operator pod run: %s", termcolor.ColorInfo(cmdLine))
			log.Logger().Info()
		}
		return IsPodReady(pod), nil
	}
	err = waitForPodSelector(client, namespace, opts, timeout, condition)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to ")
	}
	return GetReadyPodForSelector(client, namespace, selector)
}

// GetReadyPodForSelector returns the first ready pod for the given selector or nil
func GetReadyPodForSelector(client kubernetes.Interface, namespace string, selector string) (*corev1.Pod, error) {
	// lets check if its already ready
	opts := metav1.ListOptions{
		LabelSelector: selector,
	}
	podList, err := client.CoreV1().Pods(namespace).List(context.TODO(), opts)
	if err != nil && apierrors.IsNotFound(err) {
		err = nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list pods in namespace %s with selector %s", namespace, selector)
	}
	if podList != nil {
		for i := range podList.Items {
			pod := &podList.Items[i]
			if IsPodReady(pod) {
				return pod, nil
			}
		}
	}
	return nil, nil
}

// WaitForPodNameToBeComplete waits for the pod to complete (succeed or fail) using the pod name
func WaitForPodNameToBeComplete(client kubernetes.Interface, namespace string, name string,
	timeout time.Duration) error {
	options := metav1.ListOptions{
		// TODO
		//FieldSelector: fields.OneTermEqualSelector(api.ObjectNameField, name).String(),
		FieldSelector: fields.OneTermEqualSelector("metadata.name", name).String(),
	}
	condition := func(event watch.Event) (bool, error) {
		pod := event.Object.(*v1.Pod)

		return IsPodCompleted(pod), nil
	}
	return waitForPodSelector(client, namespace, options, timeout, condition)
}

func GetPodNames(client kubernetes.Interface, ns string, filter string) ([]string, error) {
	var names []string
	list, err := client.CoreV1().Pods(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return names, fmt.Errorf("Failed to load Pods %s", err)
	}
	for _, d := range list.Items {
		name := d.Name
		if filter == "" || strings.Contains(name, filter) {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names, nil
}

func GetPods(client kubernetes.Interface, ns string, filter string) ([]string, map[string]*v1.Pod, error) {
	var names []string
	m := map[string]*v1.Pod{}
	list, err := client.CoreV1().Pods(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return names, m, fmt.Errorf("Failed to load Pods %s", err)
	}
	for _, d := range list.Items {
		c := d
		name := d.Name
		m[name] = &c
		if filter == "" || strings.Contains(name, filter) && d.DeletionTimestamp == nil {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names, m, nil
}

func GetPodsWithLabels(client kubernetes.Interface, ns string, selector string) ([]string, map[string]*v1.Pod, error) {
	var names []string
	m := map[string]*v1.Pod{}
	list, err := client.CoreV1().Pods(ns).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return names, m, fmt.Errorf("Failed to load Pods %s", err)
	}
	for _, d := range list.Items {
		c := d
		name := d.Name
		m[name] = &c
		if d.DeletionTimestamp == nil {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names, m, nil
}

// GetPodRestars returns the number of restarts of a POD
func GetPodRestarts(pod *v1.Pod) int32 {
	var restarts int32
	statuses := pod.Status.ContainerStatuses
	if len(statuses) == 0 {
		return restarts
	}
	for _, status := range statuses {
		restarts += status.RestartCount
	}
	return restarts
}

// GetContainersWithStatusAndIsInit gets the containers in the pod, either init containers or non-init depending on whether
// non-init containers are present, and a flag as to whether this list of containers are init containers or not.
func GetContainersWithStatusAndIsInit(pod *v1.Pod) ([]v1.Container, []v1.ContainerStatus, bool) {
	isInit := true
	allContainers := pod.Spec.InitContainers
	statuses := pod.Status.InitContainerStatuses
	containers := pod.Spec.Containers

	if len(containers) > 1 && len(pod.Status.ContainerStatuses) == len(containers) {
		isInit = false
		// Add the non-init containers
		// If there's a "nop" container at the end, the pod was created with before Tekton 0.5.x, so trim off the no-op container at the end of the list.
		if containers[len(containers)-1].Name == "nop" {
			allContainers = append(allContainers, containers[:len(containers)-1]...)
		} else {
			allContainers = append(allContainers, containers...)
		}
		// Since status ordering is unpredictable, don't trim here - we'll be sorting/filtering below anyway.
		statuses = append(statuses, pod.Status.ContainerStatuses...)
	}

	var sortedStatuses []v1.ContainerStatus
	for _, c := range allContainers {
		for _, cs := range statuses {
			if cs.Name == c.Name {
				sortedStatuses = append(sortedStatuses, cs)
				break
			}
		}
	}
	return allContainers, sortedStatuses, isInit
}
