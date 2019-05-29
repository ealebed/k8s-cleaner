package main

import (
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	defaultMaxCount  = 10
	acceptedK8sKinds = `(Service|StatefulSet|Deployment|CronJob|LimitRange|DaemonSet)`
)

func main() {
	var (
		context    string
		dryRun     bool
		kubeconfig string
		maxCount   int64
		namespace  string
		directory  string
	)

	flags := flag.NewFlagSet("k8stail", flag.ExitOnError)
	flags.Usage = func() {
		flags.PrintDefaults()
	}

	flags.StringVar(&context, "context", "", "Kubernetes context")
	flags.BoolVar(&dryRun, "dry-run", true, "Dry run")
	flags.StringVar(&kubeconfig, "kubeconfig", "", "Path of kubeconfig")
	flags.Int64Var(&maxCount, "max-count", int64(defaultMaxCount), "Number of Jobs to remain")
	flags.StringVar(&namespace, "namespace", "", "Kubernetes namespace")
	flags.StringVar(&directory, "directory", "", "Path to directory with manifests (_commons will be added automatically)")

	if err := flags.Parse(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if directory == "" {
		fmt.Fprintln(os.Stderr, "--directory must be set")
		os.Exit(1)
	}

	directories := []string{"/Users/ealebed/Code/loopme/k8s/datacenters/_commons"}
	directories = append(directories, directory)

	if kubeconfig == "" {
		if os.Getenv("KUBECONFIG") != "" {
			kubeconfig = os.Getenv("KUBECONFIG")
		} else {
			kubeconfig = clientcmd.RecommendedHomeFile
		}
	}

	var client *Client

	c, err := NewClient(kubeconfig, context)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	client = c

	if namespace == "kube-system" {
		fmt.Println("  !!! You can't manage this namespace")
		os.Exit(1)
	}
	if namespace == "" {
		namespaceInConfig, err := c.NamespaceInConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if namespaceInConfig == "" {
			namespace = "default"
		} else {
			namespace = namespaceInConfig
		}
	}

	client.DeploymentsCleaner(namespace, dryRun, directories)
	client.ServicesCleaner(namespace, dryRun, directories)
	client.CronJobsCleaner(namespace, dryRun, directories)
	client.StatefulSetsCleaner(namespace, dryRun, directories)
	client.DaemonSetsCleaner(namespace, dryRun, directories)
	client.LimitRangesCleaner(namespace, dryRun, directories)
	client.JobAndPodCleaner(namespace, maxCount, dryRun)

}
