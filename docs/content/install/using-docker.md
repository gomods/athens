---
title: Using the Athens Docker images
description: Information about Athens' Docker images
weight: 1
---

Whether setting Athens up using [Kubernetes](/install/install-on-kubernetes/) or using the [Walkthrough](/Walkthrough), you'll most likely be using one of the images that the Athens project produces. This document details what images are available, and has a recap from the Walkthrough of how to use them on their own.

---

## Available Docker images

The Athens project produces two docker images, available via [Docker Hub](https://hub.docker.com/) 

1. A release version as [`gomods/athens`](https://hub.docker.com/r/gomods/athens), each tag corresponds with an Athens [release](https://github.com/gomods/athens/releases), e.g. `v0.7.1`. Additionally, a `canary` tag is available and tracks each commit to `main`
2. A tip version, as [`gomods/athens-dev`](https://hub.docker.com/r/gomods/athens-dev), tagged with every commit to `main`, e.g. `1573339`

For a detailed tags list, check each image's Docker Hub

## Running Athens as a Docker image

This is a quick recap of the [Walkthrough](/walkthrough)

### Using the `docker` cli

In order to run the Athens Proxy using docker, we need first to create a directory that will store the persitant modules.
In the example below, the new directory is named `athens-storage` and is located in our userspace (i.e. `$HOME`). 
Then we need to set the `ATHENS_STORAGE_TYPE` and `ATHENS_DISK_STORAGE_ROOT` environment variables when we run the Docker container.

**Bash**
```bash
export ATHENS_STORAGE=$HOME/athens-storage
mkdir -p $ATHENS_STORAGE
docker run -d -v $ATHENS_STORAGE:/var/lib/athens \
   -e ATHENS_DISK_STORAGE_ROOT=/var/lib/athens \
   -e ATHENS_STORAGE_TYPE=disk \
   --name athens-proxy \
   --restart always \
   -p 3000:3000 \
   gomods/athens:latest
```

**PowerShell**
```PowerShell
$env:ATHENS_STORAGE = "$(Join-Path $HOME athens-storage)"
md -Path $env:ATHENS_STORAGE
docker run -d -v "$($env:ATHENS_STORAGE):/var/lib/athens" `
   -e ATHENS_DISK_STORAGE_ROOT=/var/lib/athens `
   -e ATHENS_STORAGE_TYPE=disk `
   --name athens-proxy `
   --restart always `
   -p 3000:3000 `
   gomods/athens:latest
```

## Non-Root User

The Athens docker images comes with a non-root user `athens` with `uid: 1000`, `gid: 1000` and home directory `/home/athens`.
In situations where running as root is not permitted, this user can be used instead. In all other instuctions
replace `/root/` with `/home/athens/` and set the user and group ids in the run environment to `1000`.

```shell
docker run -d -v $ATHENS_STORAGE:/var/lib/athens \
   -e ATHENS_DISK_STORAGE_ROOT=/var/lib/athens \
   -e ATHENS_STORAGE_TYPE=disk \
   -v "$PWD/gitconfig/.gitconfig:/home/athens/.gitconfig" \
   --name athens-proxy \
   --restart always \
   -p 3000:3000 \
   -u 1000:1000 \
   gomods/athens:latest
```

## Troubleshooting Athens in Docker

### `init` issues

The Athens docker image uses [tini](https://github.com/krallin/tini) so that defunct processes get reaped.
Docker 1.13 and greater includes `tini` and lets you enable it by passing the `--init` flag to `docker run` or by configuring the docker deamon with `"init": true`. When running in this mode. you may see a warning like this:

```console
[WARN  tini (6)] Tini is not running as PID 1 and isn't registered as a child subreaper.
 Zombie processes will not be re-parented to Tini, so zombie reaping won't work.
 To fix the problem, use the -s option or set the environment variable TINI_SUBREAPER to register Tini as a child subreaper, or run Tini as PID 1.
```
This is the "Athens-tini" complaining that it's not running as PID 1.
There is no harm in that, since the zombie processes will be reaped by the `tini` included in Docker.
