---
title: "Intro"
date: 2018-12-07T11:44:36+00:00
---

<img src="/banner.png" width="600" alt="Athens Logo"/>

## Athens is a Server for Your Go Packages

Welcome, Gophers! We're looking forward to introducing you to Athens..

On this site, we document Athens in detail. We'll teach you what it does, why it matters, what you can do with it, and how you can run it yourself. Below is a brief summary for you.

#### How To Get Started?
    Run `docker run -p '3000:3000' gomods/athens:latest`

Then, set up your `GOPROXY` and `go get` going!

    export GOPROXY=http://localhost:3000 && go get module@v1

#### What Does Athens Do?

Athens provides a server for [Go Modules](https://github.com/golang/go/wiki/Modules) that you can run. It serves public code and your private code for you, so you don't have to pull directly from a version control system (VCS) like GitHub or GitLab.

#### Why does it matter? 

There are many reasons why you'd want a proxy server such as security and performance. [Take a look](/intro/why) at a few of them

#### How Do I Use It?

Athens is easy to run yourself. We give you a few options:

- You can run it as a binary on your system
    - Instructions coming soon for this
- You can run it as a [Docker](https://www.docker.com/) image (see [here](./install/shared-team-instance/) for how to do that)
- You can run it on [Kubernetes](https://kubernetes.io) (see [here](./install/install-on-kubernetes/) for how to do that)

We also run an experimental version of Athens so you can get started without even installing anything. To get started, set `GOPROXY="https://athens.azurefd.net"`.

>This is not a production-ready proxy deployment, though. Please deploy your own Athens instance for your builds. _If you need a hosted proxy for public code, consider using either `https://gocenter.io` or `https://proxy.golang.org`_.

**[Like what you hear? Try Athens Now!](/try-out)**


## Not Ready to try Athens Yet?

Here are some other ways to get involved:

- Read the full [walkthrough](/walkthrough) with setting up, running and testing the Athens proxy
explores this in greater depth.
* Join our [office hours](/contributing/community/office-hours/)! It's a great way to meet folks working on the project, ask questions or just hang out. All are welcome to join and participate.
* Check out our issue queue for [good first issues](https://github.com/gomods/athens/issues?q=is%3Aopen+is%3Aissue+label%3A%22good+first+issue%22)
* Join us in the `#athens` channel on the [Gophers Slack](https://invite.slack.golangbridge.org/)

---
Athens banner attributed to Golda Manuel
