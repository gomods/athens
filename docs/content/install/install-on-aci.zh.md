---
title: "在Azure Container Instances上安装Athens"
date: 2018-12-06T13:17:37-08:00
draft: false
weight: 3
---

当您按照[Walkthrough](/walkthrough)中的说明进行操作时，Athens最终使用的是本地存储空间。 这仅适用于短期试用Athens，因为您将很快耗尽内存，并且Athens在两次重启之间不会保留模块。 本指南将帮助您以一种更适合的方式运行Athens，以用于提供一个实例供开发团队共享的场景。

在本文中，我们将展示如何在 [Azure Container Instances](https://cda.ms/KR) (ACI) 上运行Athens.

## 选择存储提供商（Provider）

Athens目前支持许多存储驱动。 为了在ACI上能快捷的使用，我们建议使用本地磁盘作为存储。如果希望更持久的存储数据，我们建议使用MongoDB或其他持久化存储架构。 对于其他提供商，请参阅 [storage provider documentation](/configuration/storage/).

### 必需的环境变量

在执行以下任何命令之前，请确保在系统上设置了下列环境变量：

- `AZURE_ATHENS_RESOURCE_GROUP` - 指定用于安装容器的 [Azure Resource Group](https://www.petri.com/what-are-microsoft-azure-resource-groups) 。在安装Athens之前，你需要设置该环境变量。
    -有关如何创建资源组（resource group）的详细信息，详见[此处](https://cda.ms/KS) 。
- `AZURE_ATHENS_CONTAINER_NAME` - 容器的名称。 应为字母和数字，可以包含“-”和“uu”字符
- `LOCATION` - 指定用于安装容器的  [Azure region](https://cda.ms/KT) 。有关详细列表，请参见上一链接, 同时这里有一个有用的备忘表，你可以立即使用，而不必阅读任何文档:
    - 北美: `eastus2`
    - 欧洲: `westeurope`
    - 亚洲: `southeastasia` 
- `AZURE_ATHENS_DNS_NAME` - 要分配给容器的DNS名称。它必须在您设置的区域（region）内是全局唯一的 (`LOCATION`)


### 使用本地磁盘驱动进行安装

```console
az container create \
-g "${AZURE_ATHENS_RESOURCE_GROUP}" \
-n "${AZURE_ATHENS_CONTAINER_NAME}-${LOCATION}" \
--image gomods/athens:v0.3.0 \
-e "ATHENS_STORAGE_TYPE=disk" "ATHENS_DISK_STORAGE_ROOT=/var/lib/athens" \
--ip-address=Public \
--dns-name="${AZURE_ATHENS_DNS_NAME}" \
--ports="3000" \
--location=${LOCATION}
```

创建ACI容器后，您将看到一个JSON Blob对象，其中包含该容器的公有IP地址. 您还将看到正在运行的容器的 [fully qualified domain name](https://en.wikipedia.org/wiki/Fully_qualified_domain_name) (FQDN) (以 `AZURE_ATHENS_DNS_NAME`为前缀)。

### 使用MongoDB驱动进行安装

首先，请确保您设置了以下环境变量：

- `AZURE_ATHENS_MONGO_URL` - MongoDB 连接字符串。例如： `mongodb://username:password@mongo.server.com/?ssl=true`

然后运行下列创建的命令:

```console
az container create \
-g "${AZURE_ATHENS_RESOURCE_GROUP}" \
-n "${AZURE_ATHENS_CONTAINER_NAME}-${LOCATION}" \
--image gomods/athens:v0.3.0 \
-e "ATHENS_STORAGE_TYPE=mongo" "ATHENS_MONGO_STORAGE_URL=${AZURE_ATHENS_MONGO_URL}" \
--ip-address=Public \
--dns-name="${AZURE_ATHENS_DNS_NAME}" \
--ports="3000" \
--location=${LOCATION}
```

创建ACI容器后，您将看到一个JSON Blob对象，其中包含该容器的公有IP地址. 您还将看到正在运行的容器的 [fully qualified domain name](https://en.wikipedia.org/wiki/Fully_qualified_domain_name) (FQDN) (以 `AZURE_ATHENS_DNS_NAME`为前缀)。

