package main

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListLimitRanges returns the list of LimitRanges
func (c *Client) ListLimitRanges(namespace string) (*corev1.LimitRangeList, error) {
	limitranges, err := c.clientset.CoreV1().LimitRanges(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve LimitRanges")
	}

	return limitranges, nil
}

// DeleteLimitRange deletes the given LimitRange
func (c *Client) DeleteLimitRange(limitrange corev1.LimitRange) error {
	if err := c.clientset.CoreV1().LimitRanges(limitrange.Namespace).Delete(limitrange.Name, &metav1.DeleteOptions{}); err != nil {
		return errors.Wrap(err, "failed to delete LimitRange")
	}

	return nil
}

// LimitRangesCleaner deletes all LimitRanges in k8s cluster (left slice) which are absent in VCS (right slice)
func (c *Client) LimitRangesCleaner(namespace string, dryRun bool, directories []string) error {
	var left, right []string

	clusterLimitranges, err := c.ListLimitRanges(namespace)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, value := range clusterLimitranges.Items {
		left = append(left, value.Name)
	}

	directoryLimitranges, err := CollectObjectsFromDir(directories)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, value := range directoryLimitranges["LimitRange"] {
		right = append(right, value)
	}

	objectsToDelete := Except(left, right)
	// Debug
	fmt.Println("**************************")
	fmt.Println(objectsToDelete)
	fmt.Println("**************************")

	for _, item := range objectsToDelete {
		for _, limitrange := range clusterLimitranges.Items {
			if item == limitrange.Name {
				if dryRun {
					fmt.Printf("  Deleting LimitRange %s [dry-run]\n", limitrange.Name)
				} else {
					fmt.Printf("  Deleting LimitRange %s\n", limitrange.Name)
					if err := c.DeleteLimitRange(limitrange); err != nil {
						fmt.Fprintln(os.Stderr, err)
						os.Exit(1)
					}
				}
			}
		}
	}
	return nil
}
