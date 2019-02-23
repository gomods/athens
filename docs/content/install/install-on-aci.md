---
title: "Install on Azure Container Instances"
date: 2018-12-06T13:17:37-08:00
draft: false
weight: 3
---

When you follow the instructions in the [Walkthrough](/walkthrough), you end up with an Athens Proxy that uses in-memory storage. This is only suitable for trying out the Athens proxy for a short period of time, as you will quickly run out of memory and Athens won't persist modules between restarts. This guide will help you get Athens running in a more suitable manner for scenarios like providing an instance for your development team to share.

In this document, we'll show how to use [Azure Container Instances](https://cda.ms/KR) (ACI) to run the Athens proxy.

## Selecting a Storage Provider

Athens currently supports a number of storage drivers. For quick and easy use on ACI, we recommend using the local disk provider. For more permanent use, we recommend using MongoDB or other more persistent infrastructure. For other providers, please see the [storage provider documentation](/configuration/storage/).

### Required Environment Variables

Before executing any of the commands below, make sure you have the following environment variables set up on your system:

- `AZURE_ATHENS_RESOURCE_GROUP` - The [Azure Resource Group](https://www.petri.com/what-are-microsoft-azure-resource-groups) to install the container in. You need to already have one of these before installing Athens
    - See [here](https://cda.ms/KS) for details on how to create a resource group
- `AZURE_ATHENS_CONTAINER_NAME` - The name of the container. This should be alphanumeric and you can have `-` and `_` characters
- `LOCATION` - The [Azure region](https://cda.ms/KT) to install the container in. See the previous link for an exhaustive list, but here's a useful cheat sheet that you can use immediately, without reading any docs:
    - North America: `eastus2`
    - Europe: `westeurope`
    - Asia: `southeastasia` 
- `AZURE_ATHENS_DNS_NAME` - The DNS name to assign to the container. It has to be globally unique inside of the region you set (`LOCATION`)


### Installing with the Disk Storage Driver

```console
az container create \
-g "${AZURE_ATHENS_RESOURCE_GROUP}" \
-n "${AZURE_ATHENS_CONTAINER_NAME}-${LOCATION}" \
--image gomods/athens:v0.2.0 \
-e "ATHENS_STORAGE_TYPE=disk" "ATHENS_DISK_STORAGE_ROOT=/var/lib/athens" \
--ip-address=Public \
--dns-name="${AZURE_ATHENS_DNS_NAME}" \
--ports="3000" \
--location=${LOCATION}
```

Once you've created the ACI container, you'll see a JSON blob that includes the public IP address of the container. You'll also see the [fully qualified domain name](https://en.wikipedia.org/wiki/Fully_qualified_domain_name) (FQDN) of the running container (it will be prefixed by `AZURE_ATHENS_DNS_NAME`).

### Installing with the MongoDB Storage Driver

First, make sure you have the following environment variable set up:

- `AZURE_ATHENS_MONGO_URL` - The MongoDB connection string. For example: `mongodb://username:password@mongo.server.com/?ssl=true`

Then run the create command:

```console
az container create \
-g "${AZURE_ATHENS_RESOURCE_GROUP}" \
-n "${AZURE_ATHENS_CONTAINER_NAME}-${LOCATION}" \
--image gomods/athens:v0.2.0 \
-e "ATHENS_STORAGE_TYPE=mongo" "ATHENS_MONGO_STORAGE_URL=${AZURE_ATHENS_MONGO_URL}" \
--ip-address=Public \
--dns-name="${AZURE_ATHENS_DNS_NAME}" \
--ports="3000" \
--location=${LOCATION}
```

Once you've created the ACI container, you'll see a JSON blob that includes the public IP address of the container. You'll also see the [fully qualified domain name](https://en.wikipedia.org/wiki/Fully_qualified_domain_name) (FQDN) of the running container (it will be prefixed by `AZURE_ATHENS_DNS_NAME`).

