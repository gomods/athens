---
title: "Intro"
date: 2018-02-11T16:52:23-05:00
---

<img src="/banner.png" width="600" alt="Athens Logo"/>


## Athens is a server for your Go packages

Welcome, Gophers! We're looking forward to introducing you to Athens.. 

On this site, we document Athens in detail. We'll teach you what it does, why it matters, what you can do with it, and how you can run it yourself. Below is a brief summary for you.

#### What does Athens do?

Athens provides a server for [Go Modules](https://github.com/golang/go/wiki/Modules) that you can run. It serves public code and your private code for you, so you don't have to pull directly from a version control system (VCS) like GitHub or GitLab

#### Why does it matter?

Previously, the Go community has had lots of problems with libraries disappearing or changing without warning. If someone changes their library, your app can break overnight - oh no! Athens solves these problems by copying code from VCS's into _immutable_ storage.

#### How do I use it?

Athens is easy to run yourself. We give you a few options:

- You can run it as a binary on your system
    - Instructions coming soon for this
- You can run it as a [Docker](https://www.docker.com/) image (see [here](http://localhost:1313/install/shared-team-instance/) for how to do that)
- You can run it on [Kubernetes](https://kubernetes.io) (see [here](http://localhost:1313/install/install-on-kubernetes/) for how to do that)

We also run an experimental server for public use, so you can get started with Athens without even installing it. For details, see [here](./public_proxy).

**Sound good to you? Read on to learn more!**

## Try out Athens

To quickly see Athens in action, follow these steps:

First, make sure you have [Go 1.11 installed](https://gophersource.com/setup/) and that GOPATH/bin is on your path.

Next, use Go to install and run the Athens proxy in a background process

```console
$ go get -u github.com/gomods/athens/cmd/proxy
# the source is downloaded to GOPATH/src/github.com/gomods/athens/
$ proxy &
[1] 25243
INFO[0000] Starting application at 127.0.0.1:3000
```

Next, you will need to enable the [Go Modules](https://github.com/golang/go/wiki/Modules)
feature and configure Go to use the proxy!

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

##  Next Steps
* If you are interested
in what is happening between the proxy and the Go Modules feature, the [Walkthrough](/walkthrough)
explores this in greater depth.
* For more detailed information on how to run Athens in production, check out our other documentation (coming soon!!).

---
Athens banner attributed to Golda Manuel
