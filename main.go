package main

import (
	"fmt"
	"os"

	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	flag "github.com/spf13/pflag"
)

// Client represents the wrapper of Kubernetes API client
type Client struct {
	clientConfig clientcmd.ClientConfig
	clientset    kubernetes.Interface
}

// NewClient creates Client object using local kubecfg
func NewClient(kubeconfig, context string) (*Client, error) {
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
		&clientcmd.ConfigOverrides{CurrentContext: context})

	config, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, errors.Wrap(err, "falied to load local kubeconfig")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load clientset")
	}

	return &Client{
		clientConfig: clientConfig,
		clientset:    clientset,
	}, nil
}

// NewClientInCluster creates Client object in Kubernetes cluster
func NewClientInCluster() (*Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to load kubeconfig in cluster")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "falied to load clientset")
	}

	return &Client{
		clientset: clientset,
	}, nil
}

// ListPods returns the list of Pods
func (c *Client) ListPods(namespace string) (*corev1.PodList, error) {
	pods, err := c.clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	//pods, err := c.clientset.Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve Pods")
	}

	return pods, nil
}

// NamespaceInConfig returns namespace set in kubeconfig
func (c *Client) NamespaceInConfig() (string, error) {
	if c.clientConfig == nil {
		return "", errors.New("clientConfig is not set")
	}

	rawConfig, err := c.clientConfig.RawConfig()
	if err != nil {
		return "", errors.Wrap(err, "failed to load rawConfig")
	}

	return rawConfig.Contexts[rawConfig.CurrentContext].Namespace, nil
}

func main() {
	var (
		context    string
		dryRun     bool
		kubeconfig string
		namespace  string
	)

	flags := flag.NewFlagSet("k8stail", flag.ExitOnError)
	flags.Usage = func() {
		flags.PrintDefaults()
	}

	flags.StringVar(&context, "context", "", "Kubernetes context")
	flags.BoolVar(&dryRun, "dry-run", false, "Dry run")
	flags.StringVar(&kubeconfig, "kubeconfig", "", "Path of kubeconfig")
	flags.StringVar(&namespace, "namespace", "", "Kubernetes namespace")

	if err := flags.Parse(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
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

	if namespace == "" {
		namespaceInConfig, err := c.NamespaceInConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if namespaceInConfig == "" {
			namespace = metav1.NamespaceAll
		} else {
			namespace = namespaceInConfig
		}
	}
	client = c

	pods, err := client.ListPods(namespace)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	for _, pod := range pods.Items {
		fmt.Printf("   Pod %s\n", pod.Name)
	}
		
	if dryRun {
		fmt.Printf("DRY_RUN enabled ... [dry-run]\n")
	}
}
