---
title: Install Athens with BOSH
description: Installing an Athens Instance with BOSH
weight: 8
---

Athens can be deployed in many ways. The following guide explains how to use [BOSH](https://bosh.io), a deployment and lifecycle tool, in order to deploy an Athens server on virtual machines (VMs) on any infrastructure as a service (IaaS) that is supported by BOSH.

---

## Prerequisites

* Install [BOSH](#install-bosh)
* Setup the [infrastructure](#setup-the-infrastructure)

### Install BOSH

Make sure to have the [BOSH CLI](https://bosh.io/docs/cli-v2-install/) installed and set up a [BOSH Director](https://bosh.io/docs/quick-start/) on an infrastructure of your choice.

### Setup the Infrastructure

If you choose to deploy on a IaaS provider, there are a few prerequisites that need to be set up before starting with the deployment. Depending on which IaaS you will be deploying, you may need to create:

* **Public IP**: a public IP address for association with the Athens VM.
* **Firewall Rules**: the following ingress ports must be allowed

    - `3000/tcp` - Athens proxy port (if you specify a different port than the default port 3000 in the job properties, adapt this rule accordingly).

    Egress traffic should be restricted depending on your requirements.

#### Amazon Web Services (AWS)

AWS requires additional settings that should be added to a `credentials.yml` file using the following template:

```yaml
# security group IDs to apply to the VM
athens_security_groups: [sg-0123456abcdefgh]

# VPC subnet to deploy Athens to
athens_subnet_id: subnet-0123456789abcdefgh

# a specific, elastic IP address for the VM
external_ip: 3.123.200.100
```

The credentials need to be added to the `deploy` command, i.e.

```
-o manifests/operations/aws-ops.yml
```

#### VirtualBox

The fastest way to install Athens using BOSH is probably a Director VM running on VirtualBox which is sufficient for development or testing purposes. If you follow the bosh-lite [installation guide](https://bosh.cloudfoundry.org/docs/bosh-lite), no further preparation is required to deploy Athens.


## Deployment

A deployment manifest contains all the information for managing and updating a BOSH deployment. To aid in the deployment of Athens on BOSH, the [athens-bosh-release](https://github.com/s4heid/athens-bosh-release) repository provides manifests for basic deployment configurations inside the `manifests` directory. For quickly creating a standalone Athens server, clone the release repository and `cd` into it:

```console
git clone --recursive https://github.com/s4heid/athens-bosh-release.git
cd athens-bosh-release
```

Once the [infrastructure](#setup-the-infrastructure) has been prepared and the BOSH Director is running, make sure that a [stemcell](https://bosh.cloudfoundry.org/docs/stemcell/) has been uploaded. If this has not been done yet, choose a stemcell from the [stemcells section of bosh.io](https://bosh.io/stemcells), and upload it via the command line. Additionally, a [cloud config](https://bosh.cloudfoundry.org/docs/cloud-config/) is required for IaaS specific configuration used by the Director and the Athens deployment. The `manifests` directory also contains an example cloud config, which can be uploaded to the Director via

```console
bosh update-config --type=cloud --name=athens \
    --vars-file=credentials.yml manifests/cloud-config.yml
```

Execute the `deploy` command which can be extended with ops/vars files depending on which IaaS you will be deploying to.

```console
bosh -d athens deploy manifests/athens.yml  # add extra arguments
```

For example, when using AWS the deploy command for an Athens Proxy with disk storage would look like

```console
bosh -d athens deploy \
    -o manifests/operations/aws-ops.yml \
    -o manifests/operations/with-persistent-disk.yml \
    -v disk_size=1024 \
    --vars-file=credentials.yml manifests/athens.yml
```

This will deploy a single Athens instance in the `athens` deployment with a persistent disk of 1024MB. The IP address of that instance can be obtained with

```console
bosh -d athens instances
```

which is useful for targeting Athens, e.g. with the `GOPROXY` variable. You can follow this [quickstart guide](/try-out) for more information.