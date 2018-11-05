---
title: Filtering modules
description: Configuring modules that are stored on the proxy
weight: 1
---

It is very easy to exclude certain modules to be cached in the proxy. There are two ways in which the proxy can be configured.

1. When a single instance of the proxy that does not point to a global proxy.
2. When a single instance of the proxy also points to a global proxy.

A user may either want to 
1. Fetch a module directly from the source
2. Fetch a module from the global proxy 
3. Include a module in the local proxy.

These settings can be done by creating a configuration file which can be pointed by setting either
`FilterFile` in `config.dev.toml` or setting `ATHENS_FILTER_FILE` as an environment variable.

### Writing the configuration file

Every line of the configuration can start either with a

* `+` denoting that the module has to be included by the proxy,
* `-` denoting that the module does not have to be included in the proxy
* `D` denoting that the module has to be fetched directly from the source

It allows for `#` to add comments and new lines are skipped. Anything else would result in an error

### Sample configuration file

<pre>
# This is a comment


E github.com/manugupt1/athens
I github.com/gomods/walkthrough

# get golang tools directly
D golang.org/x/tools
</pre>


### Adding a default mode 

The list of modules can grow quickly in size and sometimes may want to specify configuration for a handful of modules. In this case, they can set a default mode for all the modules and add specific rules to certain modules that they want to apply to. The default rule is specified at the beginning of the file. It can be an either `+`, `-` or `D`

An example default mode is 

<pre>
D
- github.com/manugupt1/athens
+ github.com/gomods/athens
</pre>