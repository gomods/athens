---
title: Try it out!
description: How to see Athens in action 
---

## Try out Athens

To quickly see Athens in action, follow these steps:

First, make sure you have [Go v1.12+ installed](https://gophersource.com/setup/),
that GOPATH/bin is on your path, and that you have enabled the [Go
Modules](https://github.com/golang/go/wiki/Modules) feature.

**Bash**
```bash
export GO111MODULE=on
```

**PowerShell**
```powershell
$env:GO111MODULE = "on"
```

Next, use git and Go to install and run the Athens proxy in a background process.

```console
$ git clone https://github.com/gomods/athens
$ cd athens/cmd/proxy
$ go install
$ proxy &
[1] 37186
INFO[0000] Exporter not specified. Traces won't be exported
INFO[0000] Starting application at http://127.0.0.1:3000
```

Next, you will need to configure Go to use the Athens proxy!

**Bash**
```bash
export GOPROXY=http://127.0.0.1:3000
```

**PowerShell**
```powershell
$env:GOPROXY = "http://127.0.0.1:3000"
```

Now, when you build and run this example application, **go** will fetch dependencies via Athens!

```console
$ git clone https://github.com/athens-artifacts/walkthrough.git
$ cd walkthrough
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

This should give you an overview of what using Athens is like!

