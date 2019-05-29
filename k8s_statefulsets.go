package main

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListStatefulSets returns the list of StatefulSets
func (c *Client) ListStatefulSets(namespace string) (*appsv1.StatefulSetList, error) {
	statefulsets, err := c.clientset.AppsV1().StatefulSets(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve StatefulSets")
	}

	return statefulsets, nil
}

// DeleteStatefulSet deletes the given StatefulSet
func (c *Client) DeleteStatefulSet(statefulset appsv1.StatefulSet) error {
	if err := c.clientset.AppsV1().StatefulSets(statefulset.Namespace).Delete(statefulset.Name, &metav1.DeleteOptions{}); err != nil {
		return errors.Wrap(err, "failed to delete StatefulSet")
	}

	return nil
}

// StatefulSetsCleaner deletes all StatefulSets in k8s cluster (left slice) which are absent in VCS (right slice)
func (c *Client) StatefulSetsCleaner(namespace string, dryRun bool, directories []string) error {
	var left, right []string

	clusterStatefulsets, err := c.ListStatefulSets(namespace)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, value := range clusterStatefulsets.Items {
		left = append(left, value.Name)
	}

	directoryStatefulsets, err := CollectObjectsFromDir(directories)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, value := range directoryStatefulsets["StatefulSet"] {
		right = append(right, value)
	}

	objectsToDelete := Except(left, right)
	// Debug
	fmt.Println("**************************")
	fmt.Println(objectsToDelete)
	fmt.Println("**************************")

	for _, item := range objectsToDelete {
		for _, statefulset := range clusterStatefulsets.Items {
			if item == statefulset.Name {
				if dryRun {
					fmt.Printf("  Deleting StatefulSet %s [dry-run]\n", statefulset.Name)
				} else {
					fmt.Printf("  Deleting StatefulSet %s\n", statefulset.Name)
					if err := c.DeleteStatefulSet(statefulset); err != nil {
						fmt.Fprintln(os.Stderr, err)
						os.Exit(1)
					}
				}
			}
		}
	}
	return nil
}
