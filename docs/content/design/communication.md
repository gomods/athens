---
title: "Communication"
date: 2018-02-11T15:57:56-05:00
---

## Communication flow

This is the story of a time long ago. A time of myth and legend when the ancient Gods were petty and cruel and they plagued build with irreproducibility.
Only one project dared to challenge their power...Athens. Athens possessed a strength the world had never seen. Filling its storage.

### Clean plate

At the beginning, there's theoretical state when a storage of the proxy is empty.

When User makes a request at this ancient time, it works as described on the flow below.

- User runs `go get` to acquire new module.
- Go CLI contacts the proxy asking for module M, version v1.0
- The proxy checks whether or not it has this module in its storage. It does not.
- The proxy downloads code from the underlying VCS and converts it into the Go Module format.
- After it receives all the bits, it stores it into its own storage and serves it to the User.
- User receives module and is happy.

The process from the user using `go get` all the way to the user downloading a module is synchronous.

![Communication flow for clear state](/athens-clear-scenario.png)

### Happy path

Now that the proxy is aware of module M at version v1.0, it can serve that module immediately to the user, without fetching it from the VCS.

![Communication flow for new proxy](/athens-proxy-filled.png)
