---
title: Proxying Go toolchain downloads
description: How to make Athens serve Go release archives (go.dev/dl) to tools like actions/setup-go.
weight: 5
---

Athens caches Go *modules*. The Go *toolchain* itself is distributed
separately, as release archives under <https://go.dev/dl>, and CI systems
download one on every fresh machine: `actions/setup-go` fetches a ~70 MB
`go1.x.y.linux-amd64.tar.gz` for each job. If you run Athens so that builds
stop depending on public egress, that download is the next thing on the hot
path.

Athens can proxy and cache those archives too. Set `GoDownloadURL` to the
upstream you want to mirror:

```toml
GoDownloadURL = "https://go.dev/dl"
```

or, as an environment variable:

```bash
export ATHENS_GO_DOWNLOAD_URL="https://go.dev/dl"
```

Athens then serves a copy of that upstream under `<athens-url>/dl/`. Point
your tooling at it; for `actions/setup-go` v6.4 or later that is either the
`go-download-base-url` input or the `GO_DOWNLOAD_BASE_URL` environment
variable:

```yaml
- uses: actions/setup-go@v7
  with:
    go-version-file: go.mod
    go-download-base-url: https://athens.example.com/dl
```

Unless `GoDownloadURL` is set, requests under `/dl/` are answered with
`422 Unprocessable Entity` and a message saying the feature is disabled,
rather than a `404` that could be mistaken for a missing Go version.

## What is served

| Request | Upstream | Cached |
|---|---|---|
| `/dl/?mode=json&include=all` | `<GoDownloadURL>/?mode=json&include=all` | in memory, `GoDownloadListingTTL` (default 2 hours) |
| `/dl/go1.27.1.linux-amd64.tar.gz` and the other `go*.tar.gz`, `.zip`, `.msi`, `.pkg` archives | `<GoDownloadURL>/<file>` | on disk, forever |
| anything else | | `404` |

The listing is what lets clients resolve a version spec such as `1.25` to
the newest `1.25.x`. It is fetched once and reused for `GoDownloadListingTTL`
seconds (`ATHENS_GO_DOWNLOAD_LISTING_TTL`, default `7200`), so a new Go
release becomes visible within that window. If refreshing an expired listing
fails, whether the upstream is unreachable or answers with an error, the
previous listing is served and the next request tries again; an old listing
is harmless compared to failing the build. An upstream that publishes no
listing at all (for example the Microsoft build of Go at
`https://aka.ms/golang/release/latest`) gets its status passed through, and
`setup-go` falls back to constructing the archive name itself.

Responses carry `Last-Modified` and `Cache-Control` headers so clients and
intermediate caches can reuse them: the listing is marked cacheable until
Athens' own next refresh is due (`Last-Modified` is when Athens fetched it),
and archives are marked `immutable` with a one-year `max-age`.

Only the archives are served. The `.sha256` and `.asc` files are not:
`go.dev/dl` answers requests for them with an HTML page rather than the file,
and the tools this endpoint exists for do not fetch them.

Archives are immutable once published, so a cache hit is never revalidated
and nothing is evicted. A `404` from the upstream is passed through and not
remembered, so a release that appears later is picked up on the next request.
Concurrent requests for an archive that is not cached yet share a single
upstream download, and a partially downloaded file is never visible as a hit.

## Where the cache lives

```toml
GoDownloadCacheDir = "/var/lib/athens/godl"
```

```bash
export ATHENS_GO_DOWNLOAD_CACHE_DIR="/var/lib/athens/godl"
```

If unset, Athens uses a directory under the OS temporary directory. Unlike
module storage, this cache does not need to be durable: a CI fleet typically
uses a handful of Go versions, so the whole working set is a few hundred
megabytes and is rebuilt from the upstream on demand after a restart. An
`emptyDir` volume is enough in Kubernetes; use a persistent volume only if
you want to avoid the re-download.

The cache is local to each Athens replica and is not shared through the
module storage backend.
