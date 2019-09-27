---
title: "Install on Google App Engine"
date: 2019-09-27T17:48:40+10:00
draft: true
weight: 4
---


[Google App Engine (GAE)](https://cloud.google.com/appengine/) is a Google service allows applications to be deployed without provisioning the underlying hardware. It is similar to Azure Container Engine which is covered in a [previous section](/install/install-on-aci). This guide will demonstrate how you can get Athens running on GAE.

## Selecting a Storage Provider

Athens currently supports a number of storage drivers. For quick and easy use on GAE, we recommend using the local disk provider. For more permanent use, we recommend using MongoDB or other more persistent infrastructure. For other providers, please see the [storage provider documentation](/configuration/storage).

## Before You Begin

This guide assumes you have completed the following tasks:

- Signed up for Google Cloud
- Installed the [gcloud](https://cloud.google.com/sdk/install) command line tool 

## Setup

First, create a directory. We will be adding two files to the directory. `app.yaml` will provide configuration for GAE, and `Dockerfile` will provide instructions on how to create an athens container.

Copy the following source into the files

```yaml
# app.yaml

# General settings
runtime: custom
env: flex
service: your-service-name

# Network settings
network:
  instance_tag: athens
  forwarded_ports:
    - 3000/tcp

# Compute settings
resources:
  cpu: 1
  memory_gb: 0.6
  disk_size_gb: 100

# Health and liveness check settings
liveness_check:
  path: "/healthz"
  check_interval_sec: 30
  failure_threshold: 2
  success_threshold: 2

readiness_check:
  path: "/readyz"
  check_interval_sec: 5
  failure_threshold: 2
  success_threshold: 2
  app_start_timeout_sec: 10

# Scaling instructions
automatic_scaling:
  min_num_instances: 1
  max_num_instances: 15
  cool_down_period_sec: 180
  cpu_utilization:
    target_utilization: 0.6

# Environment variables configuring athens
env_variables:
  ATHENS_STORAGE_TYPE: disk
  ATHENS_DISK_STORAGE_ROOT: /athens
```

```Dockerfile
# Dockerfile

FROM gomods/athens:v0.5.0

RUN mkdir /athens
```

You can now run `gcloud app deploy` to deploy Athens.