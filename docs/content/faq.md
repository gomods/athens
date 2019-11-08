---
title: FAQ
description: Frequently Asked Questions
menu: shortcuts
---

### Is Athens Just a Proxy? A Registry?

_TL;DR "Registry" doesn't describe what Athens is trying to do here. That implies that there's only one service in the world that can serve Go modules to everyone. Athens isn't trying to be that. Instead, Athens is trying to be part of a federated group of module proxies._

A registry is generally run by one entity, is one logical server that provides authentication (and provenance sometimes), and is pretty much the de-facto only source of dependencies. Sometimes it's run by a for-profit company.

That's most definitely not what we in the Athens community are going for, and that would harm our community if we did go down that path.

First and foremost, Athens is an _implementation_ of the [Go Modules download API](/intro/protocol). Not only does the standard Go toolchain support any implementation of that API, the Athens proxy is designed to talk to any other server that implements that API as well. That allows Athens to talk to other proxies in the community.

Finally, we're purposefully building this project - and working with the toolchain folks - in a way that everyone who wants to write a proxy can participate.

### Does Athens integrate with the go toolchain?

Athens is currently supported by the [Go v1.12+](https://golang.org/dl) toolchain via the [download protocol](/intro/protocol/).

For the TL;DR of the protocol, it's a REST API that lets the go toolchain (i.e. go get) see lists of versions and fetch source code for a specific version.

Athens is a server that implements the protocol. Both it, the protocol and the toolchain (as you almost certainly know) is open source.

### Are the packages served by Athens immutable?

_TL;DR Athens does store code in CDNs and has the option to store code in other persistent datastores._

The longer version:

It's virtually impossible to ensure immutable builds when source code comes from Github. We have been annoyed by that problem for a long time. The Go modules download protocol is a great opportunity to solve this issue. The Athens proxy works pretty simply at a high level:

1. `go get github.com/my/module@v1` happens
1. Athens looks in its datastore, it's missing
1. Athens downloads `github.com/my/module@v1` from Github (it uses go get on the backend too)
1. Athens stores the module in its datastore
1. Athens serves `github.com/my/module@v1` from its datastore forever

To repeat, "datastore" means a CDN (we currently have support for Google Cloud Storage, Azure Blob Storage and AWS S3) or another datastore (we have support for MongoDB, disk and some others).

### Can the Athens proxy authenticate to private repositories?

_TL;DR: yes, with proper authentication configuration defined on the Athens proxy host._

When the GOPROXY environment variable is set on the client-side, the Go 1.11+ cli
does not attempt to request the meta tags, via a request that looks like `https://example.org/pkg/foo?go-get=1`.

Internally Athens uses `go get` under the hood (`go mod download` to be exact)
without the `GOPROXY` environment variable set so that `go` will in turn request
the meta tags using the standard authentication mechanisms supported by `go`.
Therefore, if `go` before v1.11 worked for you, then go 1.11+ with GOPROXY
should work as well, provided that the Athens proxy host is configured with the
proper authentication.

### Can I exclude a module completely?

Yes, this is possible. The proxy provides a configuration file that will allow users to specify which modules that should not be fetched at all. The [filtering modules configuration](/configuration/filter/) provides details about the configuration file and how to exclude certain modules.

### Can I specify that a module is fetched from an upstream proxy and not stored locally?

Yes, this is possible. Refer to the [filtering modules configuration](/configuration/filter/) provides details about the configuration file and how to exclude certain modules.

### Is there support for monitoring and observability for Proxy?

Right now, we have structured logs for proxy. Along with that, we have added tracing to help developers identify critical code paths and debug latency issues. While there is no setup required for logs, tracing requires some installation. We currently support exporting traces with [Jaeger](https://www.jaegertracing.io/), [GCP Stackdriver](https://cloud.google.com/stackdriver/) & [Datadog](https://docs.datadoghq.com/tracing/) (untested). Further support for other exporters is in progress.

To try out tracing with Jaeger, do the following:

- Set the environment to development (otherwise traces will be sampled)
- Run `docker-compose up -d` that is found in the athens source root directory to initialize the services required
- Run the walkthrough tutorial
- Open `http://localhost:16686/search`

  Observability is not a hard requirement for the Athens proxy. So, if the infrastructure is not properly set up, it will fail with an information log. For example, if Jaeger is not running or if the wrong URL to the exporter is provided, the proxy will continue to run. However, it will not collect any traces or metrics while the exporter backend is unavailable.

### What VCS servers does Athens support?

Athens uses `go mod download` under the hood, so it supports anything `go mod` suppports.

Which currently includes:

- git
- svn
- hg
- bzr
- fossil

### When should I use a vendor directory, and when should I use Athens?

The Go community has used vendor directories for a long time before module proxies like Athens came along, so naturally each group collaborating on code should decide for themselves whether they want to use a vendor directory, use Athens, or do both!

Using a vendor directory (without a proxy) is valuable when:

- CI/CD systems don't have access to an Athens (even if it's internal)
- When the vendor directory is so small that it is still faster to check it out from a repo than it is to pull zip files from the server
- If you're coming from glide/dep or another dependency management system that leveraged the vendor directory

Athens (without a vendor directory) is valuable when:

- You have a new project
- You are upgrading a Go project to use Go modules
- Your team requires that you use Athens (i.e. for isolation or dependency auditing)
- Your vendor directory is large and causing slow checkouts and downloading from Athens speeds the build up
  - For developers slow checkouts will not be as much of a problem as for ci tools which frequently need to checkout fresh copies of the project
- You want to remove the vendor directory from your project to:
  - Reduce noise in pull requests
  - Reduce difficulty doing fuzzy file searching in your project
