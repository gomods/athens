---
title: Network modes and offline dependency resolution
description: How Athens discovers module versions from upstream sources and its storage
weight: 2
---

`NetworkMode` controls version discovery through `/@v/list` and `/@latest`. Set it in the configuration file or with `ATHENS_NETWORK_MODE`.

| Network mode | `/@v/list` | `/@latest` |
| --- | --- | --- |
| `strict` (default) | Combines upstream and stored versions; unexpected upstream errors fail the request. | Uses upstream metadata; upstream errors fail the request. |
| `offline` | Returns stored tagged versions without querying upstream. | Returns cached version metadata without querying upstream; returns 404 if no version is available. |
| `fallback` | Uses upstream and storage; if upstream fails, returns stored tagged versions when available. A missing repository with no cached versions returns 404. | Uses upstream metadata; if upstream fails, returns cached metadata when available. Without a cached version, the upstream error is returned. |

For a module matching [download mode `none`](/configuration/download), both discovery endpoints use storage only, regardless of network mode. An internal repository can therefore use `sync` while public modules use `none` in the same Athens deployment.

## What Go selects from an offline cache

The `go` command resolves dependencies before compiling their packages. For a query such as `go get example.com/module@latest`, it first requests `/@v/list` and considers the tagged versions that Athens reports. Go prefers the highest release version, then the highest prerelease version, subject to exclusions and retractions.

If no listed version is suitable, Go requests `/@latest`. Athens supplies the chronologically newest cached pseudo-version, when one exists. Pseudo-versions identify specific commits. This separate fallback matters when the listed releases are excluded by the consuming project's `go.mod` or retracted by the dependency's cached `go.mod`.

When no pseudo-version is cached, `/@latest` returns metadata for the highest cached release, or the highest cached prerelease if there are no releases. If no version is available, it returns 404. Version strings and stored metadata are preserved, including `+incompatible` suffixes.

For example, suppose Athens stores:

- Release `v1.0.0`.
- Commit `v1.0.1-0.20261002000000-abcdef123456`.

A normal Go `@latest` query selects `v1.0.0` from `/@v/list`. If that release is excluded or retracted, Go can select the cached commit through `/@latest`. A tool requesting the HTTP `/@latest` endpoint directly receives the commit metadata, even when a release is also cached. The HTTP endpoint is the commit fallback; it is not the same operation as Go's `@latest` query.

These responses follow the [Go module proxy protocol](https://go.dev/ref/mod#goproxy-protocol): the list contains tagged versions, while the latest endpoint can supply a pseudo-version when no listed version is suitable.

## Implications for dependency resolution

Offline discovery treats available storage as the source of version information. It cannot discover newer upstream releases, newer retraction declarations, or the current head of a repository's default branch. Cached commits may come from different branches; their timestamps do not establish which commit is currently on the default branch. Adding versions to the cache can therefore change later `@latest` results.

This already applied to cached `/@v/list` responses. Supporting cached `/@latest` extends discovery to tools that call that endpoint directly and to Go queries that need a commit fallback. Previously, `offline` and `none` returned 404 for `/@latest`, and their lists included cached pseudo-versions. Those pseudo-versions now remain available through `/@latest` and requests for their exact version, rather than appearing in the tagged version list.

Requests for an explicit module version continue to use that version. Normal builds still use Go's minimal version selection over the module graph; adding cached latest discovery does not by itself update existing requirements. Commands that discover or update dependencies, such as `go get ...@latest`, use the versions available in the cache. Athens does not rewrite `go.mod`, module contents, or checksums.

For clients configured with multiple `GOPROXY` sources, a successful cache response may complete a lookup without trying a later source. A 404 or 410 may allow the client to try another proxy or `direct`, depending on the configured separators. This means a cached answer can take precedence over a newer version available elsewhere. See [Go's proxy fallback rules](https://go.dev/ref/mod#goproxy-protocol).

`fallback` mode can produce different discovery results during an upstream outage than after upstream recovers. Use `offline` or a module's download mode `none` when the cache should consistently define the available versions, and record explicit dependency versions rather than relying on repeated `@latest` discovery.

## Populating storage before going offline

Offline discovery uses Athens' existing module storage. There is no separate latest-version cache. When a requested version's `.info`, `.mod`, or `.zip` is missing, download mode `sync` fetches and stores all three artifacts. Requests to `/@v/list` and `/@latest` alone do not populate that storage, and downloading one module version does not recursively download its dependencies.

To prepare storage for an air-gapped deployment:

1. Run Athens with a persistent [storage backend](/configuration/storage), `NetworkMode = "strict"`, and `DownloadMode = "sync"`, while upstream sources are reachable.
2. From each project's directory, download its dependencies through Athens. Use a fresh local Go module cache so existing local files do not bypass the server:

   ```sh
   GOPROXY=http://localhost:3000 \
   GONOPROXY=none \
   GOMODCACHE="$(mktemp -d)" \
   go mod download all
   ```

3. Retain the same storage, or transfer its contents to the isolated deployment using the same storage backend and layout. Then configure Athens for offline discovery and downloads as shown below.

Repeat this for each project and any additional module versions that need to be available offline. The stored versions are those requested through Athens; warming a project does not fetch every upstream release of its dependencies. Disk storage can also be filled manually using [the pre-filling guide](/configuration/prefill-disk-cache).

## Preventing downloads in an air-gapped deployment

`NetworkMode` controls discovery, while `DownloadMode` controls what Athens does when a requested versioned artifact is missing. To prevent both upstream discovery and downloads of missing module versions, use:

```toml
NetworkMode = "offline"
DownloadMode = "none"
```

Populate storage with the required module versions and their dependencies before use. A cached latest response does not make missing transitive dependencies available. Missing versions still return 404 under `none`.

Configure the client's `GOPROXY` to use only the offline Athens server if it must not fall back to online sources, and ensure `GONOPROXY` does not bypass that server. Checksum verification is separate from module discovery: clients may also require cached checksums or a reachable checksum database mirror. See [the checksum database configuration](/configuration/sumdb).

Download mode `none` preserves access to cached modules. To deny access to a module, including cached versions and discovery responses, use [filter exclusions](/configuration/filter).
