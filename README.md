# Welcome To Athens, Gophers!

![Athens Banner](./docs/static/banner.png)

[![Build Status](https://cloud.drone.io/api/badges/gomods/athens/status.svg)](https://cloud.drone.io/gomods/athens)
[![Build Status](https://dev.azure.com/gomods/athens/_apis/build/status/gomods.athens?branchName=master)](https://dev.azure.com/gomods/athens/_build/latest?definitionId=1&branchName=master)
[![GoDoc](https://godoc.org/github.com/gomods/athens?status.svg)](https://godoc.org/github.com/gomods/athens)
[![Go Report Card](https://goreportcard.com/badge/github.com/gomods/athens)](https://goreportcard.com/report/github.com/gomods/athens)
[![codecov](https://codecov.io/gh/gomods/athens/branch/master/graph/badge.svg)](https://codecov.io/gh/gomods/athens)
[![Docker Pulls](https://img.shields.io/docker/pulls/gomods/athens.svg?maxAge=604800)](https://hub.docker.com/r/gomods/athens/)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](http://makeapullrequest.com)
[![Open Source Helpers](https://www.codetriage.com/gomods/athens/badges/users.svg)](https://www.codetriage.com/gomods/athens)

Welcome to the Athens project! We are a proxy server for the [Go Modules download API](https://docs.gomods.io/intro/protocol/).

|<img src="docs/static/meeting-icon.svg" alt="Developer Meetings" width="20" height="20" />Developer Meetings|
|------------------|
|We hold weekly developer meetings every Thursday. To join them, watch previous meeting recordings or find more information, please see [the docs](https://docs.gomods.io/contributing/community/developer-meetings/). Absolutely everyone is invited to attend these, suggest topics, and participate!|
</br>

See our documentation site [https://docs.gomods.io](https://docs.gomods.io) for more details on the project.

# Project Status

Project Athens is in beta. It's being tested and used on teams and has an experimental deployment for public use.

We encourage you to [try it out](https://docs.gomods.io/install/), let us know what you think, and [contribute](#contributing) when you can!

# More Details Please!

The proxy implements the [Go modules download protocol](https://docs.gomods.io/intro/protocol/).

There is currently an experimental public proxy, and we have plans to host a more stable public proxy with more guarantees. We also have a community of folks using Athens inside their organizations as an internal proxy. In either deployment, developers set their `GOPROXY` environment variable to point to the Athens proxy of their choice. At that point, `go get`, `go build`, and `go build`s will use the proxy to download dependencies as necessary.

Athens proxies are highly configurable, so they can work for lots of different deployments. For example, proxies support a wide variety of storage drivers including:

- Cloud blob storage services
- MongoDB
- Content distribution networks (CDNs)
- Shared disk
- In-memory

# Development

See [DEVELOPMENT.md](./DEVELOPMENT.md) for details on how to set up your development environment and start contributing code.

Speaking of contributing, read on!

# Contributing

This project has plenty of interesting and challenging work to do and a wonderful community.

If you find a bug or want to fix a bug, we :heart: PRs and issues! If you see an issue
in the [queue](https://github.com/gomods/athens/issues) that you'd like to work on, please just post a comment saying that you want to work on it. Something like "I want to work on this" is fine.

If you decide to contribute (we hope you do :smile:), the process is familiar and easy if you've used Github before. There are no long documents to read or complex setup. If you haven't used Github before, the awesome [@bketelsen](https://github.com/bketelsen) has created a good overview on how to contribute code - see [here](https://www.youtube.com/watch?v=bgSDcTyysRc).

Before you do start getting involved or contributing, we want to let you know that we follow a general [philosophy](./PHILOSOPHY.md) in how we work together, and we'd really appreciate you getting familiar with it before you start.

It's not too long and it's ok for you to "skim" it (or even just read the first two sections :smile:), just as long as you understand the spirit of who we are and how we work.

# Getting Involved

If you're not ready to contribute code yet, there are plenty of other great ways to get involved:

- Come talk to us in the `#athens` channel in the [Gophers slack](http://gophers.slack.com/). We’re a really friendly group, so come say hi and join us! Ping me (`@arschles` on slack) in the channel and I’ll give you the lowdown
- Come to our [weekly development meetings](https://docs.google.com/document/d/1xpvgmR1Fq4iy1j975Tb4H_XjeXUQUOAvn0FximUzvIk/edit#)! They are a great way to meet folks, ask questions, find some stuff to work on, or just hang out if you want to. Just like with this project, absolutely everyone is welcome to join and participate in those
- Get familiar with the system. There's lots to read about. Here are some places to start:
    - [Gentle Introduction to the Project](https://medium.com/@arschles/project-athens-c80606497ce1) - the basics of why we started this project
    - [The Download Protocol](https://medium.com/@arschles/project-athens-the-download-protocol-2b346926a818) - the core API that the proxy implements and the `go` CLI uses to download packages
    - [Proxy Design](https://docs.gomods.io/design/proxy/) - what the proxy is and how it works
    - [Go modules wiki](https://github.com/golang/go/wiki/Modules) - context and details on how Go dependency management works in general
    - ["Go and Versioning"](https://research.swtch.com/vgo) - long articles on Go dependency management details, internals, etc...

# Built on the Shoulders of Giants

The Athens project would not be possible without the amazing projects it builds on. Please see
[SHOULDERS.md](./SHOULDERS.md) to see a list of them.

# Coding Guidelines

We all strive to write nice and readable code which can be understood by every person of the team. To achieve that we follow principles described in Brian's talk `Code like the Go team`.

- [Printed version](https://learn-golang.com/en/goteam/)
- [Gophercon RU talk](https://www.youtube.com/watch?v=MzTcsI6tn-0)

# Code of Conduct

This project follows the [Contributor Covenant](https://www.contributor-covenant.org/) (English version [here](./CODE_OF_CONDUCT.md)) code of conduct.

If you have concerns, notice a code of conduct violation, or otherwise would like to talk about something
related to this code of conduct, please reach out `@arschles` on the [Gophers Slack](https://gophers.slack.com/).

---
Athens banner attributed to Golda Manuel
