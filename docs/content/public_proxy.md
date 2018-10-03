---
title: "Using the Experimental Public Proxy"
date: 2018-02-11T16:52:23-05:00
---

We currently have an Athens server deployed on the public internet at `https://microsoftgoproxy.azurewebsites.net`. We provide it for **_experimental use only_** for the community. This experimental tag means that:

- It might be buggy
- It doesn't have an [SLA](https://en.wikipedia.org/wiki/Service-level_agreement)
- It might have significant downtime without notice
- It might disappear completely without notice

We recommend that only advanced users use this experimental proxy, and that no users rely on it for mission-critical or production workloads.

To use it, simply set your `GOPROXY` environment variable to `https://microsoftgoproxy.azurewebsites.net` and follow the "Try it out" instructions on [the home page](/).
