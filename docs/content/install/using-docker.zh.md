# 使用 Docker 安装 Athens

Athens 可以通过 Docker 快速部署和运行。本指南将介绍如何使用 Docker 安装和运行 Athens 代理服务器。

## 前提条件

- 已安装 Docker
- 已安装 Docker Compose（可选）

## 快速开始

1. 拉取 Athens Docker 镜像：

   ```bash
   docker pull gomods/athens:latest
   ```

2. 运行 Athens 容器：

   ```bash
   docker run -d -p 3000:3000 --name athens gomods/athens:latest
   ```

3. 验证 Athens 是否运行：

   ```bash
   curl http://localhost:3000
   ```

   你应该会看到 Athens 的欢迎页面。

## 使用 Docker Compose

对于更复杂的部署，可以使用 Docker Compose：

1. 创建 `docker-compose.yml` 文件：

   ```yaml
   version: '3'
   services:
     athens:
       image: gomods/athens:latest
       ports:
         - "3000:3000"
       volumes:
         - athens-storage:/var/lib/athens
       environment:
         - ATHENS_DISK_STORAGE_ROOT=/var/lib/athens
         - ATHENS_STORAGE_TYPE=disk
   volumes:
     athens-storage:
   ```

2. 启动服务：

   ```bash
   docker-compose up -d
   ```

## 配置存储

Athens 支持多种存储后端：

- 磁盘存储（默认）
- AWS S3
- Google Cloud Storage
- Azure Blob Storage
- Minio

在 Docker 中，可以通过环境变量配置存储类型。例如，要使用 S3 存储：

```bash
docker run -d -p 3000:3000 \
  -e ATHENS_STORAGE_TYPE=s3 \
  -e AWS_ACCESS_KEY_ID=your-access-key \
  -e AWS_SECRET_ACCESS_KEY=your-secret-key \
  -e AWS_REGION=your-region \
  -e ATHENS_S3_BUCKET=your-bucket-name \
  --name athens gomods/athens:latest
```

## 持久化存储

为了持久化存储模块，建议将存储目录挂载到主机：

```bash
docker run -d -p 3000:3000 \
  -v /path/to/local/storage:/var/lib/athens \
  --name athens gomods/athens:latest
```

## 下一步

- 配置 Go 客户端使用 Athens 代理
- 了解 Athens 的认证和授权机制
- 探索高级配置选项

## 故障排除

如果遇到问题，可以查看容器日志：

```bash
docker logs athens
```

或者以交互模式运行容器：

```bash
docker run -it --rm -p 3000:3000 --name athens gomods/athens:latest
```

这将允许你实时查看日志输出。