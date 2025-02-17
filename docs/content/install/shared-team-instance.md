---
title: Shared Team Instance
description: Installing an Athens Instance For Your Development Team
weight: 2
---

When you follow the instructions in the [Walkthrough](/walkthrough), you end up with an Athens Proxy that uses in-memory storage. This is only suitable for trying out the Athens proxy for a short period of time, as you will quickly run out of memory and Athens won't persist modules between restarts. This guide will help you get Athens running in a more suitable manner for scenarios like providing an instance for your development team to share.

We will use Docker to run the Athens proxy, so first make sure you have Docker [installed](https://docs.docker.com/install/).

## Selecting a Storage Provider

Athens currently supports a number of storage drivers. For local use we recommend starting with the local disk provider. For other providers, please see [the Storage Provider documentation](/configuration/storage).


## Running Athens with Local Disk Storage

In order to run Athens with disk storage, you will next need to identify where you would like to persist modules. In the example below, we will create a new directory named `athens-storage` in our current directory.  Now you are ready to run Athens with disk storage enabled. To enable disk storage, you need to set the `ATHENS_STORAGE_TYPE` and `ATHENS_DISK_STORAGE_ROOT` environment variables when you run the Docker container.

The examples below use the `:latest` Docker tags for simplicity, however we strongly recommend that after your environment is up and running that you switch to using
an explicit version (for example `:v0.3.0`).

**Bash**
```bash
export ATHENS_STORAGE=~/athens-storage
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
$env:ATHENS_STORAGE = "$(Join-Path $pwd athens-storage)"
md -Path $env:ATHENS_STORAGE
docker run -d -v "$($env:ATHENS_STORAGE):/var/lib/athens" `
   -e ATHENS_DISK_STORAGE_ROOT=/var/lib/athens `
   -e ATHENS_STORAGE_TYPE=disk `
   --name athens-proxy `
   --restart always `
   -p 3000:3000 `
   gomods/athens:latest
```

Note: if you have not previously mounted this drive with Docker for Windows, you may be prompted to allow access

Athens should now be running as a Docker container with the local directory, `athens-storage` mounted as a volume. When Athens retrieves the modules, they will be stored in the directory previously created. First, let's verify that Athens is running:

```console
$ docker ps
CONTAINER ID        IMAGE                               COMMAND           PORTS                    NAMES
f0429b81a4f9        gomods/athens:latest   "/bin/app"        0.0.0.0:3000->3000/tcp   athens-proxy
```

Now, we can use Athens from any development machine that has Go v1.12+ installed. To verify this, try the following example:

**Bash**
```console
$ export GO111MODULE=on
$ export GOPROXY=http://127.0.0.1:3000
$ git clone https://github.com/athens-artifacts/walkthrough.git
$ cd walkthrough
$ go run .
go: downloading github.com/athens-artifacts/samplelib v1.0.0
The 🦁 says rawr!
```


**PowerShell**
```console
$env:GO111MODULE = "on"
$env:GOPROXY = "http://127.0.0.1:3000"
git clone https://github.com/athens-artifacts/walkthrough.git
cd walkthrough
$ go run .
go: downloading github.com/athens-artifacts/samplelib v1.0.0
The 🦁 says rawr!
```

We can verify that Athens handled this request by examining the Docker logs:

```console
$ docker logs -f athens-proxy
time="2018-08-21T17:28:53Z" level=warning msg="Unless you set SESSION_SECRET env variable, your session storage is not protected!"
time="2018-08-21T17:28:53Z" level=info msg="Starting application at 0.0.0.0:3000"
handler: GET /github.com/athens-artifacts/samplelib/@v/v1.0.0.info [200]
handler: GET /github.com/athens-artifacts/samplelib/@v/v1.0.0.mod [200]
handler: GET /github.com/athens-artifacts/samplelib/@v/v1.0.0.zip [200]
```

Now, if you view the contents of the `athens_storage` directory, you will see that you now have additional files representing the samplelib module.

**Bash**
```console
$ ls -lr $ATHENS_STORAGE/github.com/athens-artifacts/samplelib/v1.0.0/
total 24
-rwxr-xr-x  1 jeremyrickard  wheel    50 Aug 21 10:52 v1.0.0.info
-rwxr-xr-x  1 jeremyrickard  wheel  2391 Aug 21 10:52 source.zip
-rwxr-xr-x  1 jeremyrickard  wheel    45 Aug 21 10:52 go.mod
```

**PowerShell**
```console
$ dir $env:ATHENS_STORAGE\github.com\athens-artifacts\samplelib\v1.0.0\


    Directory: C:\athens-storage\github.com\athens-artifacts\samplelib\v1.0.0


Mode                LastWriteTime         Length Name
----                -------------         ------ ----
-a----        8/21/2018   3:31 PM             45 go.mod
-a----        8/21/2018   3:31 PM           2391 source.zip
-a----        8/21/2018   3:31 PM             50 v1.0.0.info
```

When Athens is restarted, it will serve the module from this location without re-downloading it. To verify that, we need to first remove the Athens container.

```console
docker rm -f athens-proxy
```

Now, we need to clear the local Go modules storage. This is needed so that your local Go command line tool will re-download the module from Athens. The following commands will clear the local module storage:

**Bash**
```bash
sudo rm -fr "$(go env GOPATH)/pkg/mod"
```

**PowerShell**
```powershell
rm -recurse -force $(go env GOPATH)\pkg\mod
```

Now, we can re-run the Athens container:

**Bash**
```console
docker run -d -v $ATHENS_STORAGE:/var/lib/athens \
   -e ATHENS_DISK_STORAGE_ROOT=/var/lib/athens \
   -e ATHENS_STORAGE_TYPE=disk \
   --name athens-proxy \
   --restart always \
   -p 3000:3000 \
   gomods/athens:latest
```

**PowerShell**
```console
docker run -d -v "$($env:ATHENS_STORAGE):/var/lib/athens" `
   -e ATHENS_DISK_STORAGE_ROOT=/var/lib/athens `
   -e ATHENS_STORAGE_TYPE=disk `
   --name athens-proxy `
   --restart always `
   -p 3000:3000 `
   gomods/athens:latest
```

When we re-run our Go example, the Go cli will again download module from Athens. Athens, however, will not need to retrieve the module. It will be served from the Athens on-disk storage.

**Bash**
```console
$ ls -lr $ATHENS_STORAGE/github.com/athens-artifacts/samplelib/v1.0.0/
total 24
-rwxr-xr-x  1 jeremyrickard  wheel    50 Aug 21 10:52 v1.0.0.info
-rwxr-xr-x  1 jeremyrickard  wheel  2391 Aug 21 10:52 source.zip
-rwxr-xr-x  1 jeremyrickard  wheel    45 Aug 21 10:52 go.mod
```

**PowerShell**
```console
$ dir $env:ATHENS_STORAGE\github.com\athens-artifacts\samplelib\v1.0.0\


    Directory: C:\athens-storage\github.com\athens-artifacts\samplelib\v1.0.0


Mode                LastWriteTime         Length Name
----                -------------         ------ ----
-a----        8/21/2018   3:31 PM             45 go.mod
-a----        8/21/2018   3:31 PM           2391 source.zip
-a----        8/21/2018   3:31 PM             50 v1.0.0.info
```

Notice that the timestamps given have not changed.

Next Steps:

* [Run the Athens Proxy on Kubernetes with Helm](/install/install-on-kubernetes)
* Explore best practices for running Athens in Production. [Coming Soon](https://github.com/gomods/athens/issues/531)
