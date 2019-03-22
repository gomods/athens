# Athens Proxy Helm Chart

## What is Athens?

[Athens](https://docs.gomods.io) is a Server for Your Go Packages.

Athens provides a server for [Go Modules](https://github.com/golang/go/wiki/Modules) that you can run. It serves public code and your private code for you, so you don't have to pull directly from a version control system (VCS) like GitHub or GitLab.

## Prerequisites Details

* Kubernetes 1.10+

## Requirements

- A running Kubernetes cluster
- [Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) installed and setup to use the cluster
- [Helm](https://helm.sh/) installed and setup to use the cluster (helm init) or [Tilerless Helm](https://github.com/rimusz/helm-tiller)

## Deploy Athens

The fastest way to install Athens using Helm is to deploy it from our public Helm chart repository. First, add the repository with this command:

```console
$ helm repo add gomods https://athens.blob.core.windows.net/charts
$ helm repo update
```

Next, install the chart using no arguments.  

```
$ helm install gomods/athens-proxy -n athens
```

This will deploy a single Athens instance in the `default` namespace with `disk` storage enabled. Additionally, a `ClusterIP` service will be created.

## Advanced Configuration

### Give Athens access to private repositories via Github Token (Optional)
1. Create a token at https://github.com/settings/tokens
2. Provide the token to the Athens proxy either through the [config.toml](https://github.com/gomods/athens/blob/master/config.dev.toml#L111) file or by setting the `ATHENS_GITHUB_TOKEN` environment variable.

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

`enabled` is used to turn on the PVC feature of the chart, while the other values relate directly to the values defined in the PersistentVolumeClaim documentation.

#### Mongo DB Configuration

To use the Mongo DB storage provider, you will first need a MongoDB instance. Once you have deployed MongoDB, you can configure Athens using the connection string via `storage.mongo.url`. You will also need to set `storage.type` to "mongo".

```
helm install ./charts/proxy -n athens --set storage.type=mongo --set storage.mongo.url=<some-mongodb-connection-string>
```

### Kubernetes Service

By default, a Kubernetes `ClusterIP` service is created for the Athens proxy. "ClusterIP" is sufficient in the case when the Athens proxy will be used from within the cluster. To expose Athens outside of the cluster, consider using a "NodePort" or "LoadBalancer" service. This can be changed by setting the `service.type` value when installing the chart. For example, to deploy Athens using a NodePort service, the following command could be used:

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

### .netrc file support

A `.netrc` file can be shared as a secret to allow the access to private modules.
The secret must be created from a `netrc` file using the following command (the name of the file **must** be netrc):

```console
kubectl create secret generic netrcsecret --from-file=./netrc
```

In order to instruct athens to fetch and use the secret, `netrc.enabled` flag must be set to true:

```console
helm install ./charts/proxy -n athens --set netrc.enabled=true
```

### GoCenter as upstream proxy server

[GoCenter](https://gocenter.jfrog.com) is a public central repository of Go modules which provides repeatable and fast module resolution. GoCenter automatically processes and validates Go projects directly from source code repositories into immutable, versioned modules, making them freely available to the Go developer community. In this way, it empowers the Go community to reliably generate reproducible builds of their Go applications.

```yaml
goCenter:
  # Set https://gocenter.io as upstream proxy
  enabled: enable
  url: "https://gocenter.io"
```

### Private git servers over ssh support

One or many git servers are added to `sshGitServers`, and the corresponding config files (git config and ssh config) and ssh keys will be created. Athens then will use these configs and keys to download the source from the git servers.

```yaml
sshGitServers: 
  ## Private git servers over ssh
  ## to enable uncomment lines with single hash below
  ## hostname of the git server
  - host: git.example.com
    ## ssh username
    user: git
    ## ssh private key for the user
    privateKey: |
      -----BEGIN RSA PRIVATE KEY-----
      ...
      -----END RSA PRIVATE KEY-----
    ## ssh port
    port: 22
```
