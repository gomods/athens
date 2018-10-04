---
title: "Installing Athens"
date: 2018-09-20T15:38:01-07:00
weight: 2
---

The Go ecosystem has always been federated and completely open. Anyone with a GitHub or GitLab (or any other supported VCS) account can effortlessly provide a library with just a `git push` (or similar). No extra accounts to create or credentials to set up.

## Federation

We feel that Athens should keep the community federated and open, and nobody should have to change their workflow when they're building apps and libraries. So, to make sure the community can stay federated and open, we've made it easy to install Athens for everyone so that:

- Anyone can run their own full-featured mirror, public or private
- Any organization can run their own private mirror, so they can manage their private code just as they would their public code

## Where to Go from Here

To make sure it's easy to install, we try to provide as many ways as possible to install and run Athens:

- It's written in Go, so we provide a self-contained binary. You can configure and run the binary on your machine(s) 
    - Instructions on how to run directly from the binary are coming soon
- We provide a [Docker image](https://hub.docker.com/r/gomods/proxy/) and [instructions on how to run it](./shared-team-instance)
- We provide [Kubernetes](https://kubernetes.io) [Helm Charts](https://helm.sh) with [instructions on how to run Athens on Kubernetes](http://localhost:1313/install/install-on-kubernetes/)
