# Athens Proxy Helm Chart

## What is Athens?

[Athens](https://docs.gomods.io) is a repository for packages used by your go packages.

Athens provides a repository for [Go Modules](https://github.com/golang/go/wiki/Modules) that you can run. It serves public code and your private code for you, so you don't have to pull directly from a version control system (VCS) like GitHub or GitLab.

## Prerequisites

* Kubernetes 1.10+

## Requirements

- A running Kubernetes cluster
- [Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) installed and setup to use the cluster
- [Helm](https://helm.sh/) [installed](https://github.com/helm/helm#install) and setup to use the cluster (helm init) or [Tillerless Helm](https://github.com/rimusz/helm-tiller)

## Deploy Athens

The fastest way to install Athens using Helm is to deploy it from our public Helm chart repository. First, add the repository with this command:

```console
$ helm repo add gomods https://athens.blob.core.windows.net/charts
$ helm repo update
```

Next, install the chart with default values to `athens` namespace:  

```
$ helm install gomods/athens-proxy -n athens --namespace athens
```

This will deploy a single Athens instance in the `athens` namespace with `disk` storage enabled. Additionally, a `ClusterIP` service will be created.


## Advanced Configuration

For more advanced configuration options please check Athens [docs](https://docs.gomods.io/install/install-on-kubernetes/#advanced-configuration).

Available options:
- [Replicas](https://docs.gomods.io/install/install-on-kubernetes/#replicas)
- [Access to private repositories via Github](https://docs.gomods.io/install/install-on-kubernetes/#give-athens-access-to-private-repositories-via-github-token-optional)
- [Storage Providers](https://docs.gomods.io/install/install-on-kubernetes/#storage-providers)
- [Kubernetes Service](https://docs.gomods.io/install/install-on-kubernetes/#kubernetes-service)
- [Ingress Resource](https://docs.gomods.io/install/install-on-kubernetes/#ingress-resource)
- [Upstream module repository](https://docs.gomods.io/install/install-on-kubernetes/#upstream-module-repository)
- [.netrc file support](https://docs.gomods.io/install/install-on-kubernetes/#netrc-file-support)
