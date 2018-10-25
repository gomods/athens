---
title: "Components"
date: 2018-02-11T16:57:56-05:00
---

From a very high-level view, we recognize 4 major components of the system.

### Client

The client is a user, powered by go binary with module support. At the moment of writing this document, it is `go1.11`

### VCS

VCS is an external source of data for Athens. Athens scans various VCSs such as `github.com` and fetches sources from there.

### Proxy

We intend proxies to be deployed primarily inside of enterprises to:

* Host private modules
* Exclude access to public modules
* Cache public modules

Importantly, a proxy is not intended to be a complete mirror of an upstream proxy. For public modules, its role is to cache and provide access control.
