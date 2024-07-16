---
title: Filtering modules (deprecated)
description: Configuring modules that are stored on the proxy
weight: 7
---

>Note: the filter file that this page documents is deprecated. Please instead see ["Filtering with the download mode file"](/configuration/download) for updated instructions on how to filter modules in Athens.

The proxy supports the following three use cases

1. Fetches a module directly from the source (upstream proxy)
2. Exclude a particular module 
3. Include a module in the local proxy.

These settings can be done by creating a configuration file which can be pointed by setting either
`FilterFile` in `config.dev.toml` or setting `ATHENS_FILTER_FILE` as an environment variable.

### Writing the configuration file

Every line of the configuration can start either with a

* `+` denoting that the module has to be included by the proxy
* `D` denoting that the module has to be fetched directly from an upstream proxy and not stored locally
* `-` denoting that the module is excluded and will not be fetched into the proxy or from the upstream proxy

It allows for `#` to add comments and new lines are skipped. Anything else would result in an error

### Sample configuration file

<pre>
# This is a comment


- github.com/azure
+ github.com/azure/azure-sdk-for-go

# get golang tools directly
D golang.org/x/tools
</pre>

In the above example, `golang.org/x/tools` is fetched directly from the upstream proxy. All the modules from `github.com/azure` are excluded except `github.com/azure/azure-sdk-for-go`

### Adding a default mode 

The list of modules can grow quickly in size and sometimes may want to specify configuration for a handful of modules. In this case, they can set a default mode for all the modules and add specific rules to certain modules that they want to apply to. The default rule is specified at the beginning of the file. It can be an either `+`, `-` or `D`

An example default mode is 

<pre>
D
- github.com/manugupt1/athens
+ github.com/gomods/athens
</pre>

In the above example, all the modules are fetched directly from the source. `github.com/manugupt1/athens` is excluded and `github.com/gomods/athens` is stored in the proxy storage.

### Adding versions to the filter

Using an "approved list" is a common practice that requires each minor or patch version to be approved before it is allowed in the codebase.  This is accomplished by adding a list of version patterns to the rule.  These version patterns are comma-separated and prefix-matching, so `v2` and `v2.3.*` both match the requested version `2.3.5`.

An example version filter is 

<pre>
-
# use internal github enterprise server directly
D enterprise.github.com/company

# external dependency approved list
+ github.com/gomods/athens v0.1,v0.2,v0.4.1
</pre>

In the above example, any module not in the rules will be excluded.  All modules from `enterprise.github.com/company` are fetched directly from the source.  The `github.com/gomods/athens` module will be stored in the proxy storage, but only for version `v0.4.1` and any patch versions under `v0.1` and `v0.2` minor versions.

### Versions Filter Modifiers

Athens provides advanced filter modifiers to cover cases such as API compatibility or when a given dependency changes its license from a given versions. The modifiers are intended to be used in the pattern list of the filter file.

<pre>
-
# external dependency approved list
+ github.com/gomods/athens <v1.2.3
</pre>

The currently supported modifiers are 

* `~1.2.3` will enable all patch versions from 1.2.3 and above (e.g. 1.2.3, 1.2.4, 1.2.5)
  * Formally, `1.2.x` where `x >= 3`

* `^1.2.3` will enable all patch and minor versions from 1.2.3 and above (e.g. 1.2.4, 1.3.0 and 1.4.5)
  * Formally, `1.x.y` where `x >= 2` and `y >= 3`

* `<1.2.3` will enable all versions lower than 1.2.3 (e.g. 1.2.2, 1.0.0 and 0.58.9)
  * Formally, `x.y.z` where `x <= 1`, `y < = 2` and `z < 3`

This kind of modifiers will work only if a three parts semantic version is specified. For example, `~4.5.6` will work while `~4.5` won't.
