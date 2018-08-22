---
title: Install Athens on Kubernetes
description: Installing an Athens Instance on Kubernetes
---

When you follow the instructions in the [Walkthrough](/walkthrough), you end up with an Athens Proxy that uses in-memory storage. This is only suitable for trying out the proxy for a short period of time, as you will quickly run out of memory and Athens won't persist modules between restarts. In order to run a more production-like proxy, you may with to run Athens on a [Kubernetes](https://kubernetes.io/) cluster. To aid in deployment of the Athens proxy on Kubernetes, a [Helm](https://www.helm.sh/) chart has been provided. This guide will walk you through installing Athens on a Kubernetes cluster using Helm.

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

The fastest way to install Athens using Helm is to simply clone this repository and install the chart using no arguments.  

```
git clone git@github.com:gomods/athens.git
cd athens
helm install ./charts/proxy -n athens
```

This will deploy a single Athens instance in the `default` namespace with `disk` storage enabled. Additionally, a `ClusterIP` service will be created.

## Advanced Configuration

### Storage Providers

The Helm chart currently supports running Athens with two different storage providers: `disk` and `mongo`. The default behavior is to use the `disk` storage provider.

#### Disk Storage Configuration

When using the `disk` storage provider, you can configure a number of options regarding data persistence. By default, Athens will deploy using an `emptyDir` volume. This probably isn't sufficient for production use cases, so the chart also allows you to configure persistence via a [PersistentVolumeClaim](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims). The chartr currently allows you to set the following values:

```yaml
persistence:
  enabled: false
  accessMode: ReadWriteOnce
  size: 4Gi
  storageClass:
```

`enabled` is used to turn on the PVC feature of the chart, while the other values relate directly to the values defined in the PersistentVolumeClaim documentation.

#### Mongo DB Configuration

To use the Mongo DB storage provider, you will first need a MongoDB instance. Once you have deployed MongoDB, you can configure Athens using the connection string via `storage.mongo.url`. You will also need to set `storage.type` to "mongo".

```
helm install ./charts/proxy -n athens --set storage.type=mongo --set storage.mongo.url=<some-mongodb-connection-string>
```

### Kuberentes Service

By default, a Kubernetes `ClusterIP` service is created for the Athens proxy. "ClusterIP" is sufficient in the case when the Proxy will be used from within the cluster. To expose Athens outside of the cluster, consider using a "NodePort" or "LoadBalancer" service. This can be changed by setting the `service.type` value when installing the chart. For example, to deploy Athens using a NodePort service, the following command could be used:

```console
helm install ./charts/proxy -n athens --set service.type=NodePort
```

### Ingress Resource

The chart can optionally create a Kubernetes [Ingress Resource](https://kubernetes.io/docs/concepts/services-networking/ingress/#the-ingress-resource) for you as well. To enable this feature, set the `ingress.enabled` resource to true. 

```console
helm install ./charts/proxy -n athens --set ingress.enabled=true
```

Further configuration values are available in the `values.yaml` file:

```yaml
ingress:
  enabled: false
  # provie key/value annotations
  annotations:
  # Provide an array of values for the ingress host mapping
  hosts:
  # Provide a base64 encoded cert for TLS use 
  tls: 
```

### Replicas

By default, the chart will install Athens with a replica count of 1. To change this, change the `replicaCount` value:

```console
helm install ./charts/proxy -n athens --set replicaCount=3
```
