---
title: "Intro"
date: 2018-12-07T11:44:36+00:00
---

<img src="/banner.png" width="600" alt="Athens Logo"/>

## Athens is a Server for Your Go Packages

Welcome, Gophers! Athens is an open source enterprise ready [Go Module proxy](https://go.dev/ref/mod#module-proxy) with extensive configuration for a variety of online and offline usecases. 
It is in use for anti-censorship, compliance, data privacy, and data continuity usecases in homes and corporations across the globe today.

#### How To Get Started?
Run `docker run -p '3000:3000' gomods/athens:latest`

Then, set up your `GOPROXY` and `go get` going!

    export GOPROXY=http://localhost:3000 && go get module@v1

When you're ready to run something more production ready, Athens can run on on a variety of platforms including [AWS, Azure, GCP, Digital Ocean, Alibaba, and bare metal](./install/).

#### What Does Athens Do?

Athens is an implementation of the [Go Module proxy](https://go.dev/ref/mod#module-proxy). Go clients talk to Athens to retrieve packages at its most basic level. 
Athens supports [many usecases](./intro/why) on top of that basic premise.

## Not Ready to try Athens Yet?

Here are some other ways to get involved:

* Read the full [walkthrough](/walkthrough) with setting up, running and testing the Athens proxy explores this in greater depth.
* Join our [office hours](/contributing/community/office-hours/)! It's a great way to meet folks working on the project, ask questions or just hang out. All are welcome to join and participate.
* Check out our issue queue for [good first issues](https://github.com/gomods/athens/issues?q=is%3Aopen+is%3Aissue+label%3A%22good+first+issue%22)
* Join us in the `#athens` channel on the [Gophers Slack](https://invite.slack.golangbridge.org/)

---
Athens banner attributed to Golda Manuel
