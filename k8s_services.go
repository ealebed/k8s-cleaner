package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListServices returns the list of Services
func (c *Client) ListServices(namespace string) (*corev1.ServiceList, error) {
	services, err := c.clientset.CoreV1().Services(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve Services")
	}

	return services, nil
}

// DeleteService deletes the given Service
func (c *Client) DeleteService(service corev1.Service) error {
	if err := c.clientset.CoreV1().Services(service.Namespace).Delete(service.Name, &metav1.DeleteOptions{}); err != nil {
		return errors.Wrap(err, "failed to delete Service")
	}

	return nil
}

// ServicesCleaner deletes all Services in k8s cluster which are absent in VCS
func (c *Client) ServicesCleaner(namespace string, dryRun bool, directories []string) error {
	var left []string

	// Get service list from cluster
	clusterServices, err := c.ListServices(namespace)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Put services names (except "kubernetes") to left slice for future comparing
	for _, value := range clusterServices.Items {
		if value.Name == "kubernetes" {
			color.Red("You can't delete service %s", value.Name)
			continue
		} else {
			left = append(left, value.Name)
		}
	}

	// Get slice with services in k8s cluster which are absent in VCS
	objectsToDelete := Except(left, "Service", directories)

	// Delete services absent in VCS
	for _, item := range objectsToDelete {
		for _, service := range clusterServices.Items {
			if item == service.Name {
				if dryRun {
					color.Yellow("******************************************************************************")
					color.Yellow("  Deleting Service %s [dry-run]\n", service.Name)
					color.Yellow("******************************************************************************")
				} else {
					color.Red("******************************************************************************")
					color.Red("  Deleting Service %s\n", service.Name)
					color.Red("******************************************************************************")
					if err := c.DeleteService(service); err != nil {
						fmt.Fprintln(os.Stderr, err)
						os.Exit(1)
					}
				}
			}
		}
	}

	return nil
}
