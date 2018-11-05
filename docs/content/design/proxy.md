---
title: "Proxy"
date: 2018-02-11T15:59:56-05:00
---

## The Athens Proxy

The Athens proxy has two primary use cases:

- Internal deployments
- Public mirror deployments

This document details features of the Athens proxy that you can use to achieve either use case.

## The Role of the Athens proxy

We intend proxies to be deployed primarily inside of enterprises to:

- Host private modules
- Exclude access to public modules
- Store public modules

Importantly, a proxy is not intended to be a complete _mirror_ of an upstream proxy. For public modules, its role is to store the modules locally and provide access control.

## What happens when a public module is not stored?

When a user requests a module `MxV1` from a proxy and the Athens proxy doesn't have `MxV1` in its store, it first determines whether `MxV1` is private or not private.

If it's private, it immediately stores the module into the proxy storage from the internal VCS.

If it's not private, the Athens proxy consults its exclude list for non-private modules (see below). If `MxV1` is on the exclude list, the Athens proxy returns 404 and does nothing else. If `MxV1` is not on the exclude list, the Athens proxy executes the following algorithm:

```
upstreamDetails := lookUpstream(MxV1)
if upstreamDetails == nil {
	return 404 // if the upstream doesn't have the thing, just bail out
}
return upstreamDetails.baseURL
```

The important part of this algorithm is `lookUpstream`. That function queries an endpoint on the upstream proxy that either:

- Returns 404 if it doesn't have `MxV1` in its storage
- Returns the base URL for MxV1 if it has `MxV1` in its storage

_In a later version of the project, we may implement an event stream on proxies that any other proxy can subscribe to and listen for deletions/deprecations on modules that it cares about_

## Exclude Lists and Private Module Filters

To accommodate private (i.e. enterprise) deployments, the Athens proxy maintains two important access control mechanisms:

- Private module filters
- Exclude lists for public modules

### Private Module Filters

Private module filters are string globs that tell the Athens proxy what is a private module. For example, the string `github.internal.com/**` tells the Athens proxy:

- To never make requests to the public internet (i.e. to upstream proxies) regarding this module
- To download module code (in its store mechanism) from the VCS at `github.internal.com`

### Exclude Lists for Public Modules

Exclude lists for public modules are also globs that tell the Athens proxy what modules it should never download from any upstream proxy. For example, the string `github.com/arschles/**` tells the Athens proxy to always return `404 Not Found` to clients.
