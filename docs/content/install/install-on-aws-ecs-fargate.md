---
title: "Install on AWS Fargate (ECS)"
date: 2021-03-27T21:53:51-07:00
draft: false
weight: 3
---

In this document, we'll show how to use [AWS Fargate (ECS)](https://docs.aws.amazon.com/AmazonECS/latest/userguide/what-is-fargate.html) to run the Athens proxy.

---

## Selecting a Storage Provider

There is documentation about how to use environment variables to configure the various storage providers. However, for 
this particular example we will use [Amazon S3 Storage](https://aws.amazon.com/s3/) (s3).

## Before You Begin

This guide assumes you already have an AWS account as well as the necessary authentication and permissions to create 
resources in the account.

Whether you choose to create your resources using the [awscli](https://aws.amazon.com/cli/) or use something like
[Terraform](https://www.terraform.io/), the resources required are the same.

## S3 Bucket

In order to persist modules, we will create a s3 bucket for storage.

Below are two examples of creating the s3 bucket using the [awscli](https://aws.amazon.com/cli/) and [Terraform](https://www.terraform.io/).

`awscli`:
```shell
$ aws s3api create-bucket --bucket athens-proxy-us-east-1-123456789012 --region us-east-1
```

`terraform`:
```terraform
resource "aws_s3_bucket" "cache" {
  bucket = "athens-proxy-us-east-1-123456789012"
}
```

_note: it is a good idea to use environment, region, and/or account ID as components to the bucket name due to their 
[globally unique naming rules](https://docs.aws.amazon.com/AmazonS3/latest/userguide/bucketnamingrules.html)._

## ECS Task IAM Role

In order for the ECS container instances to use the s3 bucket, we will need to configure the task IAM role to 
include the proper `allow` rules.

Below is a least-privileged policy document in both JSON and Terraform to enable ECS containers s3 bucket access to 
store and retrieve cache assets.

`json`:
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

`terraform`:
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

## ECS Task Definition

In order for Athens to be able to authenticate to the s3 bucket, we will need to configure the storage variables 
associated with s3.

Below is an excerpt from a task definition that shows the minimum environment variables needed.

```json
"environment": [
  {"name": "AWS_REGION", "value": "us-east-1"},
  {"name": "AWS_USE_DEFAULT_CONFIGURATION", "value": "true"},
  {"name": "ATHENS_STORAGE_TYPE", "value": "s3"},
  {"name": "ATHENS_S3_BUCKET_NAME", "value": "athens-proxy-us-east-1-123456789012"},
]
```