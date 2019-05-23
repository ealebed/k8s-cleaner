# k8s-cleaner

Command-line tool for comparing sets of Kubernetes objects. It reads object definitions from a running cluster and performs a comparison to another source (local directory) of object definitions.

## Requirements

Kubernetes 1.3 or above

## Installation

### From source

```bash
$ go get -d github.com/ealebed/k8s-cleaner
$ cd $GOPATH/src/github.com/ealebed/k8s-cleaner
$ go get -d -v
$ go build -v ./
```

## Usage

```bash
$ k8s-cleaner --context=test-k8s-cluster --namespace=my-test-ns
```

### Local machine

`k8s-cleaner` uses `~/.kube/config` as default. You can specify another path by `KUBECONFIG` environment variable or `--kubeconfig` option. `--kubeconfig` option always overrides `KUBECONFIG` environment variable.

```bash
$ KUBECONFIG=/path/to/kubeconfig k8s-cleaner
# or
$ k8s-cleaner --kubeconfig=/path/to/kubeconfig
```

### Options

|Option|Description|Required|Default|
|---------|-----------|-------|-------|
|`--context=CONTEXT`|Kubernetes context|||
|`--dry-run`|Dry run||`false`|
|`--kubeconfig=KUBECONFIG`|Path of kubeconfig||`~/.kube/config`|
|`--namespace=NAMESPACE`|Kubernetes namespace||All namespaces|
|`-h`, `-help`|Print command line usage|||

