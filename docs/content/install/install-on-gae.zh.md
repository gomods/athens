---
title: "在 Google App Engine 上安装"
date: 2019-09-27T17:48:40+10:00
draft: false
weight: 4
---

[Google App Engine (GAE)](https://cloud.google.com/appengine/) 是一项 Google 服务，允许应用程序在无需配置底层硬件的情况下进行部署。它类似于[上一节](/install/install-on-aci)中介绍的 Azure Container Engine。本指南将演示如何在 GAE 上运行 Athens。

## 选择存储提供商

有关如何使用环境变量配置各种存储提供商的文档已存在；但是，对于这个示例，我们将使用 [Google Cloud Storage](https://cloud.google.com/storage/)(GCS)，因为它与 Cloud Run 配合得很好。

## 准备工作

本指南假设您已完成以下任务：

- 已注册 Google Cloud
- 已安装 [gcloud](https://cloud.google.com/sdk/install) 命令行工具

### 设置 GCS 存储桶

如果您还没有 GCS 存储桶，可以使用 [gsutil 工具](https://cloud.google.com/storage/docs/gsutil)进行设置。

首先选择一个您希望存储所在的[全球区域](https://cloud.google.com/about/locations/?tab=americas)，然后使用以下命令在与您的区域和存储桶名称对应的位置创建一个存储桶。

```console
$ gsutil mb -l europe-west-4 gs://some-bucket
```

## 设置

首先克隆 Athens 代码库

```console
$ git clone https://github.com/gomods/athens.git
```

已为您设置好了 Google Application Engine 脚手架。将其复制到一个新文件并修改环境变量。

```console
$ cd athens
$ cp scripts/gae/app.sample.yaml scripts/gae/app.yaml
$ code scripts/gae/app.yaml
```

配置好环境变量后，您可以将 Athens 部署为 GAE 服务。

```console
$ make deploy-gae
```
