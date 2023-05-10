---
title: Network mode
description: Configures how Athens will return the results of the /list endpoint
weight: 6
---

The NetworkMode configuration in Athens determines how results are returned by the `/list` endpoint. This endpoint can retrieve information from both Athens' own [storage](/configuration/storage) and the upstream version control system (VCS).

> Note: The NetworkMode configuration can also affect the behavior of other endpoints by improving error messaging

## Network mode keywords

There are 3 modes available for the NetworkMode configuration. To configure the `NetworkMode` settings at `config.dev.toml`, set the `NetworkMode` to 1 of the 3 available modes:

1. `strict`: In this mode, Athens will merge versions from the VCS and storage, but will fail if either of them fails. This mode provides the most consistent results.

2. `offline`: This mode only retrieves versions from Athens' storage and never reaches out to the VCS.

3. `fallback`: This mode retrieves versions from Athens' storage only if the VCS fails. Note that using this mode may result in inconsistent results since fallback mode does its best to provide the available versions at the time of the request.
