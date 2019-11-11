---
title: "Install on Google Cloud Run"
date: 2019-10-10T18:16:43+11:00
draft: false
weight: 4
---

[Google Cloud Run](https://cloud.google.com/run/) is a service that aims to bridge the gap between the maintainance benefits of serverless architecture and the flexibility of Kubernetes. It is built on top of the opensource [Knative](https://knative.dev/) project. Deploying using Cloud Run is similar to deploying using [Google App Engine](/install/install-on-gae) with the benefits of a free tier and a simpler build process.

## Selecting a Storage Provider

There is documentaion about how to use environment variables to configure a large number of storage providers; however, for this prarticular example we will use [Google Cloud Storage](https://cloud.google.com/storage/)(GCS) because it fits nicely with Cloud Run.

## Before You Begin

This guide assumes you have completed the following tasks:

- Signed up for Google Cloud
- Installed the [gcloud](https://cloud.google.com/sdk/install) command line tool
- Installed the beta plugin for ghe gcloud command line tool ([this is how to set it up](https://cloud.google.com/run/docs/setup))
- Created a (GCS) bucket for your go modules

### Setup a GCS Bucket

If you do not already have GCS bucket you can set one up using the [gsutil tool](https://cloud.google.com/storage/docs/gsutil).

First select a [region](https://cloud.google.com/about/locations/?tab=americas) you would like to have your storage in. You can then create a bucket in that region using the following command substituting your in your region and bucket name.

```console
$ gsutil mb -l europe-west-4 gs://some-bucket
```

## Setup

Change the values of these environment variables to be appropriate for your application. For `GOOGLE_CLOUD_PROJECT`, this needs to be the name of the project that has your cloud run deployment in it. `ATHENS_REGION` should be the [region](https://cloud.google.com/about/locations/?tab=americas) that your cloud run instance will be in, and `GCS_BUCKET` should be the Google Cloud Storage bucket that Athens will store module code and metadata in..

```console
$ export GOOGLE_CLOUD_PROJECT=your-project
$ export ATHENS_REGION=asia-northeast1
$ export GCS_BUCKET=your-bucket-name
```

You will then need to push a copy of the Athens docker image to your google cloud container registry.

```console
$ docker pull gomods/athens:v0.6.0

$ docker tag gomods/athens:v0.6.0 gcr.io/$GOOGLE_CLOUD_PROJECT/athens:v0.6.0

$ gcloud builds submit --tag gcr.io/$GOOGLE_CLOUD_PROJECT/athens:v0.6.0
```

Once you have the container image in your registry you can use `gcloud` to provision your Athens instance.

```console
$ gcloud beta run deploy \
    --image gcr.io/$GOOGLE_CLOUD_PROJECT/athens:v0.6.0 \
    --platform managed \
    --region $ATHENS_REGION \
    --allow-unauthenticated \
    --set-env-vars=ATHENS_STORAGE_TYPE=gcp \
    --set-env-vars=GOOGLE_CLOUD_PROJECT=$GOOGLE_CLOUD_PROJECT \
    --set-env-vars=ATHENS_STORAGE_GCP_BUCKET=$GCS_BUCKET \
    athens
```

Once this command finishes is will provide a url to your instance, but you can always find this through the cli:

```console
$ gcloud beta run services describe athens --platform managed --region $ATHENS_REGION | grep hostname
```
