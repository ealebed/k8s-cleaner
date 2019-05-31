package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
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

// LimitRangesCleaner deletes all LimitRanges in k8s cluster which are absent in VCS
func (c *Client) LimitRangesCleaner(namespace string, dryRun bool, directories []string) error {
	var left []string

	clusterLimitranges, err := c.ListLimitRanges(namespace)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, value := range clusterLimitranges.Items {
		if value.Name == "limits" {
			color.Red("You can't delete LimitRange %s", value.Name)
			continue
		} else {
			left = append(left, value.Name)
		}
	}

	objectsToDelete := Except(left, "LimitRange", directories)

	for _, item := range objectsToDelete {
		for _, limitrange := range clusterLimitranges.Items {
			if item == limitrange.Name {
				if dryRun {
					color.Yellow("******************************************************************************")
					color.Yellow("  Deleting LimitRange %s [dry-run]\n", limitrange.Name)
					color.Yellow("******************************************************************************")
				} else {
					color.Red("******************************************************************************")
					color.Red("  Deleting LimitRange %s\n", limitrange.Name)
					color.Red("******************************************************************************")
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
