---
title: Checksum DB
description: Proxying A Checksum DB API
weight: 2
---

## Proxying A Checksum DB
The Athens Proxy has the ability to proxy a Checksum Database as defined by [this proposal](https://go.googlesource.com/proposal/+/master/design/25530-sumdb.md) by the Go team.

Athens by defualt will accept proxying `https://sum.golang.org`. However, if you'd like to override that behavior or proxy more Checksum DBs you can do so through the `SumDBs` config or its equivalent Environment Variable: `ATHENS_SUM_DBS` 

So for example, if you run the following command: 

```bash
GOPROXY=<athens-url> go build
```

The Go command will proxy requests to `sum.golang.org` like this: `<athens-url>/sumdb/sum.golang.org`. Feel free to read the linked proposal above for the exact requests that makes Athens successfully proxy Checksum DB APIs. 

Note that as of this documentation (May 2019), you need to explicitly set `GOSUMDB=https://sum.golang.org`, but the Go team is planning on enabling this by defualt. 

### Why a Checksum DB? 

The reasons for needing a Checksum DB is explained in the linked proposal above. However, the reasons for proxying a Checksum DB are more explained below. 

### Why Proxy a Checksum DB? 

This is quite important. Say you are a company that is running an Athens instance, and you don't want the world to konw about where your 
repositories live. For example, say you have a private repo under `github.com/mycompany/secret-repo`. In order to ensure that the Go client 
does not send a request to `https://sum.golang.org/lookup/github.com/mycompany/secret-repo@v1.0.0` and therefore leaking your private import path to the public, you need to ensure that you tell Go to skip particular import paths as such: 

```
GONOSUMDB=github.com/mycompany/* go build
```

This will make sure that Go does not send any requests to the Checksum DB for your private import paths. 
However, how can you ensure that all of your employees are building private code with the right configuration? 

Athens, in this case can help ensure that all private code flowing through it never goes to the Checksum DB. So as long as your employees are using Athens, then they will get a helpful reminder to ensure Their GONOSUMDB is rightly configured. 

As the Athens company maintainer, you can run Athens with the following configuration: 

`NoSumPatterns = ["github.com/mycompany/*] # or comma separted env var: ATHENS_GONOSUM_PATTERNS`

This will ensure that when Go sends a request to `<athens-url/sumdb/sum.golang.org/github.com/mycompany/secret-repo@v1.0.0>`, Athens will return a 403 and failing the build ensuring that the client knows something is not configured correctly and also never leaking those import paths