---
title: Network mode
description: Configures how Athens will return the results of the /list endpoint
weight: 6
---

The NetworkMode configuration in Athens determines how results are returned by the `/list` endpoint. This endpoint can retrieve information from both Athens' own [storage](/configuration/storage) and the upstream version control system (VCS).

> Note: The NetworkMode configuration can also affect the behavior of other endpoints by improving error messaging

## Network mode keywords

There are 3 modes available for the NetworkMode configuration. To configure the `NetworkMode` settings at `config.dev.toml`, set the `NetworkMode` to 1 of the 3 available modes:

1. `strict`: In this mode, Athens will merge versions from the VCS and storage but will fail if either of them fails. This mode provides the most consistent results.

2. `offline`: This mode only retrieves versions from Athens' storage and never reaches out to the VCS.

3. `fallback`: This mode retrieves versions from Athens' storage only if the VCS fails. Note that using this mode may result in inconsistent results since fallback mode does its best to provide the available versions at the time of the request.

## Use cases

### Ensuring consistency in module version retrieval

When working in environments that prioritize consistency and reliability, configuring Athens with the `strict` mode guarantees a dependable and predictable module version resolution mechanism.

Using the `strict` mode, Athens merges module versions from its storage and the VCS, while ensuring that any failure in either source results in a failure response.

Choose Athens' `strict` network mode for a reliable and consistent approach to module version retrieval.

### Fetching modules for offline environments

When working in offline environments with private networks lacking direct internet access, Athens' `offline` mode can be useful.

For example, you can [pre-download](/configuration/prefill-disk-cache/) modules using Athens from a machine with internet access. Subsequently, the pre-downloaded modules are accessible within the offline network through Athens, facilitating development and builds without requiring an active internet connection.

### Ensuring availability of modules versions

In certain situations, the stability or availability of the VCS may vary. To mitigate potential disruptions, Athens offers the `fallback` mode. If the VCS encounters issues or fails to respond, Athens falls back to serving modules from its storage to ensure continued availability.

However, using the `fallback` mode may result in inconsistent results. Athens will provide the available versions at the time of the request, which may differ from the latest versions in the VCS.
