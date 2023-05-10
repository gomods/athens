---
title: Install Athens on Kubernetes
description: Installing an Athens Instance on Kubernetes
weight: 1
---

When you follow the instructions in the [Walkthrough](/walkthrough), you end up with an Athens Proxy that uses in-memory storage. This is only suitable for trying out the Athens proxy for a short period of time, as you will quickly run out of memory and Athens won't persist modules between restarts. In order to run a more production-like proxy, you may want to run Athens on a [Kubernetes](https://kubernetes.io/) cluster. To aid in deployment of the Athens proxy on Kubernetes, a [Helm](https://www.helm.sh/) chart has been provided. This guide will walk you through installing Athens on a Kubernetes cluster using Helm.

-   [Prerequisites](#prerequisites)
-   [Configure Helm](#configure-helm)
-   [Deploy Athens](#deploy-athens)

---

## Prerequisites

In order to install Athens on your Kubernetes cluster, there are a few prerequisites that you must satisfy. If you already have completed the following steps, please continue to [configuring helm](#configure-helm). This guide assumes you have already created a Kubernetes cluster.

-   Install the [Kubernetes CLI](#install-the-kubernetes-cli).
-   Install the [Helm CLI](#install-the-helm-cli).

### Install the Kubernetes CLI

To work with your Kubernetes Cluster, it's necessary to have the Kubernetes CLI, also known as `kubectl`, installed on your computer. You can download and install it by following the instructions provided in the [official Kubernetes documentation](https://kubernetes.io/docs/tasks/tools/install-kubectl/).

### Install the Helm CLI

[Helm](https://github.com/helm/helm) is a tool for installing pre-configured applications on Kubernetes. You can download and install `helm` by following the instructions available on the [official Helm documentation](https://helm.sh/docs/intro/install/).

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
$ helm repo add gomods https://gomods.github.io/athens-charts
$ helm repo update
```

Next, install the chart with default values to `athens` namespace:

```console
$ helm install gomods/athens-proxy -n athens --namespace athens
```

This will deploy a single Athens instance in the `athens` namespace with `disk` storage enabled. Additionally, a `ClusterIP` service will be created.

## Advanced Configuration

### Replicas

By default, the chart will install Athens with a replica count of 1. To change this, change the `replicaCount` value:

```console
helm install gomods/athens-proxy -n athens --namespace athens --set replicaCount=3
```

### Resources

By default, the chart will install Athens without specific resource requests or limits. To change this, change the `resources` value:

```console
helm install gomods/athens-proxy -n athens --namespace athens \
  --set resources.requests.cpu=100m \
  --set resources.requests.memory=64Mi \
  --set resources.limits.cpu=100m \
  --set resources.limits.memory=64Mi
```

For more information, see [Managing Compute Resources for Containers](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/) in the Kubernetes documentation.

### Give Athens access to private repositories via Github Token (Optional)

1. Create a token at https://github.com/settings/tokens
2. Provide the token to the Athens proxy either through the [config.toml](https://github.com/gomods/athens/blob/main/config.dev.toml) file (the `GithubToken` field) or by setting the `ATHENS_GITHUB_TOKEN` environment variable.

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

```console
helm install gomods/athens-proxy -n athens --namespace athens --set storage.type=mongo --set storage.mongo.url=<some-mongodb-connection-string>
```

#### S3 Configuration

To use S3 storage with Athens, set `storage.type` to `s3` and set `storage.s3.region` and `storage.s3.bucket` to the desired AWS region and
S3 bucket name, respectively. By default, Athens will attempt to load AWS credentials using the AWS SDK from the chain of environment
variables, shared credentials files, and EC2 instance credentials. To manually specify AWS credentials, set `storage.s3.access_key_id`,
`storage.s3.secret_access_key`, and change `storage.s3.useDefaultConfiguration` to `false`.

```console
helm install gomods/athens-proxy -n athens --namespace athens --set storage.type=s3 --set storage.s3.region=<your-aws-region> --set storage.s3.bucket=<your-bucket>
```

#### Minio Configuration

To use S3 storage with Athens, set `storage.type` to `minio`. You need to set `storage.minio.endpoint` as the URL of your minio-installation.
This URL can also be an kubernetes-internal one (e.g. something like `minio-service.default.svc`).
You need to create a bucket inside your minio-installation or use an existing one. The bucket needs to be referenced in `storage.minio.bucket`.
Last athens need authentication credentials for your minio in `storage.minio.accessKey` and `storage.minio.secretKey`.

```console
helm install gomods/athens-proxy -n athens --namespace athens --set storage.type=minio --set storage.minio.endpoint=<your-minio-endpoint> --set storage.minio.bucket=<your-bucket> --set storage.minio.accessKey=<your-minio-access-key> --set storage.minio.secretKey=<your-minio-secret-key>
```

#### Google Cloud Storage

To use Google Cloud Storage storage with Athens, set `storage.type` to `gcp`. You need to set `storage.gcp.projectID` and `storage.gcp.bucket` to the
desired GCP project and bucket name, respectively.

Depending on your deployment environment you will also need to set `storage.gcp.serviceAccount` to a key which has read/write access to
the GCS bucket. If you are running Athens inside GCP, you will most likely not need this as GCP figures out internal authentication between products for you.

```console
helm install gomods/athens-proxy -n athens --namespace athens --set storage.type=gcp --set storage.gcp.projectID=<your-gcp-project> --set storage.gcp.bucket=<your-bucket>
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

Example above sets automatic creation/retrieval of TLS certificates from [Let's Encrypt](https://letsencrypt.org/) with [cert-manager](https://hub.helm.sh/charts/jetstack/cert-manager) and uses [nginx-ingress controller](https://hub.helm.sh/charts/stable/nginx-ingress) to expose Athens externally to the Internet.

Add it to `override-values.yaml` file and run:

```console
helm install gomods/athens-proxy -n athens --namespace athens -f override-values.yaml
```

### Upstream module repository

You can set the `URL` for the [upstream module repository](https://docs.gomods.io/configuration/upstream/) then Athens will try to download modules from the upstream when it doesn't find them in its own storage.

You have a few good options for what you can set as an upstream:

-   `https://gocenter.io` to use JFrog's GoCenter
-   `https://proxy.golang.org` to use the Go Module Mirror
-   The URL to any other Athens server

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

### gitconfig support

A [gitconfig](https://git-scm.com/book/en/v2/Customizing-Git-Git-Configuration) file can be shared as a secret to allow
the access to modules in private git repositories. For example, you can configure access to private repositories via HTTPS
using personal access tokens on GitHub, GitLab and other git services.

First of all, prepare your gitconfig file:

```console
cat << EOF > /tmp/gitconfig
[url "https://user:token@git.example.com/"]
    insteadOf = ssh://git@git.example.com/
    insteadOf = https://git.example.com/
EOF
```

Next, create the secret using the file created above:

```console
kubectl create secret generic athens-proxy-gitconfig --from-file=gitconfig=/tmp/gitconfig
```

In order to instruct athens to use the secret, set appropriate flags (or parameters in `values.yaml`):

```console
helm install gomods/athens-proxy --name athens --namespace athens \
    --set gitconfig.enabled=true \
    --set gitconfig.secretName=athens-proxy-gitconfig \
    --set gitconfig.secretKey=gitconfig
```
