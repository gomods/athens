---
title: "Install on Google App Engine"
date: 2019-09-27T17:48:40+10:00
draft: false
weight: 4
---


[Google App Engine (GAE)](https://cloud.google.com/appengine/) is a Google service allows applications to be deployed without provisioning the underlying hardware. It is similar to Azure Container Engine which is covered in a [previous section](/install/install-on-aci). This guide will demonstrate how you can get Athens running on GAE.

## Selecting a Storage Provider

There is documentaion about how to use environment variables to configure a large number of storage providers; however, for this prarticular example we will use [Google Cloud Storage](https://cloud.google.com/storage/)(GCS) because it fits nicely with Cloud Run.

## Before You Begin

This guide assumes you have completed the following tasks:

- Signed up for Google Cloud
- Installed the [gcloud](https://cloud.google.com/sdk/install) command line tool 

## Setup

First clone the Athens repository

```bash
git clone https://github.com/gomods/athens.git
```

There is already a Google Application Engine scaffold set up for you. Copy it into a new file and make changes to the environment variables.

```bash
cd athens
cp scripts/gae/app.sample.yaml scripts/gae/app.yaml
code scripts/gae/app.yaml
```

Once you have configured the environment variables you can build a GAE service

```bash
make deploy-gae
```