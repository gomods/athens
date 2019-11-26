---
title: 在Kubernetes上安装Athens
description: 在Kubernetes上安装Athens
weight: 1
---

当您按照[Walkthrough](/walkthrough)中的说明进行操作时, 最终会得到使用内存作为存储的Athens。 这仅适用于短时间试用Athens，因为您将很快耗尽内存，并且Athens在两次重启之间不会保留储存的模块（modules）。 为了使Athens运行在一个更接近生产级别的环境上, 您可能需要在 [Kubernetes](https://kubernetes.io/) 集群上运行Athens. 为了帮助在Kubernetes上部署Athens，, 我们提供了一个 [Helm](https://www.helm.sh/) chart . 本指南将指导您使用Helm将Athens安装在Kubernetes集群上。

* [前提条件](#前提条件)
* [配置Helm](#配置Helm)
* [部署Athens](#部署Athens)

---

## 前提条件

为了在Kubernetes集群上安装Athens，您必须满足一些前提条件.如果您已经完成以下步骤，请继续执行[配置Helm](#配置Helm). 本指南假设您已经创建了Kubernetes集群.

* 安装 [Kubernetes CLI](#安装Kubernetes-CLI).
* 安装 [Helm CLI](#安装Helm-CLI).

### 安装Kubernetes CLI

为了与Kubernetes集群进行交互，您需要 [安装 kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)。

### 安装Helm CLI

[Helm](https://github.com/kubernetes/helm) 是用于在Kubernetes上安装预先配置好的应用程序的工具。
可以通过运行以下命令来安装`helm`：

#### MacOS

```console
brew install kubernetes-helm
```

#### Windows

1. 下载最新版本的 [Helm release](https://storage.googleapis.com/kubernetes-helm/helm-v2.7.2-windows-amd64.tar.gz)。
1. 解压tar文件。
1. 拷贝 **helm.exe** 到系统 PATH 中的一个目录下.

#### Linux

```console
curl https://raw.githubusercontent.com/kubernetes/helm/master/scripts/get | bash
```

## 配置Helm

如果您的集群已配置为使用Helm，请继续[部署Athens](#部署Athens).

如果没有，请继续阅读。

### RBAC 集群

如果您的集群启用了RBAC，则需要创建ServiceAccount，ClusterRole和ClusterRoleBinding以供Helm使用。 以下命令将创建它们并初始化Helm。

```console
kubectl create -f https://raw.githubusercontent.com/Azure/helm-charts/master/docs/prerequisities/helm-rbac-config.yaml
helm init --service-account tiller
```

### 非RBAC 集群

如果您的集群没有启用rbac，则可以轻松初始化helm。

```console
helm init
```

在部署Athens之前, 你需要等待Tiller的pod变成 `Ready`状态. 您可以通过查看 `kube-system`中的Pod来检查状态:

```console
$ kubectl get pods -n kube-system -w
NAME                                    READY     STATUS    RESTARTS   AGE
tiller-deploy-5456568744-76c6s          1/1       Running   0          5s
```

## 部署Athens

使用Helm安装Athens的最快方法是从我们的公共Helm Chart库中进行部署。 首先，使用以下命令添加库

```console
$ helm repo add gomods https://athens.blob.core.windows.net/charts
$ helm repo update
```
接下来，将含有默认值的chart安装到`athens`命名空间：

```
$ helm install gomods/athens-proxy -n athens --namespace athens
```

这将在`athens`命名空间中部署一个启用了`disk`本地存储的Athens实例。此外，还将创建一个`ClusterIP`服务。

## 高级配置

### 多副本

默认情况下，该chart将安装副本数量为1的athens。要更改此设置，请更改`replicaCount`值：

```console
helm install gomods/athens-proxy -n athens --namespace athens --set replicaCount=3
```

### 资源

默认情况下，该chart将在没有特定资源请求或限制的情况下安装athens。 要更改此设置，请更改`resources`值：

```console
helm install gomods/athens-proxy -n athens --namespace athens \
  --set resources.requests.cpu=100m \
  --set resources.requests.memory=64Mi \
  --set resources.limits.cpu=100m \
  --set resources.limits.memory=64Mi
```

有关更多信息，请参阅Kubernetes文档中的[管理容器的计算资源](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/) 。

### 通过Github令牌(Token)授予athens访问私有存储库的权限（可选）

1. 在 https://github.com/settings/tokens 上创建一个令牌(Token)
2. 通过 [config.toml](https://github.com/gomods/athens/blob/master/config.dev.toml) 文件 ( `GithubToken` 字段) 或 通过设置`ATHENS_GITHUB_TOKEN` 环境变量，将令牌提供给Athens代理.

### 存储提供程序（storage provider）

Helm chart目前支持使用两个不同的存储提供程序来运行Athens：`disk`和`mongo`。 默认使用的是`disk`存储提供程序。

#### 磁盘存储配置

当使用`disk`存储提供程序时，可以配置许多有关数据持久性的选项。默认情况下，雅典将使用`emptyDir`卷进行部署。这可能不足以满足生产用例，因此该chart还允许您通过[PersistentVolumeClaim](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims)配置持久性。该chart目前允许您设置以下值：

```yaml
persistence:
  enabled: false
  accessMode: ReadWriteOnce
  size: 4Gi
  storageClass:
```

将其添加到`override values.yaml`文件并运行：

```console
helm install gomods/athens-proxy -n athens --namespace athens -f override-values.yaml
```

`enabled`用于打开chart的PVC功能，而其他值直接与PersistentVolumeClaim中定义的值相关。


#### Mongo DB 配置

要使用Mongo DB存储提供程序，您首先需要一个MongoDB实例。当部署了MongoDB后，就可以通过`storage.mongo.url`字段使用连接字符串配置Athens。 您还需将`storage.type`设置为“mongo”。

```
helm install gomods/athens-proxy -n athens --namespace athens --set storage.type=mongo --set storage.mongo.url=<some-mongodb-connection-string>
```

#### S3 配置

要在Athens中使用S3存储，请将`storage.type`设置为`s3`，并将`storage.s3.region`和`storage.s3.bucket`分别设置为所使用的AWS区域和S3存储桶名称。 默认情况下，Athens将尝试使用AWS SDK从环境变量、共享凭证文件（shared credentials files）和EC2实例凭证组成的链中加载AWS凭证。 要手动指定AWS凭证，请设置`storage.s3.access_key_id`，`storage.s3.secret_access_key`，并将`storage.s3.useDefaultConfiguration`更改为`false`。

```
helm install gomods/athens-proxy -n athens --namespace athens --set storage.type=s3 --set storage.s3.region=<your-aws-region> --set storage.s3.bucket=<your-bucket>
```

#### Minio 配置

若要在Athens中使用S3存储，请将`storage.type`设置为`minio`。您需要设置`storage.minio.endpoint` 作为minio安装的URL。这个URL也可以是kubernetes内部地址的（例如`minio-service.default.svc`）。您需要在minio安装过程中创建一个桶（bucket）或使用现有的一个桶。桶需要在`storage.minio.bucket`中引用。最后，Athens需要在`storage.minio.accesskey`和`storage.minio.secretkey`中设置您的minio的身份验证凭据。


```
helm install gomods/athens-proxy -n athens --namespace athens --set storage.type=minio --set storage.minio.endpoint=<your-minio-endpoint> --set storage.minio.bucket=<your-bucket> --set storage.minio.accessKey=<your-minio-access-key> --set storage.minio.secretKey=<your-minio-secret-key>
```

### Kubernetes 服务

默认情况下，Kubernetes中为Athens创建了一个`ClusterIP` 服务。在Kubernetes集群内使用Athens的场景下,`clusterip`就足够用了。如果要在Kubernetes集群外提供Athens的服务，请考虑使用“nodeport”或“loadbalancer”。可以在安装chart时通过设置`service.type`值来更改此设置。例如，要使用nodeport服务部署Athens，可以使用以下命令：

```console
helm install gomods/athens-proxy -n athens --namespace athens --set service.type=NodePort
```

### Ingress 资源

该chart可以选择性为您的创建一个Kubernetes [Ingress 资源](https://kubernetes.io/docs/concepts/services-networking/ingress/#the-ingress-resource)。要启用此功能，请将`ingress.enabled`资源设置为true。

```console
helm install gomods/athens-proxy -n athens --namespace athens --set ingress.enabled=true
```

`values.yaml`文件中提供了更多配置选项：

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

上述的示例使用[cert-manager](https://hub.helm.sh/charts/jetstack/cert-manager)从[Let's Encrypt](https://letsencrypt.org/)设置TLS证书的自动创建/检索。 并使用[nginx-ingress controller](https://hub.helm.sh/charts/stable/nginx-ingress) 将Athens对外暴露于互联网中。

将其添加到`override-values.yaml`文件中并运行：

```console
helm install gomods/athens-proxy -n athens --namespace athens -f override-values.yaml
```

### 上游的模块（module）库

您可以为[上游模块库](https://docs.gomods.io/configuration/upstream/)设置`URL`，然后当Athens在自己的存储中找不到某个模块（module）时，它将尝试从上游模块库中下载该模块。

对于可用的为上游模块库，一下是几个好的选择：

-  `https://gocenter.io`  使用JFrog的GoCenter
-  `https://proxy.golang.org`  使用Go Module 镜像
-  指向任何其他Athens服务器的URL

以下示例显示了如何将GoCenter设置为上游模块库：

```yaml
upstreamProxy:
  enabled: true
  url: "https://gocenter.io"
```

将其添加到 `override-values.yaml` 文件里并运行:

```console
helm install gomods/athens-proxy -n athens --namespace athens -f override-values.yaml
```

### .netrc文件支持

.netrc文件可以作为密钥共享，以允许访问私有模块。
使用以下命令从`netrc`文件创建密钥（文件的名称**必须**为netrc）：

```console
kubectl create secret generic netrcsecret --from-file=./netrc
```
为了指导athens获取并使用密钥，`netrc.enabled`标志必须设置为true：

```console
helm install gomods/athens-proxy -n athens --namespace athens --set netrc.enabled=true
```

### gitconfig支持

[gitconfig](https://git-scm.com/book/en/v2/Customizing-Git-Git-Configuration)可以作为私钥共享，以允许访问私有git库中的模块。 例如，您可以使用GitHub，Gitlab和其他git服务上的个人访问令牌（token），通过HTTPS配置对私有存储库的访问。

首先，准备你的gitconfig文件：

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

为了使athens使用密钥，请设置适当的标志（或`values.yaml`中的参数）：

```console
helm install gomods/athens-proxy --name athens --namespace athens \
    --set gitconfig.enabled=true \
    --set gitconfig.secretName=athens-proxy-gitconfig \
    --set gitconfig.secretKey=gitconfig
```
