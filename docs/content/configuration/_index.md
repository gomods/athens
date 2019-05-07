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
