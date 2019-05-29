package main

import (
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	jobNameLabel = "job-name"
)

// ListPods returns the list of Pods
func (c *Client) ListPods(namespace string) (*corev1.PodList, error) {
	pods, err := c.clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve Pods")
	}

	return pods, nil
}

// DeletePod deletes the given Pod
func (c *Client) DeletePod(pod corev1.Pod) error {
	if err := c.clientset.CoreV1().Pods(pod.Namespace).Delete(pod.Name, &metav1.DeleteOptions{}); err != nil {
		return errors.Wrap(err, "failed to delete Pod")
	}

	return nil
}

// IsPodFinished returns whether the given Pod has finished or not
func IsPodFinished(pod corev1.Pod) bool {
	return pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed
}

// Pods represents pod list
// Sorting Pods is not necessary
type Pods []corev1.Pod
