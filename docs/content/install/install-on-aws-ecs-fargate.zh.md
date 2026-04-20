---
title: "在 AWS Fargate (ECS) 上安装"
date: 2021-03-27T21:53:51-07:00
draft: false
weight: 3
---

本文档将展示如何使用 [AWS Fargate (ECS)](https://docs.aws.amazon.com/AmazonECS/latest/userguide/what-is-fargate.html) 来运行 Athens 代理。

---

## 选择存储提供商

有关如何使用环境变量配置各种存储提供商的文档已存在。然而，对于这个示例，我们将使用 [Amazon S3 Storage](https://aws.amazon.com/s3/)（s3）。

## 准备工作

本指南假设您已经拥有 AWS 账户，以及在账户中创建资源所需的必要认证和权限。

无论您选择使用 [awscli](https://aws.amazon.com/cli/) 还是 [Terraform](https://www.terraform.io/) 等工具来创建资源，所需的资源都是相同的。

## S3 存储桶

为了持久化模块，我们需要创建一个 S3 存储桶。

以下是使用 [awscli](https://aws.amazon.com/cli/) 和 [Terraform](https://www.terraform.io/) 创建 S3 存储桶的两个示例。

`awscli`：
```shell
$ aws s3api create-bucket --bucket athens-proxy-us-east-1-123456789012 --region us-east-1
```

`terraform`：
```terraform
resource "aws_s3_bucket" "cache" {
  bucket = "athens-proxy-us-east-1-123456789012"
}
```

_注意：由于 S3 的[全局唯一命名规则](https://docs.aws.amazon.com/AmazonS3/latest/userguide/bucketnamingrules.html)，建议将环境、区域和/或账户 ID 作为存储桶名称的组成部分。_

## ECS 任务 IAM 角色

为了让 ECS 容器实例能够使用 S3 存储桶，我们需要配置任务 IAM 角色以包含适当的 `allow` 规则。

以下是两份最小特权策略文档（JSON 和 Terraform 格式），用于启用 ECS 容器 S3 存储桶访问，以存储和检索缓存资源。

`json`：
```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "s3:ListBucket",
                "s3:GetBucketLocation"
            ],
            "Resource": "arn:aws:s3:::athens-proxy-us-east-1-123456789012"
        },
        {
            "Effect": "Allow",
            "Action": [
                "s3:PutObject",
                "s3:GetObject",
                "s3:DeleteObject"
            ],
            "Resource": "arn:aws:s3:::athens-proxy-us-east-1-123456789012/*"
        }
    ]
}
```

`terraform`：
```terraform
resource "aws_iam_policy" "task_role" {
  name   = "athens-proxy-task-role"
  path   = "/"
  policy = data.aws_iam_policy_document.task_role_policy.json
}

data "aws_iam_policy_document" "task_role_policy" {
  statement {
    effect = "Allow"
    actions = [
      "s3:ListBucket",
      "s3:GetBucketLocation"
    ]
    resources = [aws_s3_bucket.cache.arn]
  }
  statement {
    effect = "Allow"
    actions = [
      "s3:PutObject",
      "s3:GetObject",
      "s3:DeleteObject"
    ]
    resources = ["${aws_s3_bucket.cache.arn}/*"]
  }
  statement {
    effect = "Allow"
    actions = [
      "sts:AssumeRole",
      "sts:TagSession"
    ]
    resources = ["*"]
  }
}
```

## ECS 任务定义

为了让 Athens 能够向 S3 存储桶进行认证，我们需要配置与 S3 关联的存储变量。

以下是任务定义中的一个摘录，展示了所需的最小环境变量。

```json
"environment": [
  {"name": "AWS_REGION", "value": "us-east-1"},
  {"name": "AWS_USE_DEFAULT_CONFIGURATION", "value": "true"},
  {"name": "ATHENS_STORAGE_TYPE", "value": "s3"},
  {"name": "ATHENS_S3_BUCKET_NAME", "value": "athens-proxy-us-east-1-123456789012"},
]
```
