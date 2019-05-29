package main

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListDeployments returns the list of Deployments
func (c *Client) ListDeployments(namespace string) (*appsv1.DeploymentList, error) {
	deployments, err := c.clientset.AppsV1().Deployments(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve Deployments")
	}

	return deployments, nil
}

// DeleteDeployment deletes the given Deployment
func (c *Client) DeleteDeployment(deployment appsv1.Deployment) error {
	if err := c.clientset.AppsV1().Deployments(deployment.Namespace).Delete(deployment.Name, &metav1.DeleteOptions{}); err != nil {
		return errors.Wrap(err, "failed to delete Deployment")
	}

	return nil
}

// DeploymentsCleaner deletes all Deployments in k8s cluster (left slice) which are absent in VCS (right slice)
func (c *Client) DeploymentsCleaner(namespace string, dryRun bool, directories []string) error {
	var left, right []string

	clusterDeployments, err := c.ListDeployments(namespace)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, value := range clusterDeployments.Items {
		left = append(left, value.Name)
	}

	directoryDeployments, err := CollectObjectsFromDir(directories)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, value := range directoryDeployments["Deployment"] {
		right = append(right, value)
	}

	objectsToDelete := Except(left, right)
	// Debug
	fmt.Println("**************************")
	fmt.Println(objectsToDelete)
	fmt.Println("**************************")

	for _, item := range objectsToDelete {
		for _, deployment := range clusterDeployments.Items {
			if item == deployment.Name {
				if dryRun {
					fmt.Printf("  Deleting Deployment %s [dry-run]\n", deployment.Name)
				} else {
					fmt.Printf("  Deleting Deployment %s\n", deployment.Name)
					if err := c.DeleteDeployment(deployment); err != nil {
						fmt.Fprintln(os.Stderr, err)
						os.Exit(1)
					}
				}
			}
		}
	}
	return nil
}
