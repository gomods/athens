# Development Guide for Athens

The proxy is built on the [Buffalo](https://gobuffalo.io/) framework. We chose
this framework to make it as straightforward as possible to get your development environment up and running.
However, **you do not need to install buffalo to run Athens**. 

Buffalo provides nice features like a file watcher for your server, so if you'd like to install Buffalo, download [v0.12.4](https://github.com/gobuffalo/buffalo/releases/tag/v0.12.4) or later to get started on Athens,
so be sure to download the CLI and put it into your `PATH`.

Athens uses [Go Modules](https://golang.org/cmd/go/#hdr-Modules__module_versions__and_more) for dependency management. You will need Go [v1.11](https://golang.org/dl) or later to get started on Athens.

See our [Contributing Guide](CONTRIBUTING.md) for tips on how to submit a pull request when you are ready.

### Go version
Athens is developed on Go1.11+.

To point Athens to a different version of Go set the following environment variable
```
GO_BINARY_PATH=go1.11.X
or whichever binary you want to use with athens
```

# Run the Proxy
After you've set up your dependencies, the `buffalo` CLI makes it easy to launch the proxy: 

```console
cd cmd/proxy
buffalo dev
```

If you don't have Buffalo installed, you can just use the `go` command directly as such: 

```
cd cmd/proxy
go build
./proxy
```

After the server starts, you'll see some console output like:

```console
Starting application at 127.0.0.1:3000
```

### Dependencies

# Services that Athens Needs

Athens relies on several services (i.e. databases, etc...) to function properly. We use [Docker](http://docker.com/) images to configure and run those services. **However, Athens does not require any storage dependencies by default**. The default storage is in memory, you can opt-in to using the `fs` which would also require no dependencies. But if you'd like to test out Athens against a real storage backend (such as MongoDB, Minio, S3 etc), continue reading this section:

If you're not familiar with Docker, that's ok. In the spirit of Buffalo, we've tried to make
it easy to get up and running:

1. [Download and install docker-compose](https://docs.docker.com/compose/install/) (docker-compose is a tool for easily starting and stopping lots of services at once)
2. Run `make dev` from the root of this repository

That's it! After the `make dev` command is done, everything will be up and running and you can move
on to the next step.

If you want to stop everything at any time, run `make down`.

Note that `make dev` only runs the minimum amount of dependencies needed for things to work. If you'd like to run all the possible dependencies run `make alldeps` or directly the services available in the `docker-compose.yml` file. Keep in mind, though, that `make alldeps` does not start up Athens or Oympus, but **only** their dependencies.

# Run unit tests

In order to run unit tests, services they depend on must be running first:

```console
make alldeps
```

then you can run the unit tests:

```console
make test-unit
```

# Run the docs

To get started with developing the docs we provide a docker image which you can use from within the `/docs` directory. It should work on all platforms. To get it up and running:

```
docker run -it --rm \
        --name hugo-server \
        -p 1313:1313 \
        -v $(PWD):/src:cached \
        gomods/hugo
        
```
