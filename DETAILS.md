# Details on Athens functionality.

This document describes the various use cases that Athens supports or will
support. Not everything here is implemented.

[@bketelsen](https://github.com/bketelsen) and I spoke at length on Slack about
how the Athens registry should work. I'm putting our discussion down on "paper"
and adding some of my own detail (and some opinions) so we can grow this into
a definitive document.

# Where Will the Code Come From?

Here's what was discussed on slack + various other forums:

## Travis or Other CI systems (Webhook)

After code passes tests, the publisher can hit a webhook to tell the registry to fetch
code for the given tag from the given repository. The webhook will then do the following:

* Check if the repository is already "known" (the owner of the repository will need to register it prior) - Fail if it isn't
* Check if the given tag already exists (we want versions to be immutable) - Fail if it does
* Download the code from the given tag and store it in a CDN

We may need to require users to be authorized to hit the webhooks.

## Manual Uploads

This works similarly to the webhook flow, except a user would upload their module --
including code -- manually to the registry. When they upload, they'll need to
specify the module name and version, and obviously include the zip file with their
source code.

The zip file will need to have the `go.mod` file in it, and the server will
parse it out.

Athens will ship with a CLI to construct and upload the zip file.

# How Will Modules Be Served?

There are two modes of serving modules. One involves using the `GOPROXY` environment variable,
and the other doesn't.

## Using `GOPROXY`

When a user sets `GOPROXY` on their system, the `vgo` tool automatically prefixes all modules
with that value when it makes requests. For example,
`GOPROXY=localhost:3000 go get my/thing` will execute the download protocol against
`localhost:3000`.

With this in mind, we'll give users two `GOPROXY` related workflows.

### Workflow I - Local Development

We'll give users a CLI that runs a proxy server that passes through to a tree on their
local filesystem. The workflow will be the following:

```console
$ athens run local-proxy --root /someplace --vcs git
$ GOPROXY=localhost:1234 go get my/module
```

The `go get` command executes the download protocol on the local proxy, which simply mirrors
to the files in `/someplace`. In the case of `my/module`, it mirrors to `/someplace/my/module`, requires the `go.mod` to be there, and uses `git` (specified by `--vcs`) to list
tags, etc...

### Workflow II - Proxying to Github Repositories

We'll give users the ability to `go get` modules against our `gomods.io` server that are
really mirrored from Github. In the webhook flow (above), we already have the capability
to fetch code from a Github repository, so we can know the `github.com/...` package name.

So, we can make the following work properly:

```console
$ GOPROXY=gomods.io go get github.com/arschles/thing@v1
```

This not only executes the download protocol against `gomods.io`, it also means that the
import path for the module is `github.com/arschles/thing`, and that will certainly help
people migrate their codebases from previous dependency managers to vgo.

I'd prefer that our registry serves the download protocol directly (using a `<meta>` tag redirect, more on that below)
so that we can still serve from our CDN and protect against repositories disappearing. I like the idea technically, but
I don't know about it from a community perspective.

## "Vanity" Module Names

We've been able to specify "vanity" names (names that don't match the repository the
code lives in) for our packages for a while not. https://gopkg.in was one of the earliest
systems that let us do this. It "redirected" (using the `<meta>` tag) the `go get`
requests against it to the appropriate repository in Github. The `go get` tool would
then follow the redirect and use `git` to fetch the code.

Now that there's a custom HTTP-based download protocol, we can use the same rename-by-redirecting mechanism, but with the following additions:

* Allow users to specify their own "vanity" module names for a given codebase - We should probably allow 0 or more vanity names attached to an existing repository - We should allow users to turn off the repository-based proxy name (i.e. the `github.com`
  name)
* Redirect to a CDN instead of to Github

Given that new functionality, the following workflow is possible (assuming the registry
server lives at `gomods.io`):

```console
$ go get gomods.io/my/package@v1
```

The new caninocal import path for this package is `gomods.io/my/package`. This also allows for import paths to be
specified in packages, for example:

```go
package captainhook // import gomods.io/captainhook
```

We can also enforce naming schemes, naming groups, organizations, privacy settings, etc...
on the server side if we like.
