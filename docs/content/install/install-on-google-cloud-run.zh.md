---
title: "在 Google Cloud Run 上安装"
date: 2019-10-10T18:16:43+11:00
draft: false
weight: 4
---

[Google Cloud Run](https://cloud.google.com/run/) 是一项旨在弥合无服务器架构的维护优势与 Kubernetes 灵活性之间差距的服务。它建立在开源 [Knative](https://knative.dev/) 项目之上。使用 Cloud Run 部署与使用 [Google App Engine](/install/install-on-gae) 部署类似，但具有免费套餐和更简单构建过程的优势。

## 选择存储提供商

有关如何使用环境变量配置各种存储提供商的文档已存在；但是，对于这个示例，我们将使用 [Google Cloud Storage](https://cloud.google.com/storage/)(GCS)，因为它与 Cloud Run 配合得很好。

## 准备工作

本指南假设您已完成以下任务：

- 已注册 Google Cloud
- 已安装 [gcloud](https://cloud.google.com/sdk/install) 命令行工具
- 已安装 gcloud 命令行工具的 beta 插件（[设置方法如下](https://cloud.google.com/run/docs/setup)）
- 已为您的 Go 模块创建了一个（GCS）存储桶

### 设置 GCS 存储桶

如果您还没有 GCS 存储桶，可以使用 [gsutil 工具](https://cloud.google.com/storage/docs/gsutil)进行设置。

首先选择一个您希望存储所在的[区域](https://cloud.google.com/about/locations/?tab=americas)，然后使用以下命令在与您的区域和存储桶名称对应的位置创建一个存储桶。

```console
$ gsutil mb -l europe-west-4 gs://some-bucket
```

## 设置

将这些环境变量的值更改为适合您的应用程序的值。对于 `GOOGLE_CLOUD_PROJECT`，这需要是包含您的 Cloud Run 部署的项目名称。`ATHENS_REGION` 应该是您的 Cloud Run 实例所在的[区域](https://cloud.google.com/about/locations/?tab=americas)，而 `GCS_BUCKET` 应该是 Athens 用于存储模块代码和元数据的 Google Cloud Storage 存储桶。

```console
$ export GOOGLE_CLOUD_PROJECT=your-project
$ export ATHENS_REGION=asia-northeast1
$ export GCS_BUCKET=your-bucket-name
$ gcloud auth login
$ gcloud auth configure-docker
```

然后您需要将 Athens Docker 镜像的副本推送到您的 Google Cloud 容器注册表。

以下是一个使用 v0.11.0 的示例，要获取最新版本，请查看[最新 Athens 版本](https://github.com/gomods/athens/releases)
```console
$ docker pull gomods/athens:v0.11.0

$ docker tag gomods/athens:v0.11.0 gcr.io/$GOOGLE_CLOUD_PROJECT/athens:v0.11.0

$ docker push gcr.io/$GOOGLE_CLOUD_PROJECT/athens:v0.11.0
```

一旦您在注册表中有了容器镜像，就可以使用 `gcloud` 来配置您的 Athens 实例。

```console
$ gcloud beta run deploy \
    --image gcr.io/$GOOGLE_CLOUD_PROJECT/athens:v0.11.0 \
    --platform managed \
    --region $ATHENS_REGION \
    --allow-unauthenticated \
    --set-env-vars=ATHENS_STORAGE_TYPE=gcp \
    --set-env-vars=GOOGLE_CLOUD_PROJECT=$GOOGLE_CLOUD_PROJECT \
    --set-env-vars=ATHENS_STORAGE_GCP_BUCKET=$GCS_BUCKET \
    athens
```

此命令完成后会提供您实例的 URL，但您始终可以通过 CLI 找到它：

```console
$ gcloud beta run services describe athens --platform managed --region $ATHENS_REGION | grep hostname
```
