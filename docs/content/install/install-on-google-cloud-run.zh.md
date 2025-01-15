---
title: "在 Google Cloud Run 上安装 Athens"
description: "在 Google Cloud Run 上安装 Athens 代理服务器的指南"
date: 2023-08-01
weight: 3
---

# 在 Google Cloud Run 上安装 Athens

本指南将引导您完成在 Google Cloud Run 上设置 Athens 代理服务器的过程。

## 先决条件

- Google Cloud 账户
- 已安装 gcloud CLI
- 已启用 Cloud Run API
- 已启用 Artifact Registry API

## 步骤 1: 创建 Artifact Registry 仓库

```bash
gcloud artifacts repositories create athens-repo \
    --repository-format=docker \
    --location=us-central1 \
    --description="Athens Docker repository"
```

## 步骤 2: 构建并推送 Docker 镜像

```bash
# 克隆 Athens 仓库
git clone https://github.com/gomods/athens.git
cd athens

# 构建 Docker 镜像
docker build -t us-central1-docker.pkg.dev/YOUR_PROJECT_ID/athens-repo/athens:latest .

# 推送镜像到 Artifact Registry
docker push us-central1-docker.pkg.dev/YOUR_PROJECT_ID/athens-repo/athens:latest
```

## 步骤 3: 部署到 Cloud Run

```bash
gcloud run deploy athens \
    --image=us-central1-docker.pkg.dev/YOUR_PROJECT_ID/athens-repo/athens:latest \
    --region=us-central1 \
    --platform=managed \
    --allow-unauthenticated
```

## 步骤 4: 配置 Go 使用 Athens

设置 GOPROXY 环境变量：

```bash
export GOPROXY=http://YOUR_CLOUD_RUN_URL
```

## 后续步骤

- 配置存储后端（推荐使用 Google Cloud Storage）
- 设置身份验证
- 配置自动扩展

## 故障排除

- 检查 Cloud Run 日志
- 验证网络连接
- 确保正确的 IAM 权限

## 参考文档

- [Athens 官方文档](https://docs.gomods.io)
- [Google Cloud Run 文档](https://cloud.google.com/run/docs)