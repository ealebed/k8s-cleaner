package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	flag "github.com/spf13/pflag"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	defaultMaxCount  = 10
	defaultNamespace = "default"
	acceptedK8sKinds = `(Service|StatefulSet|Deployment|CronJob|LimitRange|DaemonSet)`
	defaultKind      = "All"
	debug            = false
)

func main() {
	var (
		kubeconfig           string
		context              string
		namespace            string
		kind                 string
		maxCount             int64
		dryRun               bool
		restrictedNamespaces = []string{"kube-system", "kube-public", "kube-node-lease", "spinnaker"}
	)

	flags := flag.NewFlagSet("k8stail", flag.ExitOnError)
	flags.Usage = func() {
		flags.PrintDefaults()
	}

	flags.StringVar(&kubeconfig, "kubeconfig", "", "Path of kubeconfig")
	flags.StringVar(&context, "context", "", "Kubernetes context")
	flags.StringVar(&namespace, "namespace", string(defaultNamespace), "Kubernetes namespace")
	flags.StringVar(&kind, "kind", string(defaultKind), "Kubernetes kind for cleaning. Can be one of Service|StatefulSet|Deployment|CronJob|LimitRange|DaemonSet|Jobs or All")
	flags.BoolVar(&dryRun, "dry-run", true, "Dry run")
	flags.Int64Var(&maxCount, "max-count", int64(defaultMaxCount), "Number of Jobs to remain, only if selected kind is Jobs")
	flags.StringSlice("directories", nil, "Paths to directories with manifests separated by commas")

	if err := flags.Parse(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	dirs, err := flags.GetStringSlice("directories")
	if len(dirs) == 0 {
		color.Red("No directories for analyze, exit")
		os.Exit(1)
	}

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

	if stringInSlice(namespace, restrictedNamespaces) {
		color.Red("You can't manage namespace %s\n", namespace)
		os.Exit(1)
	}
	if namespace == "" {
		namespaceInConfig, err := c.NamespaceInConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if namespaceInConfig == "" {
			namespace = defaultNamespace
		} else {
			namespace = namespaceInConfig
		}
	}

	switch kind {
	case "Service":
		client.ServicesCleaner(namespace, dryRun, dirs)
	case "StatefulSet":
		client.StatefulSetsCleaner(namespace, dryRun, dirs)
	case "Deployment":
		client.DeploymentsCleaner(namespace, dryRun, dirs)
	case "CronJob":
		client.CronJobsCleaner(namespace, dryRun, dirs)
	case "LimitRange":
		client.LimitRangesCleaner(namespace, dryRun, dirs)
	case "DaemonSet":
		client.DaemonSetsCleaner(namespace, dryRun, dirs)
	case "Jobs":
		client.JobAndPodCleaner(namespace, maxCount, dryRun)
	case "All":
		client.DeploymentsCleaner(namespace, dryRun, dirs)
		client.ServicesCleaner(namespace, dryRun, dirs)
		client.CronJobsCleaner(namespace, dryRun, dirs)
		client.StatefulSetsCleaner(namespace, dryRun, dirs)
		client.DaemonSetsCleaner(namespace, dryRun, dirs)
		client.LimitRangesCleaner(namespace, dryRun, dirs)
		client.JobAndPodCleaner(namespace, maxCount, dryRun)
	}
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
