# Development Guide for Athens

The proxy is built on the [Buffalo](https://gobuffalo.io/) framework. We chose
this framework to make it as straightforward as possible to get your development environment up and running.
However, **you do not need to install buffalo to run Athens**. 

Buffalo provides nice features like a file watcher for your server, so if you'd like to install Buffalo, download [v0.12.4](https://github.com/gobuffalo/buffalo/releases/tag/v0.12.4) or later to get started on Athens and be sure to download the CLI and put it into your `PATH`.

Athens uses [Go Modules](https://golang.org/cmd/go/#hdr-Modules__module_versions__and_more) for dependency management. You will need Go [v1.11](https://golang.org/dl) or later to get started on Athens.

See our [Contributing Guide](CONTRIBUTING.md) for tips on how to submit a pull request when you are ready.

## Go Version

You'll need Go version 1.11 or higher to build and run Athens. Since it uses the `go` CLI internally, you can tell it which CLI to use via the `GO_BINARY_PATH` environment. Set it to the binary of choice like this:

```console
GO_BINARY_PATH=/path/to/go
```

# Run the Proxy Natively

After you've set up your Go version, the `buffalo` CLI makes it easy to launch the proxy: 

```console
cd cmd/proxy
buffalo dev
```

If you don't have Buffalo installed, you can also use the `go` command directly like this:

```console
cd cmd/proxy
go build
./proxy
```

Whichever way you start the server, you'll see some console output like:

```console
Starting application at 127.0.0.1:3000
```

... and then you'll be ready to use Athens!

# Services that Athens Needs

By default, Athens doesn't require any external services to be running. By default, it uses in-memory storage and you can also opt into on-disk storage, and neither method requires any external services.

However, Athens does support other storage drivers, and takes advantage of other external services that can be helpful for more permanent and "production-like" deployments. Here are some of the services that Athens can take advantage of:

- [Minio]((https://minio.io/)) storage driver
- [MongoDB](https://www.mongodb.com/) storage driver
- [Jaeger](https://www.jaegertracing.io/) for distributed tracing
- [Datadog](https://www.datadoghq.com/) for monitoring

## How Do I Run All These Services Locally?

We use [Docker](https://docker.com) images and [Docker Compose](https://docs.docker.com/compose/) to start up Athens and all of those services, and connect them all together. Thanks to all that magic, you can start up all an Athens server alongside all its services with one command! :tada: :rocket:

If you're not familiar with Docker, that's ok. In the spirit of Buffalo, we've tried to make
it easy to get up and running:

1. [Download and install docker-compose](https://docs.docker.com/compose/install/)
1. Run `make dev` from the root of this repository
1. Access Athens on `http://localhost:3000`

That's it! Athens will take a few seconds to configure and connect to all the services it uses. When it's done, everything will be up and running and ready to use.

If you want to stop everything at any time, run `make down`.

>`make down` will stop all the resources that `make dev` started, and clean up

# Run Unit Tests

In order to run unit tests, services they depend on must be running first:

```console
make alldeps
```

then you can run the unit tests:

```console
make test-unit
```

# Build The Docs

To get started with developing the docs we provide a docker image, which runs [Hugo](https://gohugo.io/) to render the docs. Using the docker image, we mount the `/docs` directory into the container. To get it up and running, from the project root run:

```console
$ make docs
$ docker run -it --rm \
        --name hugo-server \
        -p 1313:1313 \
        -v ${PWD}/docs:/src:cached \
        gomods/hugo
```

Then open [http://localhost:1313](http://localhost:1313/).

# Run the Linter on the Code

In our CI/CD pass, we use golint, so feel free to install and run it locally beforehand:

```go
go get golang.org/x/lint/golint
```
