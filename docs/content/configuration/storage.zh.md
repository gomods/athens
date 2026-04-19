---
title: 配置存储
description: 在 Athens 中配置存储
weight: 3
---

## 存储

Athens 代理支持多种存储类型：

- [存储](#存储)
- [内存](#内存)
      - [配置：](#配置)
- [磁盘](#磁盘)
      - [配置：](#配置-1)
- [MongoDB](#mongodb)
      - [配置：](#配置-2)
- [Google Cloud Storage](#google-cloud-storage)
      - [配置：](#配置-3)
- [AWS S3](#aws-s3)
      - [配置：](#配置-4)
- [Minio](#minio)
      - [配置：](#配置-5)
    - [DigitalOcean Spaces](#digitalocean-spaces)
      - [配置：](#配置-6)
    - [阿里云 OSS](#阿里云-oss)
      - [配置：](#配置-7)
- [Azure Blob Storage](#azure-blob-storage)
      - [配置：](#配置-8)
- [外部存储](#外部存储)
      - [配置：](#配置-9)
- [运行多个指向同一存储的 Athens](#运行多个指向同一存储的-athens)
  - [使用 etcd 作为单飞机制](#使用-etcd-作为单飞机制)
  - [使用 redis 作为单飞机制](#使用-redis-作为单飞机制)
    - [直接连接 redis](#直接连接-redis)
    - [通过 redis sentinel 连接](#通过-redis-sentinel-连接)

所有这些都使用 `config.toml` 文件进行配置。您需要在 `StorageType` 值中设置有效的驱动，或者可以在服务器上设置环境变量 `ATHENS_STORAGE_TYPE`。
对于大多数驱动，您需要提供额外的配置数据，下文将详细描述。

## 内存

此存储类型不需要任何特定配置，也是 Athens 项目默认使用的。它将所有数据写入本地磁盘的 `tmp` 目录。

**此存储类型仅用于开发目的！**

##### 配置：

    # StorageType 设置代理将使用的存储后端类型。
    # 环境变量覆盖：ATHENS_STORAGE_TYPE
    StorageType = "memory"

## 磁盘

磁盘存储允许将模块存储在文件系统中。模块在磁盘上存储的位置可以配置。

>您可以预填充基于磁盘的存储，以实现无法访问互联网的 Athens 部署。请参阅[此处](/configuration/prefill-disk-cache)了解如何操作。

##### 配置：

    # StorageType 设置代理将使用的存储后端类型。
    # 环境变量覆盖：ATHENS_STORAGE_TYPE
    StorageType = "disk"

    [Storage]
        [Storage.Disk]
            RootPath = "/path/on/disk"

其中 `/path/on/disk` 是您期望的位置。也可以使用 `ATHENS_DISK_STORAGE_ROOT` 环境变量设置。

## MongoDB

此驱动使用 [Mongo](https://www.mongodb.com/) 服务器作为数据存储。启动时，此驱动将在您的 Mongo 服务器上创建一个 `athens` 数据库和 `module` 集合。

##### 配置：

    # StorageType 设置代理将使用的存储后端类型。
    # 环境变量覆盖：ATHENS_STORAGE_TYPE
    StorageType = "mongo"

    [Storage]
        [Storage.Mongo]
            # Mongo 存储的完整 URL
            # 环境变量覆盖：ATHENS_MONGO_STORAGE_URL
            URL = "mongodb://127.0.0.1:27017"

            # 非必填参数
            # Mongo 连接使用的证书路径
            # 环境变量覆盖：ATHENS_MONGO_CERT_PATH
            CertPath = "/path/to/cert/file"

            # 非必填参数
            # 允许不安全的 SSL / http 连接到 mongo 存储
            # 仅用于测试或开发
            # 环境变量覆盖：ATHENS_MONGO_INSECURE
            Insecure = false

            # 非必填参数
            # 允许使用自定义数据库
            # 环境变量覆盖：ATHENS_MONGO_DEFAULT_DATABASE
            DefaultDBName = athens

            # 非必填参数
            # 允许使用自定义集合
            # 环境变量覆盖：ATHENS_MONGO_DEFAULT_COLLECTION
            DefaultCollectionName = modules

## Google Cloud Storage

此驱动使用 [Google Cloud Storage](https://cloud.google.com/storage/)，并假定您已在其中拥有 `account` 和 `bucket`。
如果您从未使用过 Google Cloud Storage，这里有一个[快速指南](https://cloud.google.com/storage/docs/creating-buckets#storage-create-bucket-console)介绍如何在其中创建 `bucket`。

##### 配置：

    # StorageType 设置代理将使用的存储后端类型。
    # 环境变量覆盖：ATHENS_STORAGE_TYPE
    StorageType = "gcp"

    [Storage]
        [Storage.GCP]
            # GCP 存储使用的项目 ID
            # 环境变量覆盖：GOOGLE_CLOUD_PROJECT
            ProjectID = "YOUR_GCP_PROJECT_ID"

            # GCP 存储使用的存储桶
            # 环境变量覆盖：ATHENS_STORAGE_GCP_BUCKET
            Bucket = "YOUR_GCP_BUCKET"

## AWS S3

此驱动使用 [AWS S3](https://aws.amazon.com/s3/)，并假定您已在其中创建了 `account` 和 `bucket`。
如果您从未使用过 Amazon Web Services，这里有一个[快速指南](https://docs.aws.amazon.com/AmazonS3/latest/gsg/GetStartedWithS3.html)介绍如何创建 `bucket`。之后您可以在 `config.toml` 文件中传递您的凭证。如果访问密钥 ID 和秘密访问密钥未在 `config.toml` 中指定，驱动将尝试从 [AWS CLI 配置文件](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html) 中加载默认配置文件的凭证，该文件是在安装期间创建的。

##### 配置：

    # StorageType 设置代理将使用的存储后端类型。
    # 环境变量覆盖：ATHENS_STORAGE_TYPE
    StorageType = "s3"

    [Storage]
            [Storage.S3]
            ### S3 的认证模型如下所示
            ### 如果指定了 AWS_CREDENTIALS_ENDPOINT 并返回有效结果，则使用它
            ### 如果指定了配置变量且它们有效，则返回有效结果，然后使用它
            ### 否则，将默认为默认配置，如下所示
            # 尝试在环境、共享配置（~/.aws/credentials）和 ec2 实例角色
            # 凭证中查找凭证。参见
            # https://godoc.org/github.com/aws/aws-sdk-go#hdr-Configuring_Credentials
            # 和
            # https://godoc.org/github.com/aws/aws-sdk-go/aws/session#hdr-Environment_Variables
            # 了解将影响 aws 配置的环境变量。
            # 设置 UseDefaultConfiguration 将仅使用默认配置。它将在未来的版本中弃用，
            # 不建议使用。

            # S3 存储的区域
            # 环境变量覆盖：AWS_REGION
            Region = "MY_AWS_REGION"

            # S3 存储的访问密钥
            # 环境变量覆盖：AWS_ACCESS_KEY_ID
            Key = "MY_AWS_ACCESS_KEY_ID"

            # S3 存储的秘密密钥
            # 环境变量覆盖：AWS_SECRET_ACCESS_KEY
            Secret = "MY_AWS_SECRET_ACCESS_KEY"

            # S3 存储的会话令牌
            # 非必填参数
            # 环境变量覆盖：AWS_SESSION_TOKEN
            Token = ""

            # 用于存储的 S3 存储桶
            # 环境变量覆盖：ATHENS_S3_BUCKET_NAME
            Bucket = "MY_S3_BUCKET_NAME"

            # 如果为 true，则将使用 s3 端点的路径样式 url
            # 环境变量覆盖：AWS_FORCE_PATH_STYLE
            ForcePathStyle = false

            # 如果为 true，则将使用默认的 aws 配置。这将
            # 尝试在环境、共享配置（~/.aws/credentials）和 ec2 实例角色
            # 凭证中查找凭证。参见
            # https://godoc.org/github.com/aws/aws-sdk-go#hdr-Configuring_Credentials
            # 和
            # https://godoc.org/github.com/aws/aws-sdk-go/aws/session#hdr-Environment_Variables
            # 了解将影响 aws 配置的环境变量。
            # 环境变量覆盖：AWS_USE_DEFAULT_CONFIGURATION
            UseDefaultConfiguration = false

            # https://docs.aws.amazon.com/sdk-for-go/api/aws/credentials/endpointcreds/
            # 请注意，当设置 AwsContainerCredentialsRelativeURI 时，URI 不应以 / 结尾
            # 环境变量覆盖：AWS_CREDENTIALS_ENDPOINT
            CredentialsEndpoint = ""

            # 环境变量覆盖：AWS_CONTAINER_CREDENTIALS_RELATIVE_URI
            # 如果您计划使用 AWS Fargate，请对 CredentialsEndpoint 使用 http://169.254.170.2
            # 参考：https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-metadata-endpoint-v2.html
            AwsContainerCredentialsRelativeURI = ""

            # 可选的端点 URL（仅主机名或完全限定的 URI）
            # 覆盖为 S3 存储客户端生成的默认端点。
            #
            # 当指定端点时，您仍需提供 `Region` 值。
            # 环境变量覆盖：AWS_ENDPOINT
            Endpoint = ""

## Minio

[Minio](https://www.minio.io/) 是一个开源对象存储服务器，提供 S3 兼容块存储的接口。如果您从未使用过 minio，可以阅读此[快速入门指南](https://docs.minio.io/)。Athens 通过 minio 接口支持任何 S3 兼容的对象存储。下面，您可以找到我们为 Minio 提供的不同配置选项。下面提供了 Digital Ocean 和阿里云 OSS 块存储的配置示例。

##### 配置：

    # StorageType 设置代理将使用的存储后端类型。
    # 环境变量覆盖：ATHENS_STORAGE_TYPE
    StorageType = "minio"

    [Storage]
        [Storage.Minio]
            # Minio 存储的端点
            # 环境变量覆盖：ATHENS_MINIO_ENDPOINT
            Endpoint = "127.0.0.1:9001"

            # Minio 存储的访问密钥
            # 环境变量覆盖：ATHENS_MINIO_ACCESS_KEY_ID
            Key = "YOUR_MINIO_SECRET_KEY"

            # Minio 存储的秘密密钥
            # 环境变量覆盖：ATHENS_MINIO_SECRET_ACCESS_KEY
            Secret = "YOUR_MINIO_SECRET_KEY"

            # 为 Minio 连接启用 SSL
            # 默认为 true
            # 环境变量覆盖：ATHENS_MINIO_USE_SSL
            EnableSSL = false

            # 用于存储的 Minio 存储桶
            # 环境变量覆盖：ATHENS_MINIO_BUCKET_NAME
            Bucket = "gomods"

#### DigitalOcean Spaces

为了让 Athens 与 [DigitalOcean Spaces](https://www.digitalocean.com/products/spaces/) 通信，我们使用 Minio 驱动，因为 DO Spaces 尝试[完全兼容它](https://developers.digitalocean.com/documentation/spaces/)。
此外，此存储的配置与我们的代理中的 [Minio](#minio) 配置几乎相同。

##### 配置：

    # StorageType 设置代理将使用的存储后端类型。
    # 环境变量覆盖：ATHENS_STORAGE_TYPE
    StorageType = "minio"

    [Storage]
        [Storage.Minio]
            # DO Spaces 存储的地址
            # 环境变量覆盖：ATHENS_MINIO_ENDPOINT
            Endpoint = "YOUR_ADDRESS.digitaloceanspaces.com"

            # DO Spaces 存储的访问密钥
            # 环境变量覆盖：ATHENS_MINIO_ACCESS_KEY_ID
            Key = "YOUR_DO_SPACE_KEY_ID"

            # DO Spaces 存储的秘密密钥
            # 环境变量覆盖：ATHENS_MINIO_SECRET_ACCESS_KEY
            Secret = "YOUR_DO_SPACE_SECRET_KEY"

            # 启用 SSL
            # 环境变量覆盖：ATHENS_MINIO_USE_SSL
            EnableSSL = true

            # DO Spaces 存储中的空间名称
            # 环境变量覆盖：ATHENS_MINIO_BUCKET_NAME
            Bucket = "YOUR_DO_SPACE_NAME"

            # DO Spaces 存储的区域
            # 环境变量覆盖：ATHENS_MINIO_REGION
            Region = "YOUR_DO_SPACE_REGION"

#### 阿里云 OSS

为了让 Athens 与 [阿里云对象存储服务](https://www.alibabacloud.com/product/oss) 通信，我们使用 Minio 驱动。
此外，此存储的配置与我们的代理中的 [Minio](#minio) 配置几乎相同。

##### 配置：

    # StorageType 设置代理将使用的存储后端类型。
    # 环境变量覆盖：ATHENS_STORAGE_TYPE
    StorageType = "minio"

    [Storage]
        [Storage.Minio]
            # 阿里云 OSS 存储的地址
            # 环境变量覆盖：ATHENS_MINIO_ENDPOINT
            Endpoint = "YOUR_ADDRESS.aliyuncs.com"

            # Minio 存储的访问密钥
            # 环境变量覆盖：ATHENS_MINIO_ACCESS_KEY_ID
            Key = "YOUR_OSS_KEY_ID"

            # 阿里云 OSS 存储的秘密密钥
            # 环境变量覆盖：ATHENS_MINIO_SECRET_ACCESS_KEY
            Secret = "YOUR_OSS_SECRET_KEY"

            # 启用 SSL
            # 环境变量覆盖：ATHENS_MINIO_USE_SSL
            EnableSSL = true

            # 阿里云 OSS 存储中的父文件夹
            # 环境变量覆盖：ATHENS_MINIO_BUCKET_NAME
            Bucket = "YOUR_OSS_FOLDER_PREFIX"

## Azure Blob Storage

此驱动使用 [Azure Blob Storage](https://azure.microsoft.com/services/storage/blobs/)

>如果您从未使用过 Azure Blob Storage，这里有一个[快速入门](https://aka.ms/azureblob-quickstart)

它假定您已拥有以下内容：

- [一个 Azure 存储账户](https://docs.microsoft.com/azure/storage/common/storage-account-overview?toc=%2fazure%2fstorage%2fblobs%2ftoc.json)
- [凭证（存储账户密钥）](https://docs.microsoft.com/rest/api/storageservices/authorize-with-shared-key)
- 一个容器（用于存储 blob）


##### 配置：

    # StorageType 设置代理将使用的存储后端类型。
    # 环境变量覆盖：ATHENS_STORAGE_TYPE
    StorageType = "azureblob"

    [Storage]
        [Storage.AzureBlob]
            # Azure Blob 的存储账户名称
            # 环境变量覆盖：ATHENS_AZURE_ACCOUNT_NAME
            AccountName = "MY_AZURE_BLOB_ACCOUNT_NAME"

            # 与存储账户一起使用的账户密钥
            # 环境变量覆盖：ATHENS_AZURE_ACCOUNT_KEY
            AccountKey = "MY_AZURE_BLOB_ACCOUNT_KEY"

            # 与存储账户一起使用的托管标识资源 Id
            # 环境变量覆盖：ATHENS_AZURE_MANAGED_IDENTITY_RESOURCE_ID
            ManagedIdentityResourceId = "MY_AZURE_MANAGED_IDENTITY_RESOURCE_ID"

            # 与存储账户一起使用的存储资源
            # 环境变量覆盖：ATHENS_AZURE_STORAGE_RESOURCE
            StorageResource = "MY_AZURE_STORAGE_RESOURCE"

            # blob 存储中的容器名称
            # 环境变量覆盖：ATHENS_AZURE_CONTAINER_NAME
            ContainerName = "MY_AZURE_BLOB_CONTAINER_NAME"

## 外部存储

外部存储让 Athens 连接到您自己的存储后端实现。
您只需实现 [storage.Backend](https://github.com/gomods/athens/blob/main/pkg/storage/backend.go#L4) 接口，并将其运行在 http 服务器后面。

实现后端服务器后，您必须将 Athens 配置为使用该存储后端，如下所示：

##### 配置：
    # 环境变量覆盖：ATHENS_STORAGE_TYPE
    StorageType = "external"

    [Storage]
        [Storage.External]
            # 环境变量覆盖：ATHENS_EXTERNAL_STORAGE_URL
            URL = "http://localhost:9090"

Athens 提供了一个便捷包装器，让您可以轻松实现存储后端。请参阅以下示例：


```golang
package main

import (
    "github.com/gomods/athens/pkg/storage"
    "github.com/gomods/athens/pkg/storage/external"
)

// TODO: 实现 storage.Backend
type myCustomStorage struct {
    storage.Backend
}

func main() {
    handler := external.NewServer(&myCustomStorage{})
    http.ListenAndServe(":9090", handler)
}
```

## 运行多个指向同一存储的 Athens

Athens 能够同时运行多个指向同一存储介质的实例，使用称为"单飞（single flight）"的分布式锁定机制。

默认情况下，Athens 配置为使用 `memory` 单飞，它将锁存储在本地内存中。在运行单个 Athens 实例时这很有效，因为进程可以访问其自己的内存。但是，当运行多个指向同一存储的 Athens 实例时，需要分布式锁定机制。

Athens 支持多种分布式锁定机制：

- `etcd`
- `redis`
- `redis-sentinel`
- `gcp`（使用 `gcp` 存储类型时可用）
- `azureblob`（使用 `azureblob` 存储类型时可用）

设置 `SingleFlightType`（或环境中的 `ATHENS_SINGLE_FLIGHT TYPE`）配置值将启用上述一种机制。`azureblob` 和 `gcp` 类型不需要额外配置。

### 使用 etcd 作为单飞机制

使用 `etcd` 机制非常简单，只需一个逗号分隔的 etcd 端点列表。
推荐配置是 3 个端点，但可以使用更多。

    SingleFlightType = "etcd"

    [SingleFlight]
        [SingleFlight.Etcd]
            # 环境变量覆盖：ATHENS_ETCD_ENDPOINTS
            Endpoints = "localhost:2379,localhost:22379,localhost:32379"

### 使用 redis 作为单飞机制

Athens 支持两种与 redis 通信的机制：直接连接，以及通过 redis sentinel 连接。

#### 直接连接 redis

使用 redis 直接连接很简单，只需要一个 `redis-server`。
您还可以选择指定密码以连接 redis 服务器。

    SingleFlightType = "redis"

    [SingleFlight]
        [SingleFlight.Redis]
            # 端点是单飞机制的 redis 端点
            # 环境变量覆盖：ATHENS_REDIS_ENDPOINT
            Endpoint = "127.0.0.1:6379"

            # 密码是 redis 实例的密码
            # 环境变量覆盖：ATHENS_REDIS_PASSWORD
            Password = ""

也支持通过 [redis url](https://github.com/redis/redis-specifications/blob/master/uri/redis.txt) 连接 Redis：

    SingleFlightType = "redis"

    [SingleFlight]
        [SingleFlight.Redis]
            # 端点是单飞机制的 redis 端点
            # 环境变量覆盖：ATHENS_REDIS_ENDPOINT
            # 注意，如果需要 TLS，请使用 rediss://。
            Endpoint = "redis://user:password@127.0.0.1:6379/0?protocol=3"

如果 redis url 无效或无法解析，Athens 将回退到将 `Endpoint` 视为正常的 `host:port` 对。如果在 redis url 中提供了密码，并且也在 `Password` 配置选项中提供，则两个值必须匹配，否则 Athens 将无法启动。

##### 自定义锁配置：
如果您想自定义分布式锁选项，可以选择覆盖默认锁配置以更好地适应您的用例：

    [SingleFlight.Redis]
        ...
        [SingleFlight.Redis.LockConfig]
            # 锁的 TTL（以秒为单位）。默认为 900 秒（15 分钟）。
            # 环境变量覆盖：ATHENS_REDIS_LOCK_TTL
            TTL = 900
            # 获取锁的超时时间（以秒为单位）。默认为 15 秒。
            # 环境变量覆盖：ATHENS_REDIS_LOCK_TIMEOUT
            Timeout = 15
            # 获取锁时的最大重试次数。默认为 10。
            # 环境变量覆盖：ATHENS_REDIS_LOCK_MAX_RETRIES
            MaxRetries = 10

在某些情况下可能需要自定义，例如，如果您的场景通常需要超过 5 分钟来获取模块，您可以设置更高的 TTL。

#### 通过 redis sentinel 连接

**注意**：redis-sentinel 需要 redis 的工作知识，不推荐给所有人。

redis sentinel 是 redis 的高可用性设置，它提供自动监控、复制、故障转移和多个 redis 服务器在主从设置中的配置。它比运行单个 redis 服务器更复杂，需要多个分散的 redis 实例分布在各个节点上。

有关 redis-sentinel 的更多详细信息，请参阅[文档](https://redis.io/topics/sentinel)

由于 redis-sentinel 是 redis 的更复杂设置，它需要比标准 redis 更多的配置。

必需的配置：

- `Endpoints` 是要连接的 redis-sentinel 端点列表，通常是 3 个，但可以使用更多
- `MasterName` 是命名的主实例，配置在 redis-sentinel [配置](https://redis.io/topics/sentinel#configuring-sentinel) 中

与 `redis` 一样，您也可以选择指定密码以连接 `redis-sentinel` 端点：

    SingleFlightType = "redis-sentinel"

    [SingleFlight]
      [SingleFlight.RedisSentinel]
          # Endpoints 是用于发现 redis 主实例以获取 SingleFlight 锁的 redis sentinel 端点。
          # 环境变量覆盖：ATHENS_REDIS_SENTINEL_ENDPOINTS
          Endpoints = ["127.0.0.1:26379"]
          # MasterName 是 redis sentinel 主名称，用于发现 SingleFlight 锁的主实例。
          MasterName = "redis-1"
          # SentinelPassword 是用于与 redis sentinel 进行认证的可选密码。
          SentinelPassword = "sekret"

分布式锁选项也可以为 redis sentinel 自定义，方式与上述 redis 描述的类似。


### 使用 GCP 作为单飞机制

GCP 单飞机制不需要配置，开箱即用。它有一个可自定义的选项：

    [SingleFlight.GCP]
        # 用于判断进行中的 GCP 上传是否未能解锁的等待时间阈值（以秒为单位）。
        StaleThreshold = 120
