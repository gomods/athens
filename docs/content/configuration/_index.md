---
title: "Configuring Athens"
date: 2018-10-16T12:14:01-07:00
weight: 3
---

## Configuring Athens
Here we'll cover how to configure the Athens application utilizing various configuration scenarios.

>This section covers some of the more commonly used configuration variables, but there are more! If you want to see all the configuration variables you can set, we've documented them all in [this configuration file](https://github.com/gomods/athens/blob/main/config.dev.toml).

### Authentication
There are numerous version control systems available to us as developers.  In this section we'll outline how they can be used by supplying required credentials in various formats for the Athens project.

 - [Authentication](/configuration/authentication)
 
### Storage
In Athens we support many storage options. In this section we'll describe how they can be configured

 - [Storage](/configuration/storage)

### Upstream proxy
In this section we'll describe how the upstream proxy can be configured to fetch all modules from a Go Modules Repository such as [GoCenter](https://gocenter.io), [The Go Module Mirror](https://proxy.golang.org), or another Athens Server.

  - [Upstream](/configuration/upstream)

### Proxying A Checksum DB
In this section we'll describe how to proxy a Checksum DB as per https://go.googlesource.com/proposal/+/master/design/25530-sumdb.md

- [Checksum](/configuration/sumdb)

### Accessing the configuration

By providing an optional `ConfigPort` parameter, Athens will expose a `/config` service endpoint that will return the configuration used by the proxy. 

##### Configuration:

    # ConfigPort optionally enables a config endpoint on the given port
    # Env override: ATHENS_CONFIG_PORT
    ConfigPort = ":3001"

The service is exposed on a different port, as the configuration may contain sensitive informations and different access rules may be implemented (i.e., access on the config port may only be granted locally).
