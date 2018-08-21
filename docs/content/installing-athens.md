---
title: Installing Athens
description: Installing an Athens Instance For Your Development Team
menu: shortcuts
---

When you follow the instructions in the [Walkthrough](/walkthrough), you end up with an Athens Proxy that uses in-memory storage. This is only suitable for trying out the proxy for a short period of time, as you will quickly run out of memory and Athens won't persist modules between restarts. This guide will help you get Athens running in a more suitable manner for scenarios like providing an instance for your development team to share.

Note: [Currently, the proxy does not work on Windows](https://github.com/gomods/athens/issues/532).

First, make sure you have [Go 1.11 installed](https://gophersource.com/setup/) and that GOPATH/bin is on your path. We will use Go to build and run the proxy.

## Selecting a Storage Provider

Athens currently supports a number of storage drivers. For local use we recommend starting with the local disk provider. For other providers, please see [Storage Providers]() [Coming Soon].

## Running Athens with Local Disk Storage

First, let's build the proxy and install it locally.

**Bash**
```bash
go get -u github.com/gomods/athens/cmd/proxy
mv $GOPATH/bin/proxy /usr/local/bin/athens
```

In order to run Athens with disk storage, you will next need to identify where you would like to persist modules. In the example below, we will create a new directory located at **/var/lib/athens**. Next, create an environment variable named **ATHENS_DISK_STORAGE_ROOT** and set it to this new directory location. This directory needs to be writeable by the user that will run Athens. Additionally, you will need to set the **ATHENS_STORAGE_TYPE** environment variable to the value **disk**. Now you are ready to run Athens with disk storage enabled.

**Bash**
```console
$ sudo mkdir -p /var/lib/athens
$ sudo chown -R $(whoami) /var/lib/athens
$ export ATHENS_DISK_STORAGE_ROOT=/var/lib/athens
$ export ATHENS_STORAGE_TYPE=disk
$ /usr/local/bin/athens
INFO[0000] Starting application at 127.0.0.1:3000
```

Now, you can use the proxy with the Go command line tool and modules will be stored locally on disk. To verify this, open a new terminal window and try the following example:

**Bash**
```console
$ git clone https://github.com/athens-artifacts/walkthrough.git
$ cd walkthrough
$ GO111MODULE=on go run .
go: downloading github.com/athens-artifacts/samplelib v1.0.0
The 🦁 says rawr!
```

In the terminal where you ran Athens, you should see output like:

```console
handler: GET /github.com/athens-artifacts/samplelib/@v/v1.0.0.info [200]
handler: GET /github.com/athens-artifacts/samplelib/@v/v1.0.0.mod [200]
handler: GET /github.com/athens-artifacts/samplelib/@v/v1.0.0.zip [200]
```

Now, if you view the contents of the **ATHENS_DISK_STORAGE_ROOT** directory, you will see that you now have additional files representing the samplelib module.

```console
$ ls -lr $ATHENS_DISK_STORAGE_ROOT/github.com/athens-artifacts/samplelib/v1.0.0/
total 24
-rwxr-xr-x  1 jeremyrickard  wheel    50 Aug 21 10:52 v1.0.0.info
-rwxr-xr-x  1 jeremyrickard  wheel  2391 Aug 21 10:52 source.zip
-rwxr-xr-x  1 jeremyrickard  wheel    45 Aug 21 10:52 go.mod
```

Now if Athens is restarted, it will serve the module from this location.