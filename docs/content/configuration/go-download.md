---
title: Proxying Go toolchain downloads
description: Store verified Go release archives in the configured Athens backend.
weight: 5
---

Athens can serve Go release archives under `<athens-url>/dl/`, alongside its
module proxy. Archives and release discovery metadata use the configured
Athens storage backend. They survive restarts when that backend is durable and
are shared by replicas using the same storage and distribution identity.

Enable downloads from the official Go release distribution:

```toml
GoDownloadURL = "https://go.dev/dl"
```

The equivalent environment variable is `ATHENS_GO_DOWNLOAD_URL`. Point
`actions/setup-go` v6.4.0 or later at Athens using its `go-download-base-url`
input:

```yaml
- uses: actions/setup-go@v7
  with:
    go-version: '1.25.1'
    go-download-base-url: https://athens.example.com/dl
```

`GO_DOWNLOAD_BASE_URL` is also supported by that action. Use an explicit version
or version range; its `stable` and `oldstable` aliases require the action's
standard distribution source. With `PathPrefix` configured, include it in the
base URL, for example `https://athens.example.com/proxy/dl`.

This endpoint serves release archives used by tools such as `setup-go`. The Go
command's automatic toolchain downloads use the module proxy protocol under
`golang.org/toolchain`, which continues to use the existing module storage path.

## Storage and integrity

Filesystem (including the memory backend), S3, MinIO, GCP, Azure Blob, MongoDB,
and external storage implement the optional `storage.ToolchainStorage`
capability. External storage requires a server supporting `/toolchains/v1`;
Athens checks that capability at startup and reports an error for older servers.

Module and release storage use the same configured credentials, endpoint,
bucket or database, and applicable encryption options. Release objects live in
a reserved namespace and do not appear in module lists or catalogs. MongoDB
uses separate release collections and GridFS bodies in the configured database.

Every archive stored by Athens must match the size and SHA-256 from its upstream
JSON release listing. Publication is atomic, identical saves are idempotent, and
conflicting bytes cannot replace an existing filename within a distribution.
Truncated downloads, HTML responses and checksum mismatches never become stored
hits. Temporary disk space is used for verification before publication; it is
staging space, not a separate cache. Each archive is limited to 1 GiB and a
listing to 16 MiB.

An upstream must provide the `go.dev/dl` JSON format with checksums and sizes.
Historical entries missing a checksum or size are omitted from discovery;
metadata-free archive mirrors cannot be fetched safely by this endpoint.
Checksums protect against corrupted downloads; they are trusted metadata from
the configured distribution, not an independent signature or sumdb check.

`GoDownloadSource` (`ATHENS_GO_DOWNLOAD_SOURCE`) identifies a distribution. It
defaults to `GoDownloadURL` without its trailing slash, or `https://go.dev/dl`
when no upstream is configured. Set it explicitly when replicas use different
URLs for the same distribution, or when switching an existing mirror offline.
Different distributions must use different source identities even if their
filenames match. `GoDownloadCacheDir` and its environment variable are no longer
used; existing files in that PR's standalone cache are not automatically imported.

## Discovery and network policy

| Request | Behavior |
|---|---|
| `/dl/?mode=json&include=all` | All verifiable releases including prereleases, or committed local archives in storage-only operation |
| `/dl/?mode=json` | Latest patch in each of the two newest stable series |
| `/dl/go1.25.1.linux-amd64.tar.gz` | Serve the stored archive, then apply the miss policy |
| Other paths, including `.sha256` and `.asc` | `404` |

The upstream listing is persisted and reused for `GoDownloadListingTTL` seconds
(`ATHENS_GO_DOWNLOAD_LISTING_TTL`, default `7200`). All listing query variants
share one canonical upstream snapshot; their responses are selected separately.

- `NetworkMode = "strict"`: an unavailable listing refresh fails the request.
- `NetworkMode = "fallback"`: a failed refresh returns the committed archive
  inventory, without advertising files the mirror does not hold.
- `NetworkMode = "offline"`: listings and archive requests use storage only;
  there are no upstream downloads or redirects.

A failed refresh has a short retry backoff to avoid repeatedly hammering an
unavailable upstream. Storage failures are reported as errors, not hidden as
cache misses. Offline listings contain archive/source files usable by download
clients; an installer alone does not advertise an installable release.

To run an already populated mirror without configuring any upstream:

```toml
GoDownloadEnabled = true
GoDownloadURL = ""
GoDownloadSource = "https://go.dev/dl"
NetworkMode = "offline"
```

`GoDownloadEnabled` maps to `ATHENS_GO_DOWNLOAD_ENABLED`. Populate the desired
platforms online by requesting their archives before switching offline. An
unpopulated version/platform returns `404`. Without either `GoDownloadEnabled`
or `GoDownloadURL`, `/dl/` returns `422` with an explanation that it is disabled.

## Archive miss policy

`GoDownloadMode` (`ATHENS_GO_DOWNLOAD_MODE`) overrides the global `DownloadMode`
for release archives. When unset it inherits the global mode, including its
configured default in a download-mode file. Module-path overrides, module
filters, module validation hooks and the module checksum database do not apply
to release archives.

| Mode | Unstored archive |
|---|---|
| `sync` | Fetch, verify, save, then serve |
| `none` | Return `404`; discovery also uses only stored archives |
| `async` | Schedule a bounded background fill, return `404`; retry later |
| `redirect` | Return `307` to the release distribution without saving |
| `async_redirect` | Schedule a background fill and return `307` |

Offline network mode takes precedence over every miss policy. Stored archives
are served in every mode. Background fills within an Athens instance share concurrent work for the same
filename, use the configured worker and timeout limits, and are cancelled when
Athens shuts down.

`GoDownloadRedirectURL` (`ATHENS_GO_DOWNLOAD_REDIRECT_URL`) defaults to
`GoDownloadURL`. Set a release-compatible base URL if redirects need another
destination; the module `DownloadURL` is not reused for release paths. Redirected
bytes are handled by the downstream client; Athens verifies only bytes it saves.

## HTTP behavior

GET and HEAD are supported. Stored archives support ranges, `ETag`,
`Last-Modified` and conditional requests. A cold HEAD follows the same miss
policy as GET and can populate an archive in sync mode. Range requests operate
on stored bytes; streaming backends may read and discard the prefix before the
requested range, so ranges near the end can incur additional backend traffic.

Archive responses respect `CacheControl` and otherwise default to
`private, no-cache`. Listings use `no-cache` because the committed inventory can
change. Authentication and the configured route prefix apply through the normal
Athens router. Archives remain immutable until explicitly deleted through the
storage API; there is no automatic archive eviction.
