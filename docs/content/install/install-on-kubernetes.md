---
title: Install Athens on Kubernetes
description: Installing an Athens Instance on Kubernetes
weight: 1
---

When you follow the instructions in the [Walkthrough](/walkthrough), you end up with an Athens Proxy that uses in-memory storage. This is only suitable for trying out the Athens proxy for a short period of time, as you will quickly run out of memory and Athens won't persist modules between restarts. In order to run a more production-like proxy, you may with to run Athens on a [Kubernetes](https://kubernetes.io/) cluster. To aid in deployment of the Athens proxy on Kubernetes, a [Helm](https://www.helm.sh/) chart has been provided. This guide will walk you through installing Athens on a Kubernetes cluster using Helm.

* [Prerequisites](#prerequisites)
* [Configure Helm](#configure-helm)
* [Deploy Athens](#deploy-athens)

---

## Prerequisites

In order to install Athens on your Kubernetes cluster, there are a few prerequisites that you must satisfy. If you already have completed the following steps, please continue to [configuring helm](#configure-helm). This guide assumes you have already created a Kubernetes cluster.

* Install the [Kubernetes CLI](#install-the-kubernetes-cli).
* Install the [Helm CLI](#install-the-helm-cli).

### Install the Kubernetes CLI

In order to interact with your Kubernetes Cluster, you will need to [install kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/).

### Install The Helm CLI

[Helm](https://github.com/kubernetes/helm) is a tool for installing pre-configured applications on Kubernetes.
Install `helm` by running the following command:

#### MacOS

```console
brew install kubernetes-helm
```

#### Windows

1. Download the latest [Helm release](https://storage.googleapis.com/kubernetes-helm/helm-v2.7.2-windows-amd64.tar.gz).
1. Decompress the tar file.
1. Copy **helm.exe** to a directory on your PATH.

#### Linux

```console
curl https://raw.githubusercontent.com/kubernetes/helm/master/scripts/get | bash
```

## Configure Helm

If your cluster has already been configured to use Helm, please continue to [deploy Athens](#deploy-athens).

If not, please read on.

### RBAC Cluster

If your cluster has RBAC enabled, you will need to create a ServiceAccount, ClusterRole and ClusterRoleBinding for Helm to use. The following command will create these and initialize Helm.

```console
kubectl create -f https://raw.githubusercontent.com/Azure/helm-charts/master/docs/prerequisities/helm-rbac-config.yaml
helm init --service-account tiller
```

### Non RBAC Cluster

If your cluster has does not have RBAC enabled, you can simply initialize Helm.

```console
helm init
```

Before deploying Athens, you will need to wait for the Tiller pod to become `Ready`. You can check the status by watching the pods in `kube-system`:

```console
$ kubectl get pods -n kube-system -w
NAME                                    READY     STATUS    RESTARTS   AGE
tiller-deploy-5456568744-76c6s          1/1       Running   0          5s
```

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

This will deploy a single Athens instance in the `default` namespace with `disk` storage enabled. Additionally, a `ClusterIP` service will be created.

By default, the chart will install Athens with a replica count of 1. To change this, change the `replicaCount` value:

```console
helm install gomods/athens-proxy -n athens --namespace athens --set replicaCount=3
```

## Advanced Configuration

### Replicas

By default, the chart will install Athens with a replica count of 1. To change this, change the `replicaCount` value:

```console
helm install gomods/athens-proxy -n athens --namespace athens --set replicaCount=3
```

### Give Athens access to private repositories via Github Token (Optional)

1. Create a token at https://github.com/settings/tokens
2. Provide the token to the Athens proxy either through the [config.toml](https://github.com/gomods/athens/blob/master/config.dev.toml) file (the `GithubToken` field) or by setting the `ATHENS_GITHUB_TOKEN` environment variable.

### Storage Providers

The Helm chart currently supports running Athens with two different storage providers: `disk` and `mongo`. The default behavior is to use the `disk` storage provider.

#### Disk Storage Configuration

When using the `disk` storage provider, you can configure a number of options regarding data persistence. By default, Athens will deploy using an `emptyDir` volume. This probably isn't sufficient for production use cases, so the chart also allows you to configure persistence via a [PersistentVolumeClaim](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims). The chart currently allows you to set the following values:

```yaml
persistence:
  enabled: false
  accessMode: ReadWriteOnce
  size: 4Gi
  storageClass:
```

Add it to `override-values.yaml` file and run:

```console
helm install gomods/athens-proxy -n athens --namespace athens -f override-values.yaml
```

`enabled` is used to turn on the PVC feature of the chart, while the other values relate directly to the values defined in the PersistentVolumeClaim documentation.

#### Mongo DB Configuration

To use the Mongo DB storage provider, you will first need a MongoDB instance. Once you have deployed MongoDB, you can configure Athens using the connection string via `storage.mongo.url`. You will also need to set `storage.type` to "mongo".

```
helm install gomods/athens-proxy -n athens --namespace athens --set storage.type=mongo --set storage.mongo.url=<some-mongodb-connection-string>
```

### Kubernetes Service

By default, a Kubernetes `ClusterIP` service is created for the Athens proxy. "ClusterIP" is sufficient in the case when the Athens proxy will be used from within the cluster. To expose Athens outside of the cluster, consider using a "NodePort" or "LoadBalancer" service. This can be changed by setting the `service.type` value when installing the chart. For example, to deploy Athens using a NodePort service, the following command could be used:

```console
helm install gomods/athens-proxy -n athens --namespace athens --set service.type=NodePort
```

### Ingress Resource

The chart can optionally create a Kubernetes [Ingress Resource](https://kubernetes.io/docs/concepts/services-networking/ingress/#the-ingress-resource) for you as well. To enable this feature, set the `ingress.enabled` resource to true. 

```console
helm install gomods/athens-proxy -n athens --namespace athens --set ingress.enabled=true
```

Further configuration values are available in the `values.yaml` file:

```yaml
ingress:
  enabled: true
  annotations:
    certmanager.k8s.io/cluster-issuer: "letsencrypt-prod"
    kubernetes.io/tls-acme: "true"
    ingress.kubernetes.io/force-ssl-redirect: "true"
    kubernetes.io/ingress.class: nginx
  hosts: 
    - athens.mydomain.com
  tls:
    - secretName: athens.mydomain.com
      hosts:
        - "athens.mydomain.com
```

Example above sets automatic creation/retrieval of TLS certificates from [Let's Encrypt](https://letsencrypt.org/) with [cert-manager](https://hub.helm.sh/charts/jetstack/cert-manager) and uses [nginx-ingress controller](https://hub.helm.sh/charts/stable/nginx-ingress) to expose Athens externally to internet.

Add it to `override-values.yaml` file and run:

```console
helm install gomods/athens-proxy -n athens --namespace athens -f override-values.yaml
```

### Upstream module repository

You can set the `URL` for the [upstream module repository](https://docs.gomods.io/configuration/upstream/) then Athens will try to download modules from the upstream when it doesn't find them in its own storage.

You can use `https://gocenter.io` to use JFrog's GoCenter as an upstream here, or you can also use another Athens server as well.

The example below shows you how to set GoCenter up as upstream module repository:

```yaml
upstreamProxy:
  enabled: true
  url: "https://gocenter.io"
```

Add it to `override-values.yaml` file and run:

```console
helm install gomods/athens-proxy -n athens --namespace athens -f override-values.yaml
```

### .netrc file support

A `.netrc` file can be shared as a secret to allow the access to private modules.
The secret must be created from a `netrc` file using the following command (the name of the file **must** be netrc):

```console
kubectl create secret generic netrcsecret --from-file=./netrc
```

In order to instruct athens to fetch and use the secret, `netrc.enabled` flag must be set to true:

```console
helm install gomods/athens-proxy -n athens --namespace athens --set netrc.enabled=true
```
