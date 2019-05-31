package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
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

// DeploymentsCleaner deletes all Deployments in k8s cluster which are absent in VCS
func (c *Client) DeploymentsCleaner(namespace string, dryRun bool, directories []string) error {
	var left []string

	clusterDeployments, err := c.ListDeployments(namespace)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, value := range clusterDeployments.Items {
		left = append(left, value.Name)
	}

	objectsToDelete := Except(left, "Deployment", directories)

	for _, item := range objectsToDelete {
		for _, deployment := range clusterDeployments.Items {
			if item == deployment.Name {
				if dryRun {
					color.Yellow("******************************************************************************")
					color.Yellow("  Deleting Deployment %s [dry-run]\n", deployment.Name)
					color.Yellow("******************************************************************************")
				} else {
					color.Red("******************************************************************************")
					color.Red("  Deleting Deployment %s\n", deployment.Name)
					color.Red("******************************************************************************")
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
