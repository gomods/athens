---
title: 在 Kubernetes 上安装 Athens
description: 在 Kubernetes 上安装 Athens 实例
weight: 1
---

按照[演练](/walkthrough)中的说明进行操作后，您将得到一个使用内存存储的 Athens 代理。这仅适用于短时间体验 Athens，因为您会很快耗尽内存，而且 Athens 在两次重启之间不会保留存储的模块。若要运行更接近生产环境的代理，您可能需要在 [Kubernetes](https://kubernetes.io/) 集群上运行 Athens。为了帮助在 Kubernetes 上部署 Athens，我们提供了一个 [Helm](https://www.helm.sh/) Chart。本指南将指导您使用 Helm 在 Kubernetes 集群上安装 Athens。

* [前提条件](#前提条件)
* [配置 Helm](#配置-helm)
* [部署 Athens](#部署-athens)

---

## 前提条件

要在 Kubernetes 集群上安装 Athens，您需要满足一些前提条件。如果您已经完成以下步骤，请继续[配置 Helm](#配置-helm)。本指南假设您已经创建了 Kubernetes 集群。

* 安装 [Kubernetes CLI](#安装-kubernetes-cli)
* 安装 [Helm CLI](#安装-helm-cli)

### 安装 Kubernetes CLI

为了与 Kubernetes 集群交互，您需要[安装 kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)。

### 安装 Helm CLI

[Helm](https://github.com/kubernetes/helm) 是用于在 Kubernetes 上安装预配置应用程序的工具。可以通过运行以下命令来安装 `helm`：

#### MacOS

```console
brew install kubernetes-helm
```

#### Windows

1. 下载最新版本的 [Helm release](https://storage.googleapis.com/kubernetes-helm/helm-v2.7.2-windows-amd64.tar.gz)。
1. 解压 tar 文件。
1. 将 **helm.exe** 复制到系统 PATH 中的一个目录下。

#### Linux

```console
curl https://raw.githubusercontent.com/kubernetes/helm/master/scripts/get | bash
```

## 配置 Helm

如果您的集群已配置为使用 Helm，请继续[部署 Athens](#部署-athens)。

如果没有，请继续阅读。

### RBAC 集群

如果您的集群启用了 RBAC，则需要创建 ServiceAccount、ClusterRole 和 ClusterRoleBinding 以供 Helm 使用。以下命令将创建它们并初始化 Helm。

```console
kubectl create -f https://raw.githubusercontent.com/Azure/helm-charts/master/docs/prerequisities/helm-rbac-config.yaml
helm init --service-account tiller
```

### 非 RBAC 集群

如果您的集群没有启用 RBAC，则可以简单地初始化 Helm。

```console
helm init
```

在部署 Athens 之前，您需要等待 Tiller 的 Pod 变成 `Ready` 状态。可以通过查看 `kube-system` 中的 Pod 来检查状态：

```console
$ kubectl get pods -n kube-system -w
NAME                                    READY     STATUS    RESTARTS   AGE
tiller-deploy-5456568744-76c6s          1/1       Running   0          5s
```

## 部署 Athens

使用 Helm 安装 Athens 的最快方法是从我们的公共 Helm Chart 仓库中进行部署。首先，使用以下命令添加仓库

```console
$ helm repo add gomods https://gomods.github.io/athens-charts
$ helm repo update
```
接下来，将含有默认值的 Chart 安装到 `athens` 命名空间：

```
$ helm install gomods/athens-proxy -n athens --namespace athens
```

这将在 `athens` 命名空间中部署一个启用了 `disk` 存储的 Athens 实例。此外，还将创建一个 `ClusterIP` 服务。

## 高级配置

### 副本数

默认情况下，该 Chart 将安装副本数为 1 的 Athens。要更改此设置，请更改 `replicaCount` 值：

```console
helm install gomods/athens-proxy -n athens --namespace athens --set replicaCount=3
```

### 资源

默认情况下，该 Chart 将在没有特定资源请求或限制的情况下安装 Athens。要更改此设置，请更改 `resources` 值：

```console
helm install gomods/athens-proxy -n athens --namespace athens \
  --set resources.requests.cpu=100m \
  --set resources.requests.memory=64Mi \
  --set resources.limits.cpu=100m \
  --set resources.limits.memory=64Mi
```

有关更多信息，请参阅 Kubernetes 文档中的[管理容器的计算资源](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/)。

### 通过 Github Token 授予 Athens 访问私有仓库的权限（可选）

1. 在 https://github.com/settings/tokens 上创建一个 Token
2. 通过 [config.toml](https://github.com/gomods/athens/blob/main/config.dev.toml) 文件（`GithubToken` 字段）或通过设置 `ATHENS_GITHUB_TOKEN` 环境变量，将 Token 提供给 Athens 代理。

### 存储提供商

Helm Chart 目前支持使用两个不同的存储提供商来运行 Athens：`disk` 和 `mongo`。默认使用的是 `disk` 存储提供商。

#### 磁盘存储配置

当使用 `disk` 存储提供商时，可以配置许多有关数据持久化的选项。默认情况下，Athens 将使用 `emptyDir` 卷进行部署。这可能不足以满足生产用例，因此该 Chart 还允许您通过 [PersistentVolumeClaim](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims) 配置持久化。该 Chart 目前允许您设置以下值：

```yaml
persistence:
  enabled: false
  accessMode: ReadWriteOnce
  size: 4Gi
  storageClass:
```

将其添加到 `override-values.yaml` 文件并运行：

```console
helm install gomods/athens-proxy -n athens --namespace athens -f override-values.yaml
```

`enabled` 用于开启 Chart 的 PVC 功能，而其他值直接与 PersistentVolumeClaim 中定义的值相关。


#### MongoDB 配置

要使用 MongoDB 存储提供商，您首先需要一个 MongoDB 实例。部署 MongoDB 后，就可以通过 `storage.mongo.url` 字段使用连接字符串配置 Athens。您还需将 `storage.type` 设置为 "mongo"。

```
helm install gomods/athens-proxy -n athens --namespace athens --set storage.type=mongo --set storage.mongo.url=<some-mongodb-connection-string>
```

#### S3 配置

要在 Athens 中使用 S3 存储，请将 `storage.type` 设置为 `s3`，并将 `storage.s3.region` 和 `storage.s3.bucket` 分别设置为所使用的 AWS 区域和 S3 存储桶名称。默认情况下，Athens 将尝试使用 AWS SDK 从环境变量、共享凭证文件和 EC2 实例凭证组成的链中加载 AWS 凭证。要手动指定 AWS 凭证，请设置 `storage.s3.access_key_id`、`storage.s3.secret_access_key`，并将 `storage.s3.useDefaultConfiguration` 更改为 `false`。

```
helm install gomods/athens-proxy -n athens --namespace athens --set storage.type=s3 --set storage.s3.region=<your-aws-region> --set storage.s3.bucket=<your-bucket>
```

#### Minio 配置

若要在 Athens 中使用 Minio 存储，请将 `storage.type` 设置为 `minio`。您需要设置 `storage.minio.endpoint` 作为 Minio 安装的 URL。这个 URL 也可以是 Kubernetes 内部地址（例如 `minio-service.default.svc`）。您需要在 Minio 安装过程中创建一个存储桶或使用现有的一个。存储桶需要在 `storage.minio.bucket` 中引用。最后，Athens 需要在 `storage.minio.accessKey` 和 `storage.minio.secretKey` 中设置您的 Minio 认证凭证。


```
helm install gomods/athens-proxy -n athens --namespace athens --set storage.type=minio --set storage.minio.endpoint=<your-minio-endpoint> --set storage.minio.bucket=<your-bucket> --set storage.minio.accessKey=<your-minio-access-key> --set storage.minio.secretKey=<your-minio-secret-key>
```

#### Google Cloud Storage

要将 Google Cloud Storage 与 Athens 一起使用，请将 `storage.type` 设置为 `gcp`。您需要将 `storage.gcp.projectID` 和 `storage.gcp.bucket` 设置为所需的 GCP 项目和存储桶名称。

根据您的部署环境，您还需要将 `storage.gcp.serviceAccount` 设置为具有对 GCS 存储桶读写权限的密钥。如果您在 GCP 内运行 Athens，您很可能不需要这个，因为 GCP 会自动处理其内部产品之间的认证。

```
helm install gomods/athens-proxy -n athens --namespace athens --set storage.type=gcp --set storage.gcp.projectID=<your-gcp-project> --set storage.gcp.bucket=<your-bucket>
```

### Kubernetes 服务

默认情况下，Kubernetes 中为 Athens 创建了一个 `ClusterIP` 服务。在 Kubernetes 集群内使用 Athens 的场景下，`ClusterIP` 就足够用了。如果要在 Kubernetes 集群外向 Athens提供服务，请考虑使用 "NodePort" 或 "LoadBalancer"。可以在安装 Chart 时通过设置 `service.type` 值来更改此设置。例如，要使用 NodePort 服务部署 Athens，可以使用以下命令：

```console
helm install gomods/athens-proxy -n athens --namespace athens --set service.type=NodePort
```

### Ingress 资源

该 Chart 可以选择性地为您创建一个 Kubernetes [Ingress 资源](https://kubernetes.io/docs/concepts/services-networking/ingress/#the-ingress-resource)。要启用此功能，请将 `ingress.enabled` 设置为 true。

```console
helm install gomods/athens-proxy -n athens --namespace athens --set ingress.enabled=true
```

`values.yaml` 文件中提供了更多配置选项：

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

上述示例使用 [cert-manager](https://hub.helm.sh/charts/jetstack/cert-manager) 从 [Let's Encrypt](https://letsencrypt.org/) 设置 TLS 证书的自动创建/检索，并使用 [nginx-ingress controller](https://hub.helm.sh/charts/stable/nginx-ingress) 将 Athens 对外暴露于互联网中。

将其添加到 `override-values.yaml` 文件中并运行：

```console
helm install gomods/athens-proxy -n athens --namespace athens -f override-values.yaml
```

### 上游模块仓库

您可以为[上游模块仓库](https://docs.gomods.io/configuration/upstream/)设置 `URL`，然后当 Athens 在自己的存储中找不到某个模块时，它将尝试从上游模块仓库中下载该模块。

对于可作为上游的模块仓库，以下是几个好的选择：

-  `https://gocenter.io` 使用 JFrog 的 GoCenter
-  `https://proxy.golang.org` 使用 Go Module 镜像
-  指向任何其他 Athens 服务器的 URL

以下示例显示了如何将 GoCenter 设置为上游模块仓库：

```yaml
upstreamProxy:
  enabled: true
  url: "https://gocenter.io"
```

将其添加到 `override-values.yaml` 文件中并运行：

```console
helm install gomods/athens-proxy -n athens --namespace athens -f override-values.yaml
```

### .netrc 文件支持

.netrc 文件可以作为密钥共享，以允许访问私有模块。使用以下命令从 `netrc` 文件创建密钥（文件的名称**必须**为 netrc）：

```console
kubectl create secret generic netrcsecret --from-file=./netrc
```

为了指示 Athens 获取并使用密钥，`netrc.enabled` 标志必须设置为 true：

```console
helm install gomods/athens-proxy -n athens --namespace athens --set netrc.enabled=true
```

### gitconfig 支持

[gitconfig](https://git-scm.com/book/en/v2/Customizing-Git-Git-Configuration) 可以作为密钥共享，以允许访问私有 git 仓库中的模块。例如，您可以使用 GitHub、GitLab 和其他 git 服务上的个人访问令牌（Token），通过 HTTPS 配置对私有仓库的访问。

首先，准备您的 gitconfig 文件：

```console
cat << EOF > /tmp/gitconfig
[url "https://user:token@git.example.com/"]
    insteadOf = ssh://git@git.example.com/
    insteadOf = https://git.example.com/
EOF
```

接下来，使用上面创建的文件创建密钥：

```console
kubectl create secret generic athens-proxy-gitconfig --from-file=gitconfig=/tmp/gitconfig
```

为了指示 Athens 使用密钥，请设置适当的标志（或 `values.yaml` 中的参数）：

```console
helm install athens gomods/athens-proxy --namespace athens \
    --set gitconfig.enabled=true \
    --set gitconfig.secretName=athens-proxy-gitconfig \
    --set gitconfig.secretKey=gitconfig
```

## 进一步链接

- [Kubernetes 上的 Athens 即服务（适用于 GitLab）](https://medium.com/gitconnected/athens-go-proxy-as-a-service-on-kubernetes-8fb1f5fa320d)
