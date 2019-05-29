package main

import (
	"fmt"
	"log"
	"os"

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

// ServicesCleaner deletes all Services in k8s cluster (left slice) which are absent in VCS (right slice)
func (c *Client) ServicesCleaner(namespace string, dryRun bool, directories []string) error {
	var left, right []string

	clusterServices, err := c.ListServices(namespace)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, value := range clusterServices.Items {
		if value.Name == "kubernetes" {
			log.Printf("You can't delete service %s", value.Name)
			continue
		} else {
			left = append(left, value.Name)
		}
	}

	directoryServices, err := CollectObjectsFromDir(directories)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, value := range directoryServices["Service"] {
		right = append(right, value)
	}

	objectsToDelete := Except(left, right)
	// Debug
	fmt.Println("**************************")
	fmt.Println(objectsToDelete)
	fmt.Println("**************************")

	for _, item := range objectsToDelete {
		for _, service := range clusterServices.Items {
			if item == service.Name {
				if dryRun {
					fmt.Printf("  Deleting Service %s [dry-run]\n", service.Name)
				} else {
					fmt.Printf("  Deleting Service %s\n", service.Name)
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
