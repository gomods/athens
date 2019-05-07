---
title: "Configuring Athens"
date: 2018-10-16T12:14:01-07:00
weight: 3
---

## Configuring Athens
Here we'll cover how to configure the Athens application utilizing various configuration scenarios.

### Authentication
There are numerous version control systems available to us as developers.  In this section we'll outline how they can be used by supplying required credentials in various formats for the Athens project.

 - [Authentication](/configuration/authentication)
 
### Storage
In Athens we support many storage options. In this section we'll describe how they can be configured

 - [Storage](/configuration/storage)


 ### Upstream proxy
 In this section we'll describe how the upstream proxy can be configured to fetch all modules from a Go Modules Repository such as [GoCenter](https://gocenter.io) or another Athens Server.

  - [Upstream](/configuration/upstream)

### Proxying A Checksum DB
The Athens Proxy has the ability to prox ya Checksum Database as defined by [this proposal](https://go.googlesource.com/proposal/+/master/design/25530-sumdb.md) by the GO team.

Athens by defualt will accept proxying `https://sum.golang.org`. However, if you'd like to override that behavior or proxy more Checksum DBs you can do so through the `SumDBs` config or its equivalent Enviroenment Variable: `ATHENS_SUM_DBS` 

So for example, if you run the following command: 

```bash
GOPORXY=<athens-url> go build
```

The Go command will proxy a requests to `sum.golang.org` like this: `<athens-url>/sumdb/sum.golang.org`. Feel free to read the linked proposal above for the exact requests that makes Athens succesfully proxy Checksum DB APIs. 

Note that as of this documentation (May 2019), you need to explicitly set `GOSUMDB=https://sum.golang.org`, but the Go team is planning on enabling this by defualt. 

### Why a Checksum DB? 

The reasons for needing a Checksum DB is explained in the linked proposal above. However, the reasons for proxying a Checksum DB are more explained below. 

### Why Proxy a Checksum DB? 

This is quite important. Say you are a company that is running an Athens instance, and you don't want the world to konw about where your 
repositories live. For example, say you have a private repo under `github.com/mycompany/secret-repo`. In order to ensure that the Go client 
does not send a request to `https://sum.golang.org/look/github.com/mycompany/secret-repo@v1.0.0` and therefore leaking your private import path to the public, you need to ensure that you tell Go to skip particular import paths as such: 

```
GONOSUMDB=github.com/mycompany/* go build
```

This will make sure that Go does not send any requests to the Checksum DB for your private import paths. 
However, how can you ensure that all of your employees are building private code with the right confiugration? 

Athens, in this case can help ensure that all private code flowing through it never goes to the Checksum DB. So as long as your employees are using Athens, then they will get a helpful reminder to ensure Their GONOSUMDB is rightly configured. 

As the Athens company maintainer, you can run Athens with the following configuration: 

`NoSumPatterns = ["github.com/mycompany/*] # or comma separted env var: ATHENS_GONOSUM_PATTERNS`

This will ensure that when Go sends a request to `<athens-url/sumdb/sum.golang.org/github.com/mycompany/secret-repo@v1.0.0>`, Athens will return a 403 and failing the build ensuring that the client knows something is not configured correctly and also never leaking those import paths