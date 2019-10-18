# Development Guide for Athens

The proxy is written in idiomatic Go and uses standard tools. If you know Go, you'll be able to read the code and run the server.

Athens uses [Go Modules](https://golang.org/cmd/go/#hdr-Modules__module_versions__and_more) for dependency management. You will need Go [v1.11](https://golang.org/dl) or later to get started on Athens.

See our [Contributing Guide](CONTRIBUTING.md) for tips on how to submit a pull request when you are ready.

**All the instructions in this document assume that you have checked out the code to your local machine.**

If you haven't done that, please do with the below command before you proceed:

```console
$ git clone https://github.com/gomods/athens.git
```

### Go version

Athens is developed on Go 1.11+.

To point Athens to a different version of Go set the following environment variable:

```console
GO_BINARY_PATH=go1.11.X
# or whichever binary you want to use with athens
```

# Run the Proxy

We provide three ways to run the proxy on your local machine:

1. Using [Docker](https://www.docker.com/) and [Docker Compose](https://docs.docker.com/compose/) (_we suggest this one if you're getting started_)
2. Natively on your host
3. Using [Sail](https://sail.dev)

See below for instructions for each option!

## Using Docker

As we said above, we suggest that you use this approach because it simulates a more realistic Athens deployment. This technique does the following, completely inside containers:

1. Builds Athens from scratch
2. Starts up [MongoDB](https://www.mongodb.com/) and [Jaeger](https://www.jaegertracing.io/)
3. Configures Athens to use MongoDB for its storage and Jaeger for its distributed tracing
4. Runs Athens

You'll need [Docker](https://www.docker.com/) and [Docker Compose](https://docs.docker.com/compose/) installed. Once you do, run the below command to set everything up:

```console
$ make run-docker
```

You should see some output that look like this:

```console
docker-compose -p athensdockerdev up -d dev
Creating network "athensdockerdev_default" with the default driver
Creating athensdockerdev_jaeger_1 ... done
Creating athensdockerdev_minio_1  ... done
Creating athensdockerdev_mongo_1  ... done
Creating athensdockerdev_dev_1    ... done
```

After that runs, Athens should be up and running and serving on port 3000. To test it out, run this command:

```console
$ curl localhost:3000
```

... and you should see the standard Athens response:

```console
"Welcome to The Athens Proxy"
```

When you're ready to stop Athens and all its dependencies, run this command:

```console
$ make run-docker-teardown
```

## Natively on Your Host

If you're inside GOPATH, make sure `GO111MODULE=on`, if you're outside GOPATH, then Go Modules are on by default.

The main package is inside `cmd/proxy` and is run like any go project as follows:

```console
$ cd cmd/proxy
$ go build
$ ./proxy
```

After the server starts, you'll see some console output like:

```console
Starting application at 127.0.0.1:3000
```

## Using Sail

Follow instructions at [sail.dev](https://sail.dev) to setup the sail CLI. Then simply run:

```
sail run gomods/athens
```

The command will automatically clone the athens repo and give you a local URL that you can use to open an editor and development environment directly in your browser.

### Dependencies

# Services that Athens Needs

Depending on its configuration, Athens may rely on several external services (i.e. databases, etc...) to function properly. We use [Docker](http://docker.com/) images to configure and run those services. **However, Athens does not require any of these external services by default**. For example, the default storage driver is memory, but you can opt-in to using the `fs` driver. Neither would require any external service dependencies.

But if you'd like to test out Athens against a different storage backend like MongoDB, Minio, or a cloud blob storage system, this section is for you.

If you're not familiar with Docker, that's ok. We've tried to make it easy to get up and running with the below steps.

1. [Download and install docker-compose](https://docs.docker.com/compose/install/) (docker-compose is a tool for easily starting and stopping lots of services at once)
2. Run `make dev` from the root of this repository

That's it! After the `make dev` command is done, everything will be up and running and you can move on to the next step.

If you want to stop everything at any time, run `make down`.

> Note: `make dev` only runs the minimum dependencies needed for things to work. If you'd like to run all the possible dependencies, run `make alldeps`. Keep in mind, though, that `make alldeps` does not start up Athens, but **only** its dependencies.
> All the services that get started by `make alldeps` are also available in the `docker-compose.yml` file, so if you're familiar with Docker Compose, you can also start up services as you need.

# Run unit tests

There are two methods for running unit tests:

## Completely In Containers

This method uses [Docker Compose](https://docs.docker.com/compose/) to set up and run all the unit tests completely inside Docker containers.

**We highly recommend you use this approach to run unit tests on your local machine.**

It's nice because:

- You don't have to set up anything in advance or clean anything up
- It's completely isolated
- All you need is to have [Docker Compose](https://docs.docker.com/compose/) installed

To run unit tests in this manner, use this command:

```console
make test-unit-docker
```

## On the Host

This method uses Docker Compose to set up all the dependencies of the unit tests (databases, etc...), but runs the unit tests directly on your host, not in a Docker container. This is a nice approach because you can keep all the dependency services running at all times, and you can run the actual unit tests very quickly.

To run unit tests in this manner, first run this command to set up all the dependencies:

```console
make alldeps
```

Then run this to execute the unit tests themselves:

```console
make test-unit
```

And when you're done with unit tests and want to clean up all the dependencies, run this command:

```console
make dev-teardown
```

# Run End to End Tests

End to end tests ensure that the Athens server behaves as expected from the `go` CLI tool. These tests run exclusively inside Docker containers using [Docker Compose](https://docs.docker.com/compose/), so you'll have to have those dependencies installed. To run the tests, execute this command:

```console
make test-e2e-docker
```

This will create the e2e test containers, run the tests themselves, and then shut everything down.

# Build the Docs

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
