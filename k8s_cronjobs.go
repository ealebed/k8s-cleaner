package main

import (
	"fmt"
	"os"

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

// CronJobsCleaner deletes all CronJobs in k8s cluster (left slice) which are absent in VCS (right slice)
func (c *Client) CronJobsCleaner(namespace string, dryRun bool, directories []string) error {
	var left, right []string

	clusterCronjobs, err := c.ListCronJobs(namespace)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, value := range clusterCronjobs.Items {
		left = append(left, value.Name)
	}

	directoryCronjobs, err := CollectObjectsFromDir(directories)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, value := range directoryCronjobs["CronJob"] {
		right = append(right, value)
	}

	objectsToDelete := Except(left, right)
	// Debug
	fmt.Println("**************************")
	fmt.Println(objectsToDelete)
	fmt.Println("**************************")

	for _, item := range objectsToDelete {
		for _, cronjob := range clusterCronjobs.Items {
			if item == cronjob.Name {
				if dryRun {
					fmt.Printf("  Deleting CronJob %s [dry-run]\n", cronjob.Name)
				} else {
					fmt.Printf("  Deleting CronJob %s\n", cronjob.Name)
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
