---
title: Download Mode
description: What to do when a module is not in storage
weight: 2
---

Athens accepts an HCL formatted Download File that has instructions for how it should behave when a module@version isn't found in its storage.

You configure this download file by setting the `ATHENS_DOWNLOAD_MODE` environment variable in one of two ways:

1. Set its value to `file:$FILE_PATH`, where `FILE_PATH` is the path to the HCL file
1. Set its value to `custom$BASE_64` where `BASE_64` is the base64 encoded HCL file

## What should Athens do when a module@version is not found in storage?

Say a client sends an HTTP request with the path `/github.com/pkg/errors/@v/v0.8.1` and Athens
does not have this module in storage. Athens will look at the Download File for one of the following Modes:

1. **`sync`**: Synchronously download the module from VCS via `go mod download`, persist it to the Athens storage, and serve it back to the user immediately. Note that this is the default behavior. 
2. **`async`**: Return a 404 to the client, and asynchronously download and persist the module@version to storage.
3. **`none`**: Return a 404 and do nothing. 
4. **`redirect`**: Redirect to an upstream proxy (such as proxy.golang.org) and do nothing after.
5. **`async_redirect`**: Redirect to an upstream proxy (such as proxy.golang.org) and asynchronously download and persist the module@version to storage. 

Furthermore, the Download File can describe any of the above behavior for different modules and module patterns alike using [path.Match](https://golang.org/pkg/path/#Match). Take a look at the following example:

```javascript
downloadURL = "https://proxy.golang.org"

mode = "async_redirect"

download "github.com/gomods/*" {
    mode = "sync"
}

download "golang.org/x/*" {
    mode = "none"
}

download "github.com/pkg/*" {
    mode = "redirect"
    downloadURL = "https://gocenter.io"
}
```

The first two lines describe the behavior and the destination of all packages: redirect to `https://proxy.golang.org` and asynchronously persist the module to storage. 

The following two blocks describe what to do if the requested module matches the given pattern:

Any module that matches "github.com/gomods/*" such as "github.com/gomods/athens", will be synchronously fetched, stored, and returned to the user. 

Any module that matches "golang.org/x/*" such as "golang.org/x/text" will just return a 404. Note that this behavior allows the client to set GOPROXY to multiple comma separated proxies so that the Go command can move into the second argument. 

Any module that matches "github.com/pkg/*" such as "github.com/pkg/errors" will be redirected to https://gocenter.io (and not proxy.golang.org) and will also never persist the module to the Athens storage.


## Use cases

So why would you want to use the Download File to configure the behavior above? Here are a few use cases where it might make sense for you to do so: 

**Limited storage:**

If you have limited storage, then it might be a good idea to only persist some moduels (such as private ones) and then redirect to a public proxy for everything else. 

**Limited resources:**

If you are running Athens with low memory/cpu, then you can redirect all public modules to proxy.golang.org but asynchronously fetch them so that the client does not timeout. At the same time, you can return a 404 for private modules through the `none` mode and let the client (the Go command) fetch private modules directly through `GOPROXY=<athens-url>,direct` 
