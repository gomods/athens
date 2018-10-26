---
title: "Intro"
date: 2018-02-11T16:52:23-05:00
---

<img src="/banner.png" width="600" alt="Athens Logo"/>

## Athens is a Server for Your Go Packages

Welcome, Gophers! We're looking forward to introducing you to Athens..

On this site, we document Athens in detail. We'll teach you what it does, why it matters, what you can do with it, and how you can run it yourself. Below is a brief summary for you.

#### What Does Athens Do?

Athens provides a server for [Go Modules](https://github.com/golang/go/wiki/Modules) that you can run. It serves public code and your private code for you, so you don't have to pull directly from a version control system (VCS) like GitHub or GitLab.

#### Why Does it Matter?

Previously, the Go community has had lots of problems with libraries disappearing or changing without warning. It's easy for package maintainers to make changes to their code that can break yours - and much of the time it's an accident! Could your build break if one of your dependencies did this?

- Commit `abdef` was deleted
- Tag `v0.1.0` was force pushed
- The repository was deleted altogether

 Since your app's dependencies come directly from GitHub, any of those above cases can happen to you and your builds can break when they do - oh no! Athens solves these problems by copying code from VCS's into _immutable_ storage.

#### How Do I Use It?

Athens is easy to run yourself. We give you a few options:

- You can run it as a binary on your system
    - Instructions coming soon for this
- You can run it as a [Docker](https://www.docker.com/) image (see [here](./install/shared-team-instance/) for how to do that)
- You can run it on [Kubernetes](https://kubernetes.io) (see [here](./install/install-on-kubernetes/) for how to do that)

We also run an experimental server for public use, so you can get started with Athens without even installing it. For details, see [here](./public_proxy).

**[Like what you hear? Try Athens Now!](/try-out)**


## Not Ready to try Athens Yet?

Here are some other ways to get involved:

- Read the full [walkthrough](/walkthrough) with setting up, running and testing the Athens proxy
explores this in greater depth.
* Join our [weekly development meeting](/contributing/community/developer-meetings/)! It's a great way to meet folks working on the project, ask questions or just hang out. All are welcome to join and participate.
* Check out our issue queue for [good first issues](https://github.com/gomods/athens/issues?q=is%3Aopen+is%3Aissue+label%3A%22good+first+issue%22)
* Join us in the `#athens` channel on the [Gophers Slack](https://invite.slack.golangbridge.org/)

---
Athens banner attributed to Golda Manuel
