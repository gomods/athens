---
title: Walkthrough
description: Understanding the Athens proxy and Go Modules
---

First, make sure you have [Go v1.12+ installed](https://gophersource.com/setup/) and that GOPATH/bin is on your path.

## Without the Athens proxy
Let's review what everything looks like in Go without the Athens proxy in the picture:

**Bash**
```console
$ git clone https://github.com/athens-artifacts/walkthrough.git
$ cd walkthrough
$ GO111MODULE=on go run .
go: downloading github.com/athens-artifacts/samplelib v1.0.0
The 🦁 says rawr!
```

**PowerShell**
```console
$ git clone https://github.com/athens-artifacts/walkthrough.git
$ cd walkthrough
$ $env:GO111MODULE = "on"
$ go run .
go: downloading github.com/athens-artifacts/samplelib v1.0.0
The 🦁 says rawr!
```

The end result of running this command is that Go downloaded the package source and packaged
it into a module, saving it in the Go Modules local storage.

Now that we have seen Go Modules in action without the Athens proxy, let's take a look at
how the Athens proxy changes the workflow and the output.

## With the Athens proxy
Using the most simple installation possible, let's walk through how to use the
Athens proxy, and figure out what is happening at each step.

Before moving on, let's clear our Go Modules local files so that we can see the Athens proxy
in action without any modules locally populated:

**Bash**
```bash
sudo rm -fr $(go env GOPATH)/pkg/mod
```

**PowerShell**
```powershell
rm -recurse -force "$(go env GOPATH)\pkg\mod"
```

Now run the Athens proxy in a background process:

**Bash**
```console
$ mkdir -p $(go env GOPATH)/src/github.com/gomods
$ cd $(go env GOPATH)/src/github.com/gomods
$ git clone https://github.com/gomods/athens.git
$ cd athens
$ GO111MODULE=on go run ./cmd/proxy -config_file=./config.dev.toml &
[1] 25243
INFO[0000] Starting application at 127.0.0.1:3000
```

**PowerShell**
```console
$ mkdir "$(go env GOPATH)\src\github.com\gomods"
$ cd "$(go env GOPATH)\src\github.com\gomods"
$ git clone https://github.com/gomods/athens.git
$ cd athens
$ $env:GO111MODULE = "on"
$ $env:GOPROXY = "https://proxy.golang.org"
$ Start-Process -NoNewWindow go 'run .\cmd\proxy -config_file=".\config.dev.toml"'
[1] 25243
INFO[0000] Starting application at 127.0.0.1:3000
```

The Athens proxy is now running in the background and is listening for requests
from localhost (127.0.0.1) on port 3000.

Since we didn't provide any specific configuration
the Athens proxy is using in-memory storage, which is only suitable for trying out the Athens proxy
for a short period of time, as you will quickly run out of memory and the storage
doesn't persist between restarts.

### With Docker

For more details on running Athens in docker, take a look at the [install documentation](/install/using-docker)

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

Next, you will need to enable the [Go Modules](https://github.com/golang/go/wiki/Modules)
feature and configure Go to use the Athens proxy!

### Using the Athens proxy
**Bash**
```bash
export GO111MODULE=on
export GOPROXY=http://127.0.0.1:3000
```

**PowerShell**
```powershell
$env:GO111MODULE = "on"
$env:GOPROXY = "http://127.0.0.1:3000"
```

The `GO111MODULE` environment variable controls the Go Modules feature in Go 1.11 only.
Possible values are:

* `on`: Always use Go Modules
* `auto` (default): Only use Go Modules when a go.mod file is present, or the go command is run from _outside_ the GOPATH
* `off`: Never use Go Modules

The `GOPROXY` environment variable tells the `go` binary that instead of talking to
the version control system, such as github.com, directly when resolving your package
dependencies, instead it should communicate with a proxy. The Athens proxy implements
the [Go Download Protocol](/intro/protocol), and is responsible for listing available
versions for a package in addition to providing a zip of particular package versions.

Now, you when you build and run this example application, `go` will fetch dependencies via Athens!

```console
$ cd ../walkthrough
$ go run .
go: finding github.com/athens-artifacts/samplelib v1.0.0
handler: GET /github.com/athens-artifacts/samplelib/@v/v1.0.0.info [200]
handler: GET /github.com/athens-artifacts/samplelib/@v/v1.0.0.mod [200]
go: downloading github.com/athens-artifacts/samplelib v1.0.0
handler: GET /github.com/athens-artifacts/samplelib/@v/v1.0.0.zip [200]
The 🦁 says rawr!
```

The output from `go run .` includes attempts to find the **github.com/athens-artifacts/samplelib** dependency. Since the
proxy was run in the background, you should also see output from Athens indicating that it is handling requests for the dependency.

Let's break down what is happening here:

1. Before Go runs our code, it detects that our code depends on the **github.com/athens-artifacts/samplelib** package
   which is not present in the Go Modules local storage.
2. At this point the Go Modules feature comes into play because we have it enabled.
    Instead of looking in the GOPATH for the package, Go reads our **go.mod** file
    and sees that we want a particular version of that package, v1.0.0.

    ```go
    module github.com/athens-artifacts/walkthrough
    
    require github.com/athens-artifacts/samplelib v1.0.0
    ```
3. Go first checks for **github.com/athens-artifacts/samplelib@v1.0.0** in the Go Modules local storage,
    located in GOPATH/pkg/mod. If that version of the package is already local storage,
    then Go will use it and stop looking. But since this is our first time
    running this, our local storage is empty and Go keeps looking.
4. Go requests **github.com/athens-artifacts/samplelib@v1.0.0** from our proxy because
    it is set in the GOPROXY environment variable.
5. The Athens proxy checks its own storage (in this case is in-memory) for the package and doesn't find it. So it
    retrieves it from github.com and then saves it for subsequent requests.
6. Go downloads the module zip and puts it in the Go Modules local storage
    GOPATH/pkg/mod.
7. Go will use the module and build our application!

Subsequent calls to `go run .` will be much less verbose:

```
$ go run .
The 🦁 says rawr!
```

No additional output is printed because Go found **github.com/athens-artifacts/samplelib@v1.0.0** in the Go Module
local storage and did not need to request it from the Athens proxy.

Lastly, quitting from the Athens proxy. This cannot be done directly because we are starting the Athens proxy in the background, thus we must kill it by finding it's process ID and killing it manually.

**Bash**
```bash
lsof -i @localhost:3000
kill -9 <<PID>>
```

**PowerShell**
```powershell
netstat -ano | findstr :3000 (local host Port number)
taskkill /PID typeyourPIDhere /F
```

## Next Steps

Now that you have seen Athens in Action:

* Learn how to [install a shared team Athens](/install/shared-team-instance) with persistent storage.
* Explore best practices for running Athens in Production. [Coming Soon/Help Wanted](https://github.com/gomods/athens/issues/531)
