# k8s-cleaner

Command-line tool for comparing sets of Kubernetes objects. It reads object definitions from a running cluster and performs a comparsion to another source (local directory) of object definitions. At this moment, k8s-cleaner supports next kubernetes kinds: (Service|StatefulSet|Deployment|CronJob|LimitRange|DaemonSet).
If object is present in running cluster, but absent in local directory tool will delete it. Also, completed Jobs and attached Pods (except last `maxCount`) can be deleted.

## Installation

### From source

```bash
$ go build -v ./
```

## Usage

```bash
$ k8s-cleaner --context=my-k8s-test-cluster --namespace=default --kind=Deployment --directories=${pwd}/manifests/dir1/,/full/path/to/manifests/dir2/ --dry-run=false
```

`k8s-cleaner` uses `~/.kube/config` as default. You can specify another path by `KUBECONFIG` environment variable or `--kubeconfig` option. `--kubeconfig` option always overrides `KUBECONFIG` environment variable.

```bash
$ KUBECONFIG=/path/to/kubeconfig k8s-cleaner
# or
$ k8s-cleaner --kubeconfig=/path/to/kubeconfig
```

### Options

|Option|Description|Required|Default|
|---------|-----------|-------|-------|
|`--kubeconfig=KUBECONFIG`|Path of kubeconfig||`~/.kube/config`|
|`--context=CONTEXT`|Kubernetes context||current context|
|`--namespace=NAMESPACE`|Kubernetes namespace||`default`|
|`--kind=KIND`|Kubernetes kind (only supported)||`All`|
|`--dry-run`|Dry run||`true`|
|`--max-count`|Number of Jobs to remain (only if selected kind is Jobs)||`10`|
|`--directories`|Paths to directories with manifests (separated by commas)|yes|`nil`|
