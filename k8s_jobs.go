package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListJobs returns the list of Jobs
func (c *Client) ListJobs(namespace string) (*batchv1.JobList, error) {
	jobs, err := c.clientset.BatchV1().Jobs(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve Jobs")
	}

	return jobs, nil
}

// DeleteJob deletes the given Job
func (c *Client) DeleteJob(job batchv1.Job) error {
	if err := c.clientset.BatchV1().Jobs(job.Namespace).Delete(job.Name, &metav1.DeleteOptions{}); err != nil {
		return errors.Wrap(err, "failed to delete Job")
	}

	return nil
}

// IsJobFinished returns whether the given Job has finished or not
func IsJobFinished(job batchv1.Job) bool {
	return job.Status.Succeeded > 0
}

// Jobs represents job list
type Jobs []batchv1.Job

// Len return the length of job list
func (j Jobs) Len() int {
	return len(j)
}

// Less returns whether the former item is less than the latter item or not
func (j Jobs) Less(m, n int) bool {
	if j[m].Status.CompletionTime == nil {
		return false
	}

	if j[n].Status.CompletionTime == nil {
		return true
	}
	return j[m].Status.CompletionTime.Before(j[n].Status.CompletionTime)
}

// Swap swaps two items
func (j Jobs) Swap(m, n int) {
	j[m], j[n] = j[n], j[m]
}

// JobAndPodCleaner deletes completed Jobs and attached Pods
func (c *Client) JobAndPodCleaner(namespace string, maxCount int64, dryRun bool) error {

	jobs, err := c.ListJobs(namespace)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	jobGroup := map[string]Jobs{}

	for _, job := range jobs.Items {
		if !IsJobFinished(job) {
			continue
		}

		label := job.Labels["jobgroup"]

		if label == "" {
			continue
		}

		if jobGroup[label] == nil {
			jobGroup[label] = Jobs{}
		}

		jobGroup[label] = append(jobGroup[label], job)
	}

	pods, err := c.ListPods(namespace)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	podGroup := map[string]Pods{}

	for _, pod := range pods.Items {
		if !IsPodFinished(pod) {
			continue
		}

		label := pod.Labels[jobNameLabel]

		if label == "" {
			continue
		}

		if podGroup[label] == nil {
			podGroup[label] = Pods{}
		}

		podGroup[label] = append(podGroup[label], pod)
	}

	for _, jobs := range jobGroup {
		i := int64(0)
		sort.Sort(sort.Reverse(jobs))

		for _, job := range jobs {
			if i < maxCount {
				i++
				continue
			}

			if dryRun {
				color.Yellow("******************************************************************************")
				color.Yellow("Deleting Job %s  [dry-run]\n", job.Name)
				color.Yellow("******************************************************************************")
			} else {
				color.Red("******************************************************************************")
				color.Red("Deleting Job %s \n", job.Name)
				color.Red("******************************************************************************")
				if err := c.DeleteJob(job); err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
			}

			for _, pod := range podGroup[job.Name] {
				if dryRun {
					color.Yellow("******************************************************************************")
					color.Yellow("  Deleting Pod %s [dry-run]\n", pod.Name)
					color.Yellow("******************************************************************************")
				} else {
					color.Red("******************************************************************************")
					color.Red("  Deleting Pod %s\n", pod.Name)
					color.Red("******************************************************************************")
					if err := c.DeletePod(pod); err != nil {
						fmt.Fprintln(os.Stderr, err)
						os.Exit(1)
					}
				}
			}
		}
	}

	return nil
}
