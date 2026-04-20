---
title: "使用 BOSH 安装 Athens"
description: "使用 BOSH 安装 Athens 实例"
weight: 8
---

Athens 可以通过多种方式部署。以下指南说明了如何使用 [BOSH](https://bosh.io)，一个部署和生命周期管理工具，在任何受 BOSH 支持的基础设施即服务（IaaS）上的虚拟机（VM）上部署 Athens 服务器。

---

## 前提条件

* 安装 [BOSH](#install-bosh)
* 设置[基础设施](#setup-the-infrastructure)

### 安装 BOSH

请确保已安装 [BOSH CLI](https://bosh.io/docs/cli-v2-install/)，并在您选择的基础设施上设置了 [BOSH Director](https://bosh.io/docs/quick-start/)。

### 设置基础设施

如果您选择部署在 IaaS 提供商上，在开始部署之前需要设置一些前提条件。根据您要部署到哪个 IaaS，您可能需要创建：

* **公网 IP**：用于关联 Athens 虚拟机的公网 IP 地址。
* **防火墙规则**：必须允许以下入站端口

    - `3000/tcp` - Athens 代理端口（如果您在作业属性中指定了不同于默认端口 3000 的端口，请相应地调整此规则）。

    出站流量应根据您的需求进行限制。

#### Amazon Web Services (AWS)

AWS 需要额外的设置，应使用以下模板添加到 `credentials.yml` 文件中：

```yaml
# 要应用到 VM 的安全组 ID
athens_security_groups: [sg-0123456abcdefgh]

# 部署 Athens 的 VPC 子网
athens_subnet_id: subnet-0123456789abcdefgh

# VM 的特定弹性 IP 地址
external_ip: 3.123.200.100
```

凭证需要添加到 `deploy` 命令中，即

```
-o manifests/operations/aws-ops.yml
```

#### VirtualBox

使用 BOSH 安装 Athens 最快的方法可能是使用 VirtualBox 上运行的 Director VM，这对于开发或测试目的已经足够。如果您按照 bosh-lite [安装指南](https://bosh.cloudfoundry.org/docs/bosh-lite) 操作，则部署 Athens 不需要额外的准备工作。


## 部署

部署清单包含管理和更新 BOSH 部署的所有信息。为了帮助在 BOSH 上部署 Athens，[athens-bosh-release](https://github.com/s4heid/athens-bosh-release) 代码库在 `manifests` 目录中提供了基本部署配置的清单。要快速创建一个独立的 Athens 服务器，请克隆 release 代码库并进入该目录：

```console
git clone --recursive https://github.com/s4heid/athens-bosh-release.git
cd athens-bosh-release
```

一旦[基础设施](#setup-the-infrastructure)准备就绪且 BOSH Director 正在运行，请确保已上传 [stemcell](https://bosh.cloudfoundry.org/docs/stemcell/)。如果尚未完成，请从 [bosh.io 的 stemcells 部分](https://bosh.io/stemcells)选择一个 stemcell，并通过命令行上传。此外，还需要一个 [cloud config](https://bosh.cloudfoundry.org/docs/cloud-config/) 用于 Director 和 Athens 部署的 IaaS 特定配置。`manifests` 目录中还包含一个示例 cloud config，可以通过以下命令上传到 Director

```console
bosh update-config --type=cloud --name=athens \
    --vars-file=credentials.yml manifests/cloud-config.yml
```

执行 `deploy` 命令，可以根据您要部署到的 IaaS 使用 ops/vars 文件进行扩展。

```console
bosh -d athens deploy manifests/athens.yml  # 添加额外参数
```

例如，在 AWS 上使用磁盘存储部署 Athens 代理的 deploy 命令如下

```console
bosh -d athens deploy \
    -o manifests/operations/aws-ops.yml \
    -o manifests/operations/with-persistent-disk.yml \
    -v disk_size=1024 \
    --vars-file=credentials.yml manifests/athens.yml
```

这将部署一个带有 1024MB 持久磁盘的单个 Athens 实例到 `athens` 部署中。该实例的 IP 地址可以通过以下命令获取

```console
bosh -d athens instances
```

这对于使用 `GOPROXY` 变量定位 Athens 很有用。您可以按照此[快速入门指南](/try-out)获取更多信息。
