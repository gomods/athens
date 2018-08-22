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

