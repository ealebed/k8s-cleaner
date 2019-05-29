package main

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListDaemonSets returns the list of DaemonSets
func (c *Client) ListDaemonSets(namespace string) (*appsv1.DaemonSetList, error) {
	daemonsets, err := c.clientset.AppsV1().DaemonSets(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve DaemonSets")
	}

	return daemonsets, nil
}

// DeleteDaemonSet deletes the given DaemonSet
func (c *Client) DeleteDaemonSet(daemonset appsv1.DaemonSet) error {
	if err := c.clientset.AppsV1().DaemonSets(daemonset.Namespace).Delete(daemonset.Name, &metav1.DeleteOptions{}); err != nil {
		return errors.Wrap(err, "failed to delete DaemonSet")
	}

	return nil
}

// DaemonSetsCleaner deletes all DaemonSets in k8s cluster (left slice) which are absent in VCS (right slice)
func (c *Client) DaemonSetsCleaner(namespace string, dryRun bool, directories []string) error {
	var left, right []string

	clusterDaemonsets, err := c.ListDaemonSets(namespace)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, value := range clusterDaemonsets.Items {
		left = append(left, value.Name)
	}

	directoryDaemonsets, err := CollectObjectsFromDir(directories)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, value := range directoryDaemonsets["DaemonSet"] {
		right = append(right, value)
	}

	objectsToDelete := Except(left, right)
	// Debug
	fmt.Println("**************************")
	fmt.Println(objectsToDelete)
	fmt.Println("**************************")

	for _, item := range objectsToDelete {
		for _, daemonset := range clusterDaemonsets.Items {
			if item == daemonset.Name {
				if dryRun {
					fmt.Printf("  Deleting DaemonSet %s [dry-run]\n", daemonset.Name)
				} else {
					fmt.Printf("  Deleting DaemonSet %s\n", daemonset.Name)
					if err := c.DeleteDaemonSet(daemonset); err != nil {
						fmt.Fprintln(os.Stderr, err)
						os.Exit(1)
					}
				}
			}
		}
	}
	return nil
}
