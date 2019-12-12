---
title: Proxying a checksum database API
description: How to configure Athens to proxy a checksum database API, and why you might want to.
weight: 4
---

If you run `go get github.com/mycompany/secret-repo@v1.0.0` and that module version is not yet in your `go.sum` file, Go will by default send a request to `https://sum.golang.org/lookup/github.com/mycompany/secret-repo@v1.0.0`. That request will fail because the Go tool requires a checksum, but `sum.golang.org` doesn't have access to your private code.

The result is that *(1) your build will fail*, and *(2) your private module names have been sent over the internet to an opaque public server that you don't control*.

>You can read more about this `sum.golang.org` service [here](https://go.googlesource.com/proposal/+/master/design/25530-sumdb.md)

## Proxying a checksum DB

Many companies use Athens to host their private code, but Athens is not only a module proxy. It's also a checksum database proxy. That means that anyone inside of your company can configure `go` to send these checksum requests to Athens instead of the public `sum.golang.org` server.

If the Athens server is configured with checksum filters, then you can prevent these problems.

If you run the below command using Go 1.13 or later:

```bash
$ GOPROXY=<athens-url> go build .
```

... then the Go tool will automatically send all checksum requests to `<athens-url>/sumdb/sum.golang.org` instead of `https://sum.golang.org`.

By default, when Athens receives a `/sumdb/...` request, it automatically proxies it to `https://sum.golang.org`, even if it's a private module that `sum.golang.org` doesn't and can't know about. So if you are working with private modules, you'll want to change the default behavior.

>If you Athens to _not_ send some module names up to the global checksum database, set those module names in the `NoSumPatterns` value in `config.toml` or using the `ATHENS_GONOSUM_PATTERNS` environment variable.

The following sections will go into more detail on how checksum databases work, how Athens fits in, and how this all impacts your workflow.

## How to set this all up

Before you begin, you'll need to run Athens with configuration values that tell it to not proxy certain modules. If you're using `config.toml`, use this configuration:

```toml
NoSumPatterns = ["github.com/mycompany/*", "github.com/secret/*"]
```

And if you're using an environment variable, use this configuration:

```bash
$ export ATHENS_GONOSUM_PATTERNS="github.com/mycompany/*,github.com/secret/*"
```

>You can use any string compatible with [`path.Match`](https://pkg.go.dev/path?tab=doc#Match) in these environment variables

After you start Athens up with this configuration, all checksum requests for modules that start with `github.com/mycompany` or `github.com/secret` will not be forwarded, and Athens will return an error to the `go` CLI tool. 

This behavior will ensure that none of your private module names leak to the public internet, but your builds will still fail. To fix that problem, set another environment variable on your machine (that you run your `go` commands)

```bash
$ export GONOSUMDB="github.com/mycompany/*,github.com/secret/*"
```

Now, your builds will work and you won't be sending information about your private codebase to the internet.

## I'm confused, why is this hard?

When the Go tool has to download _new_ code that isn't currently in the project's `go.sum` file, it tries its hardest to get a checksum from a server it trusts, and compare it to the checksum in the actual code it downloads. It does all of this to ensure _provenance_. That is, to ensure that the code you just downloaded wasn't tampered with.

The trusted checksums are all stored in `sum.golang.org`, and that server is centrally controlled.

>These build failures and potential privacy leaks can only happen when you try to get a module version that is _not_ already in your `go.sum` file.

Athens does its best to respect and use the trusted checksums while also ensuring that your private names don't get leaked to the public server. In some cases, it has to choose whether to fail your build or leak information, so it chooses to fail your build. That's why everybody using that Athens server needs to set up their `GONOSUMDB` environment variable.

We believe that along with good documentation - which we hope this is! - we have struck the right balance between convenience and privacy.