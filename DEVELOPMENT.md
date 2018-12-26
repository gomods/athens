# Development Guide for Athens

The proxy is written in idiomatic Go and uses standard tools. If you know Go, you'll be able to read the code and run the server.

Athens uses [Go Modules](https://golang.org/cmd/go/#hdr-Modules__module_versions__and_more) for dependency management. You will need Go [v1.11](https://golang.org/dl) or later to get started on Athens.

See our [Contributing Guide](CONTRIBUTING.md) for tips on how to submit a pull request when you are ready.

### Go version
Athens is developed on Go1.11+.

To point Athens to a different version of Go set the following environment variable
```
GO_BINARY_PATH=go1.11.X
# or whichever binary you want to use with athens
```

# Run the Proxy
If you're inside GOPATH, make sure `GO111MODULE=on`, if you're outside GOPATH, then Go Modules are on by default.
The main package is inside `cmd/proxy` and is run like any go project as follows: 

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

If you're not familiar with Docker, that's ok. We've tried to make it easy to get up and running:

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

To get started with developing the docs we provide a docker image, which runs [Hugo](https://gohugo.io/) to render the docs. Using the docker image, we mount the `/docs` directory into the container. To get it up and running, from the project root run:

```
make docs
docker run -it --rm \
        --name hugo-server \
        -p 1313:1313 \
        -v ${PWD}/docs:/src:cached \
        gomods/hugo
```

Then open [http://localhost:1313](http://localhost:1313/).

# Linting

In our CI/CD pass, we use golint, so feel free to install and run it locally beforehand:

```
go get golang.org/x/lint/golint
```
