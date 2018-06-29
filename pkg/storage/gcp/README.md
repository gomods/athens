# Google Cloud Storage Driver

This driver provides support for storing module files in Google Cloud storage.
You may host as little as the storage only on GCP, the entire project does not need to run there.

# Configuration

Minimal configuration is needed, just the name of a storage bucket and an authentication method for that project, and then tell Athens you want to use that as your storage medium.

## Driver Configuration

The only configuration for this driver other than authentication is an environment variable for the bucket name.
`ATHENS_STORAGE_GCP_BUCKET`

## Authentication

The only currently supported authentication type is a service account json key file.
For instructions on creating a new service account see [here](###)
This file is referenced via an environment variable `ATHENS_STORAGE_GCP_SA`, which should point to the json file.

## Athens Configuration

In order to tell Olympus to use GCP storage set `ATHENS_STORAGE_TYPE` to `gcp`.

# Contributing
