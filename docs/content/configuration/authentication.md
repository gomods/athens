---
title: Authentication to private repositories
description: Configuring Authentication on Athens
weight: 2
---

## Authentication

## SVN private repositories

1.  Subversion creates an authentication structure in

        ~/.subversion/auth/svn.simple/<hash>

2.  In order to properly create the authentication file for your SVN servers you will need to authenticate to them and let svn build out the proper hashed files.

        $ svn list http://<domain:port>/svn/<somerepo>
        Authentication realm: <http://<domain> Subversion Repository
        Username: test
        Password for 'test':

3.  Once we've properly authenticated we want to share the .subversion directory with the Athens proxy server in order to reuse those credentials. Below we're setting it as a volume on our proxy container.

    **Bash**

    ```bash
    export ATHENS_STORAGE=~/athens-storage
    export ATHENS_SVN=~/.subversion
    mkdir -p $ATHENS_STORAGE
    docker run -d -v $ATHENS_STORAGE:/var/lib/athens \
      -v $ATHENS_SVN:/root/.subversion \
      -e ATHENS_DISK_STORAGE_ROOT=/var/lib/athens \
      -e ATHENS_STORAGE_TYPE=disk \
      --name athens-proxy \
      --restart always \
      -p 3000:3000 \
      gomods/athens:latest
    ```

    **PowerShell**

    ```PowerShell
    $env:ATHENS_STORAGE = "$(Join-Path $pwd athens-storage)"
    $env:ATHENS_SVN = "$(Join-Path $pwd .subversion)"
    md -Path $env:ATHENS_STORAGE
    docker run -d -v "$($env:ATHENS_STORAGE):/var/lib/athens" `
      -v "$($env:ATHENS_SVN):/root/.subversion" `
      -e ATHENS_DISK_STORAGE_ROOT=/var/lib/athens `
      -e ATHENS_STORAGE_TYPE=disk `
      --name athens-proxy `
      --restart always `
      -p 3000:3000 `
      gomods/athens:latest
    ```

## Bazaar(bzr) private repositories

* Bazaar is not supported with the Dockerfile provided by Athens. but the instructions are valid for custom Athens build with bazaar.*

1. Bazaaar config files are located in

- Unix

      ~/.bazaar/

- Windows

      C:\Documents and Settings\<username>\Application Data\Bazaar\2.0

- You can check your location using

      bzr version

2. There are 3 typical configuration files

- bazaar.conf
  - default config options
- locations.conf
  - branch specific overrides and/or settings
- authentication.conf
  - credential information for remote servers

3. Configuration file syntax

- \# this is a comment
- [header] this denotes a section header
- section options reside in a header section and contain an option name an equals sign and a value

  - EXAMPLE:

        [DEFAULT]
        email = John Doe <jdoe@isp.com>

4. Authentication Configuration

   Allows one to specify credentials for remote servers.
   This can be used for all the supported transports and any part of bzr that requires authentication(smtp for example).
   The syntax obeys the same rules as the others except for the option policies which don't apply.

   Example:

   [myprojects]
   scheme=ftp
   host=host.com
   user=joe
   password=secret

   # Pet projects on hobby.net

   [hobby]
   host=r.hobby.net
   user=jim
   password=obvious1234

   # Home server

   [home]
   scheme=https
   host=home.net
   user=joe
   password=lessobV10us

   [DEFAULT]

   # Our local user is barbaz, on all remote sites we're known as foobar

   user=foobar

   NOTE: when using sftp the scheme is ssh and a password isn't supported you should use PPK

   [reference code]
   scheme=https
   host=dev.company.com
   path=/dev
   user=user1
   password=pass1

   # development branches on dev server

   [dev]
   scheme=ssh # bzr+ssh and sftp are availablehere
   host=dev.company.com
   path=/dev/integration
   user=user2

   #proxy
   [proxy]
   scheme=http
   host=proxy.company.com
   port=3128
   user=proxyuser1
   password=proxypass1

5. Once we've properly setup our authentication we want to share the bazaar configuration directory with the Athens proxy server in order to reuse those credentials. Below we're setting it as a volume on our proxy container.

   **Bash**

   ```bash
   export ATHENS_STORAGE=~/athens-storage
   export ATHENS_BZR=~/.bazaar
   mkdir -p $ATHENS_STORAGE
   docker run -d -v $ATHENS_STORAGE:/var/lib/athens \
     -v $ATHENS_BZR:/root/.bazaar \
     -e ATHENS_DISK_STORAGE_ROOT=/var/lib/athens \
     -e ATHENS_STORAGE_TYPE=disk \
     --name athens-proxy \
     --restart always \
     -p 3000:3000 \
     gomods/athens:latest
   ```

   **PowerShell**

   ```PowerShell
   $env:ATHENS_STORAGE = "$(Join-Path $pwd athens-storage)"
   $env:ATHENS_BZR = "$(Join-Path $pwd .bazaar)"
   md -Path $env:ATHENS_STORAGE
   docker run -d -v "$($env:ATHENS_STORAGE):/var/lib/athens" `
     -v "$($env:ATHENS_BZR):/root/.bazaar" `
     -e ATHENS_DISK_STORAGE_ROOT=/var/lib/athens `
     -e ATHENS_STORAGE_TYPE=disk `
     --name athens-proxy `
     --restart always `
     -p 3000:3000 `
     gomods/athens:latest
   ```

## Atlassian Bitbucket and SSH-secured git VCS's

This section was originally written to describe configuring the
Athens git client to fetch specific Go imports over SSH instead of
HTTP against an on-prem instance of Atlassian Bitbucket.  With some
adjustment it may point the way to configuring the Athens proxy for
authenticated access to hosted Bitbucket and other SSH-secured
VCS's. If your developer workflow requires that you clone, push,
and pull Git repositories over SSH and you want Athens to perform
the same way, please read on.

As a developer at example.com, assume your application has a
dependency described by this import which is hosted on Bitbucket

```go
import "git.example.com/golibs/logo"
```

Further, assume that you would manually clone this import like this

```bash
$ git clone ssh://git@git.example.com:7999/golibs/logo.git
```

A `go-get` client, such as that called by Athens, would [begin
resolving](https://golang.org/cmd/go/) this dependency by looking
for a `go-import` meta tag in this output

```bash
$ curl -s https://git.example.com/golibs/logo?go-get=1
<?xml version="1.0"?>
<!DOCTYPE html>
<html lang="en">
   <head>
      <meta charset="utf-8">
         <meta name="go-import" content="git.example.com/golibs/logo git https://git.example.com/scm/golibs/logo.git"/>
         <body/>
      </meta>
   </head>
</html>
```

which says the content of the Go import actually resides at
`https://git.example.com/scm/golibs/logo.git`. Comparing this URL
to what we would normally use to clone this project over SSH (above)
suggests this [global Git config](https://git-scm.com/docs/git-config)
http to ssh rewrite rule

```
[url "ssh://git@git.example.com:7999"]
	insteadOf = https://git.example.com/scm
```

So to fetch the `git.example.com/golibs/logo` dependency over SSH
to populate its storage cache, Athens ultimately calls git, which,
given that rewrite rule, in turn needs an SSH private key matching
a public key bound to the cloning developer or service account on
Bitbucket. This is essentially the github.com SSH model.  At a bare
minimum, we need to provide Athens with an SSH private key and the
http to ssh git rewrite rule, mounted inside the Athens container
for use by the root user

```bash
$ mkdir -p storage
$ ATHENS_STORAGE=storage
$ docker run --rm -d \
    -v "$PWD/$ATHENS_STORAGE:/var/lib/athens" \
    -v "$PWD/gitconfig/.gitconfig:/root/.gitconfig" \
    -v "$PWD/ssh-keys:/root/.ssh" \
    -e ATHENS_DISK_STORAGE_ROOT=/var/lib/athens -e ATHENS_STORAGE_TYPE=disk --name athens-proxy -p 3000:3000 gomods/athens:canary
```

`$PWD/gitconfig/.gitconfig` contains the http to ssh rewrite rule

```
[url "ssh://git@git.example.com:7999"]
	insteadOf = https://git.example.com/scm
```

`$PWD/ssh-keys` contains the aforementioned private key and a minimal ssh-config

```bash
$ ls ssh-keys/
config		id_rsa
```

We also provide an ssh config to bypass host SSH key verification
and to show how to bind different hosts to different SSH keys

`$PWD/ssh-keys/config` contains

```
Host git.example.com
Hostname git.example.com
StrictHostKeyChecking no
IdentityFile /root/.ssh/id_rsa
```

Now, builds executed through the Athens proxy should be able to clone the `git.example.com/golibs/logo` dependency over authenticated SSH.

### `SSH_AUTH_SOCK` and `ssh-agent` Support

As an alternative to passwordless SSH keys, one can use an [`ssh-agent`](https://en.wikipedia.org/wiki/Ssh-agent).
The `ssh-agent`-set `SSH_AUTH_SOCK` environment variable will propagate to
`go mod download` if it contains a path to a valid unix socket (after
following symlinks).

As a result, if running with a working ssh agent (and a shell with
`SSH_AUTH_SOCK` set), after setting up a `gitconfig` as mentioned in the
previous section, one can run athens in docker as such:


```bash
$ mkdir -p storage
$ ssh-add .ssh/id_rsa_something
$ ATHENS_STORAGE=storage
$ docker run --rm -d \
    -v "$PWD/$ATHENS_STORAGE:/var/lib/athens" \
    -v "$PWD/gitconfig/.gitconfig:/root/.gitconfig" \
    -v "${SSH_AUTH_SOCK}:/.ssh_agent_sock" \
    -e "SSH_AUTH_SOCK=/.ssh_agent_sock" \
    -e ATHENS_DISK_STORAGE_ROOT=/var/lib/athens -e ATHENS_STORAGE_TYPE=disk --name athens-proxy -p 3000:3000 gomods/athens:canary
```

## GitHub Apps

Instead of using a Machine User on GitHub, it is possible to create a GitHub App and authenticate via it. 

Create a GitHub App in **Settings > Developer settings > GitHub Apps** and install it. The AppID/ClientID, Installation ID and Private Key are
required from the App.

Install the [GitHub App Git Credential Helper](https://github.com/bdellegrazie/git-credential-github-app) in your `$PATH`. The Athens Docker image comes
with this pre-installed.

Configure your [global Git config](https://git-scm.com/docs/git-config) as follows:

```
[credential "https://github.com/your-org"]
    helper = "github-app -username <app-name> -appId <app-id> -privateKeyFile <path-to-private-key> -installationId <installation-id>"
    useHttpPath = true

[credential "https://github.com"]
    helper = "cache --timeout=3600"

[url "https://github.com"]
    insteadOf = ssh://git@github.com
```

This instructs Git to authenticate with the GitHub App and cache the results for 3600s (the authentication token is valid for 1 hour).

Now, builds executed through the Athens proxy should be able to clone the `github.com/your-org/your-repo` dependency over GitHub Apps.

### GitHub Enterprise Self-hosted

To authenticate against a self-hosted GitHub Enterprise, the instructions are the same for GitHub hosted Apps
with the exception for the Git config, which should include your domain, as follows:

```
[credential "https://github.example.com/your-org"]
    helper = "github-app -username <app-name> -appId <app-id> -privateKeyFile <path-to-private-key> -installationId <installation-id> -domain github.example.com"
    useHttpPath = true

[credential "https://github.example.com"]
    helper = "cache --timeout=3600"

[url "https://github.example.com"]
    insteadOf = ssh://git@github.com
```

