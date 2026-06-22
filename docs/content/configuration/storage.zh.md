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
- [Mongo](#mongo)
      - [配置：](#配置-2)
- [Google Cloud Storage](#google-cloud-storage)
      - [配置：](#配置-3)
- [AWS S3](#aws-s3)
      - [配置：](#配置-4)
- [Minio](#minio)
      - [配置：](#配置-5)
    - [DigitalOcean Spaces](#digitalocean-spaces)
      - [配置：](#配置-6)
    - [Alibaba OSS](#alibaba-oss)
      - [配置：](#配置-7)
- [Azure Blob Storage](#azure-blob-storage)
      - [配置：](#配置-8)
- [外部存储](#外部存储)
      - [配置：](#配置-9)
- [多个 Athens 实例指向同一存储](#多个-athens-实例指向同一存储)
  - [使用 etcd 作为单一flight机制](#使用-etcd-作为单一flight机制)
  - [使用 redis 作为单一flight机制](#使用-redis-作为单一flight机制)
    - [直接连接 redis](#直接连接-redis)
    - [通过 redis sentinel 连接 redis](#通过-redis-sentinel-连接-redis)

所有存储类型均可在 `config.toml` 文件中配置。你需要在 `StorageType` 中设置有效的存储驱动，也可通过环境变量 `ATHENS_STORAGE_TYPE` 在服务器上进行设置。此外，大多数驱动需要提供额外的配置，下文将详细说明。

## 内存

此存储类型无需特定配置，也是 Athens 的默认存储。所有数据将写入本地磁盘的 `tmp` 目录。

**此存储类型仅适用于开发环境！**

##### 配置：
```toml
# StorageType 设置代理将使用的存储后端类型。
# 环境变量覆盖：ATHENS_STORAGE_TYPE
StorageType = "memory"
```
## 磁盘

磁盘存储允许将模块存储在文件系统中。可以配置模块在磁盘上的存储位置。

> 若要支持离线 Athens 部署，可以先初始化磁盘存储。请参阅[此处](/configuration/prefill-disk-cache)获取相关说明。

##### 配置：
```toml
# StorageType 设置代理将使用的存储后端类型。
# 环境变量覆盖：ATHENS_STORAGE_TYPE
StorageType = "disk"

[Storage]
    [Storage.Disk]
        RootPath = "/path/on/disk"
```
其中 `/path/on/disk` 是存储路径。也可使用 `ATHENS_DISK_STORAGE_ROOT` 环境变量设置。

## Mongo

此驱动使用 [Mongo](https://www.mongodb.com/) 服务器作为数据存储。启动时，驱动会在 Mongo 服务器上创建 `athens` 数据库和 `module` 集合。

##### 配置：
```toml
# StorageType 设置代理将使用的存储后端类型。
# 环境变量覆盖：ATHENS_STORAGE_TYPE
StorageType = "mongo"

[Storage]
    [Storage.Mongo]
        # Mongo 存储的完整 URL
        # 环境变量覆盖：ATHENS_MONGO_STORAGE_URL
        URL = "mongodb://127.0.0.1:27017"

        # 可选参数
        # Mongo 连接使用的证书路径
        # 环境变量覆盖：ATHENS_MONGO_CERT_PATH
        CertPath = "/path/to/cert/file"

        # 可选参数
        # 允许不安全 SSL / http 连接 Mongo 存储
        # 仅用于测试或开发环境
        # 环境变量覆盖：ATHENS_MONGO_INSECURE
        Insecure = false

        # 可选参数
        # 指定自定义数据库
        # 环境变量覆盖：ATHENS_MONGO_DEFAULT_DATABASE
        DefaultDBName = "athens"

        # 可选参数
        # 指定自定义集合
        # 环境变量覆盖：ATHENS_MONGO_DEFAULT_COLLECTION
        DefaultCollectionName = "modules"
```
## Google Cloud Storage

此驱动使用 [Google Cloud Storage](https://cloud.google.com/storage/)，假定你已拥有 `account` 和 `bucket`。
若从未使用过 Google Cloud Storage，可以参阅这份[快速指南](https://cloud.google.com/storage/docs/creating-buckets#storage-create-bucket-console)了解如何创建 `bucket`。

##### 配置：
```toml
# StorageType 设置代理将使用的存储后端类型。
# 环境变量覆盖：ATHENS_STORAGE_TYPE
StorageType = "gcp"

[Storage]
    [Storage.GCP]
        # GCP Storage 使用的项目 ID
        # 环境变量覆盖：GOOGLE_CLOUD_PROJECT
        ProjectID = "YOUR_GCP_PROJECT_ID"

        # GCP Storage 使用的存储桶
        # 环境变量覆盖：ATHENS_STORAGE_GCP_BUCKET
        Bucket = "YOUR_GCP_BUCKET"
```

## AWS S3

此驱动使用 [AWS S3](https://aws.amazon.com/s3/)，假定你已创建 `account` 和 `bucket`。
若从未使用过 Amazon Web Services，可以参阅这份[快速指南](https://docs.aws.amazon.com/AmazonS3/latest/gsg/GetStartedWithS3.html)了解如何创建 `bucket`。
之后可在 `config.toml` 中配置凭证。如未在 `config.toml` 中指定访问密钥 ID 和秘密访问密钥，此驱动将尝试从 [AWS CLI 配置文件](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html)（安装期间创建）中加载默认配置文件获取凭证。

##### 配置：
```toml
# StorageType 设置代理将使用的存储后端类型。
# 环境变量覆盖：ATHENS_STORAGE_TYPE
StorageType = "s3"

[Storage]
    [Storage.S3]
    ### 认证模型如下，按优先级排序
    ### 若指定了 AWS_CREDENTIALS_ENDPOINT 且返回有效结果，则使用它
    ### 若指定了配置变量且有效，则使用它们
    ### 否则，将使用如下默认配置
    # 尝试从环境变量、共享配置文件（~/.aws/credentials）和 ec2 实例角色
    # 凭证中获取凭证。参见
    # https://godoc.org/github.com/aws/aws-sdk-go#hdr-Configuring_Credentials
    # 和
    # https://godoc.org/github.com/aws/aws-sdk-go/aws/session#hdr-Environment_Variables
    # 获取影响 aws 配置的环境变量。
    # 设置 UseDefaultConfiguration 将仅使用默认配置。该选项将在后续版本中弃用，
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
    # 可选参数
    # 环境变量覆盖：AWS_SESSION_TOKEN
    Token = ""

    # 用于存储的 S3 存储桶
    # 环境变量覆盖：ATHENS_S3_BUCKET_NAME
    Bucket = "MY_S3_BUCKET_NAME"

    # 若为 true，则使用 S3 端点的路径风格的 URL
    # 环境变量覆盖：AWS_FORCE_PATH_STYLE
    ForcePathStyle = false

    # 若为 true，则使用默认的 aws 配置。这将
    # 尝试从环境变量、共享配置文件（~/.aws/credentials）和 ec2 实例角色
    # 凭证中获取凭证。参见
    # https://godoc.org/github.com/aws/aws-sdk-go#hdr-Configuring_Credentials
    # 和
    # https://godoc.org/github.com/aws/aws-sdk-go/aws/session#hdr-Environment_Variables
    # 获取影响 aws 配置的环境变量。
    # 环境变量覆盖：AWS_USE_DEFAULT_CONFIGURATION
    UseDefaultConfiguration = false

    # https://docs.aws.amazon.com/sdk-for-go/api/aws/credentials/endpointcreds/
    # 注意：设置 AwsContainerCredentialsRelativeURI 时，URI 不要以 / 结尾
    # 环境变量覆盖：AWS_CREDENTIALS_ENDPOINT
    CredentialsEndpoint = ""

    # 环境变量覆盖：AWS_CONTAINER_CREDENTIALS_RELATIVE_URI
    # 若计划使用 AWS Fargate，请使用 http://169.254.170.2 作为 CredentialsEndpoint
    # 参考：https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-metadata-endpoint-v2.html
    AwsContainerCredentialsRelativeURI = ""

    # 可选的端点 URL（仅主机名或完整 URI）
    # 覆盖 S3 存储客户端生成的默认端点。
    #
    # 指定端点时仍需提供 `Region` 值。
    # 环境变量覆盖：AWS_ENDPOINT
    Endpoint = ""
```

## Minio

[Minio](https://www.minio.io/) 是一个开源对象存储服务器，提供 S3 兼容块存储的接口。若从未使用过 Minio，可以阅读这份[快速入门指南](https://docs.minio.io/)。Athens 通过 Minio 接口支持任何 S3 兼容的对象存储。下文提供了 Minio 的不同配置选项。Digital Ocean 和 Alibaba OSS 块存储的示例配置见下方。

##### 配置：
```toml
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

        # 启用 Minio 连接的 SSL
        # 默认为 true
        # 环境变量覆盖：ATHENS_MINIO_USE_SSL
        EnableSSL = false

        # 用于存储的 Minio 存储桶
        # 环境变量覆盖：ATHENS_MINIO_BUCKET_NAME
        Bucket = "gomods"
```

#### DigitalOcean Spaces

为了让 Athens 与 [DigitalOcean Spaces](https://www.digitalocean.com/products/spaces/) 通信，我们使用 Minio 驱动，因为 DO Spaces 与其[完全兼容](https://developers.digitalocean.com/documentation/spaces/)。此存储的配置与代理中的 [Minio](#minio) 配置几乎相同。

##### 配置：
```toml
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

        # DO Spaces 存储中的 Space 名称
        # 环境变量覆盖：ATHENS_MINIO_BUCKET_NAME
        Bucket = "YOUR_DO_SPACE_NAME"

        # DO Spaces 存储的区域
        # 环境变量覆盖：ATHENS_MINIO_REGION
        Region = "YOUR_DO_SPACE_REGION"
```

#### Alibaba OSS

为了让 Athens 与 [Alibaba Cloud Object Storage Service](https://www.alibabacloud.com/product/oss) 通信，我们使用 Minio 驱动。此存储的配置与代理中的 [Minio](#minio) 配置几乎相同。

##### 配置：
```toml
# StorageType 设置代理将使用的存储后端类型。
# 环境变量覆盖：ATHENS_STORAGE_TYPE
StorageType = "minio"

[Storage]
    [Storage.Minio]
        # Alibaba OSS 存储的地址
        # 环境变量覆盖：ATHENS_MINIO_ENDPOINT
        Endpoint = "YOUR_ADDRESS.aliyuncs.com"

        # Minio 存储的访问密钥
        # 环境变量覆盖：ATHENS_MINIO_ACCESS_KEY_ID
        Key = "YOUR_OSS_KEY_ID"

        # Alibaba OSS 存储的秘密密钥
        # 环境变量覆盖：ATHENS_MINIO_SECRET_ACCESS_KEY
        Secret = "YOUR_OSS_SECRET_KEY"

        # 启用 SSL
        # 环境变量覆盖：ATHENS_MINIO_USE_SSL
        EnableSSL = true

        # Alibaba OSS 存储中的父文件夹
        # 环境变量覆盖：ATHENS_MINIO_BUCKET_NAME
        Bucket = "YOUR_OSS_FOLDER_PREFIX"
```

## Azure Blob Storage

此驱动使用 [Azure Blob Storage](https://azure.microsoft.com/services/storage/blobs/)

> 若从未使用过 Azure Blob Storage，可以参阅这份[快速入门指南](https://aka.ms/azureblob-quickstart)

它假定你已具备以下条件：

- [一个 Azure 存储账户](https://docs.microsoft.com/azure/storage/common/storage-account-overview?toc=%2fazure%2fstorage%2fblobs%2ftoc.json)
- [凭证（存储账户密钥）](https://docs.microsoft.com/rest/api/storageservices/authorize-with-shared-key)
- 一个容器（用于存储数据块）


##### 配置：
```toml
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

        # 与存储账户一起使用的托管标识资源 ID
        # 环境变量覆盖：ATHENS_AZURE_MANAGED_IDENTITY_RESOURCE_ID
        ManagedIdentityResourceId = "MY_AZURE_MANAGED_IDENTITY_RESOURCE_ID"

        # 与存储账户一起使用的存储资源
        # 环境变量覆盖：ATHENS_AZURE_STORAGE_RESOURCE
        StorageResource = "MY_AZURE_STORAGE_RESOURCE"

        # Blob 存储中的容器名称
        # 环境变量覆盖：ATHENS_AZURE_CONTAINER_NAME
        ContainerName = "MY_AZURE_BLOB_CONTAINER_NAME"
```

## 外部存储

外部存储让 Athens 连接到自定义的存储后端实现。只需实现 [storage.Backend](https://github.com/gomods/athens/blob/main/pkg/storage/backend.go#L4) 接口，并在 HTTP 服务器上运行它。

实现后端服务器后，需按如下方式配置 Athens 使用该存储后端：

##### 配置：
```toml
# 环境变量覆盖：ATHENS_STORAGE_TYPE
StorageType = "external"

[Storage]
    [Storage.External]
        # 环境变量覆盖：ATHENS_EXTERNAL_STORAGE_URL
        URL = "http://localhost:9090"
```

Athens 提供了便捷的包装器，方便实现存储后端。请参见以下示例：


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

## 多个 Athens 实例指向同一存储

Athens 支持通过名为 `single flight` 的分布式锁机制，让多个实例并发运行在同一存储介质上。

默认情况下，Athens 配置使用 `memory` single flight，将锁存储在本地内存中。运行单个 Athens 实例时可正常工作，但在同一存储上运行多个实例时，需要分布式锁机制。

Athens 支持多种分布式锁机制：

- `etcd`
- `redis`
- `redis-sentinel`
- `gcp`（使用 `gcp` 存储类型时可用）
- `azureblob`（使用 `azureblob` 存储类型时可用）

设置 `SingleFlightType`（或环境变量 `ATHENS_SINGLE_FLIGHT_TYPE`）配置值即可启用上述某种机制。`azureblob` 和 `gcp` 类型无需额外配置。

### 使用 etcd 作为 single flight 实现

使用 `etcd` 实现 single flight 非常简单，只需提供逗号分隔的 etcd 端点列表。建议配置 3 个端点，也可使用更多。
```toml
SingleFlightType = "etcd"

[SingleFlight]
    [SingleFlight.Etcd]
        # 环境变量覆盖：ATHENS_ETCD_ENDPOINTS
        Endpoints = "localhost:2379,localhost:22379,localhost:32379"
```

### 使用 redis 作为单一flight机制

Athens 支持两种与 redis 通信的方式：直接连接，以及通过 redis sentinel 连接。

#### 直接连接 redis

使用直接连接 redis 很简单，只需一个 `redis-server`，也可指定密码。

```toml
SingleFlightType = "redis"

[SingleFlight]
    [SingleFlight.Redis]
        # 端点是 single flight 机制的 redis 端点
        # 环境变量覆盖：ATHENS_REDIS_ENDPOINT
        Endpoint = "127.0.0.1:6379"

        # 密码是 redis 实例的密码
        # 环境变量覆盖：ATHENS_REDIS_PASSWORD
        Password = ""
```

也支持通过 [redis url](https://github.com/redis/redis-specifications/blob/master/uri/redis.txt) 连接 Redis：

```toml
SingleFlightType = "redis"

[SingleFlight]
    [SingleFlight.Redis]
        # 端点是 single flight 机制的 redis 端点
        # 环境变量覆盖：ATHENS_REDIS_ENDPOINT
        # 注意：如需 TLS，请使用 rediss:// 而不是 redis://。
        Endpoint = "redis://user:password@127.0.0.1:6379/0?protocol=3"
```

若 redis URL 无效或无法解析，Athens 将回退到将 `Endpoint` 视为普通的 `host:port` 对。若在 redis URL 中提供了密码，同时也在 `Password` 配置选项中提供，则这两个值必须相同，否则 Athens 将无法启动。

##### 自定义锁配置：
如需自定义分布式锁，可覆盖默认锁配置来更好适应使用场景：
```toml
[SingleFlight.Redis]
    #...
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
```

某些情况下可能需要自定义，例如获取模块通常需要超过 5 分钟时，可设置更高的 TTL。

#### 通过 redis sentinel 连接 redis

**注意**：redis-sentinel 需要 Redis 的相关知识，不建议所有人都使用。

redis sentinel 是 Redis 的高可用方案，提供自动监控、复制、故障转移和多 Redis 服务器的主从配置。它比运行单个 Redis 服务器更复杂，需要在分布式节点上运行多个 Redis 实例。

有关 redis-sentinel 的更多详细信息，请参阅[文档](https://redis.io/topics/sentinel)

由于 redis-sentinel 是更复杂的一系列 Redis，需要比标准 Redis 更多的配置。

必需的配置：

- `Endpoints`：要连接的 redis-sentinel 端点列表，通常为 3 个，也可使用更多
- `MasterName`：主节点名称，如 `redis-sentinel` [配置](https://redis.io/topics/sentinel#configuring-sentinel)中所设置

与 `redis` 一样，也可指定密码连接 `redis-sentinel` 端点
```toml
SingleFlightType = "redis-sentinel"

[SingleFlight]
  [SingleFlight.RedisSentinel]
      # Endpoints 是用于发现 single flight 锁的 redis 主节点的 redis sentinel 端点。
      # 环境变量覆盖：ATHENS_REDIS_SENTINEL_ENDPOINTS
      Endpoints = ["127.0.0.1:26379"]
      # MasterName 是用于发现 single flight 锁的主节点的 redis sentinel 主节点名称
      MasterName = "redis-1"
      # SentinelPassword 是与 redis sentinel 认证的可选密码
      SentinelPassword = "sekret"
```

redis sentinel 的分布式锁选项也可自定义，方式与上述 redis 类似。

### 使用 GCP 作为单一flight机制

GCP singleflight 机制无需配置，开箱即用。它有一个可自定义的选项：
```toml
[SingleFlight.GCP]
    # 阈值（以秒为单位），用于判断进行中的 GCP 上传是否已失败而未解锁。
    StaleThreshold = 120
```
