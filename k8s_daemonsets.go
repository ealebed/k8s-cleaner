package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
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

// DaemonSetsCleaner deletes all DaemonSets in k8s cluster which are absent in VCS
func (c *Client) DaemonSetsCleaner(namespace string, dryRun bool, directories []string) error {
	var left []string

	clusterDaemonsets, err := c.ListDaemonSets(namespace)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, value := range clusterDaemonsets.Items {
		left = append(left, value.Name)
	}

	objectsToDelete := Except(left, "DaemonSet", directories)

	for _, item := range objectsToDelete {
		for _, daemonset := range clusterDaemonsets.Items {
			if item == daemonset.Name {
				if dryRun {
					color.Yellow("******************************************************************************")
					color.Yellow("  Deleting DaemonSet %s [dry-run]\n", daemonset.Name)
					color.Yellow("******************************************************************************")
				} else {
					color.Red("******************************************************************************")
					color.Red("  Deleting DaemonSet %s\n", daemonset.Name)
					color.Red("******************************************************************************")
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
