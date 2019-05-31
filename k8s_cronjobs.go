package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	v1beta1 "k8s.io/api/batch/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListCronJobs returns the list of CronJobs
func (c *Client) ListCronJobs(namespace string) (*v1beta1.CronJobList, error) {
	cronjobs, err := c.clientset.BatchV1beta1().CronJobs(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve CronJobs")
	}

	return cronjobs, nil
}

// DeleteCronJob deletes the given CronJob
func (c *Client) DeleteCronJob(cronjob v1beta1.CronJob) error {
	if err := c.clientset.BatchV1beta1().CronJobs(cronjob.Namespace).Delete(cronjob.Name, &metav1.DeleteOptions{}); err != nil {
		return errors.Wrap(err, "failed to delete CronJob")
	}

	return nil
}

// CronJobsCleaner deletes all CronJobs in k8s cluster which are absent in VCS
func (c *Client) CronJobsCleaner(namespace string, dryRun bool, directories []string) error {
	var left []string

	clusterCronjobs, err := c.ListCronJobs(namespace)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, value := range clusterCronjobs.Items {
		if value.Name == "cert-manager-webhook-ca-sync" {
			color.Red("You can't delete cronjob %s", value.Name)
			continue
		} else {
			left = append(left, value.Name)
		}
	}

	objectsToDelete := Except(left, "CronJob", directories)

	for _, item := range objectsToDelete {
		for _, cronjob := range clusterCronjobs.Items {
			if item == cronjob.Name {
				if dryRun {
					color.Yellow("******************************************************************************")
					color.Yellow("  Deleting CronJob %s [dry-run]\n", cronjob.Name)
					color.Yellow("******************************************************************************")
				} else {
					color.Red("******************************************************************************")
					color.Red("  Deleting CronJob %s\n", cronjob.Name)
					color.Red("******************************************************************************")
					if err := c.DeleteCronJob(cronjob); err != nil {
						fmt.Fprintln(os.Stderr, err)
						os.Exit(1)
					}
				}
			}
		}
	}

	return nil
}
