---
title: Verifying stored modules
description: Detect and purge stored module zips that disagree with the Go checksum database (issue #2145).
weight: 6
---

Athens builds module zips itself on a cache miss and stores them. If Athens ever
ran with a Go toolchain that built a zip incorrectly, that wrong zip stays in
storage and is re-served on every cache hit — a toolchain upgrade alone does not
fix it, because a cache hit never rebuilds the module.

`athens-proxy -verify-storage` finds stored module versions whose zip bytes
disagree with the Go **checksum database** (`sum.golang.org` by default) and,
with `-purge`, deletes them so the next request rebuilds them correctly.

## Am I affected?

You only need this if **all three** are true:

1. You ran **Athens `v0.12.0`–`v0.15.1`** (these bundle Go 1.20.x). `v0.15.2`+
   bundle Go ≥ 1.22 and are unaffected:

   | Bundled Go | Athens releases | Affected? |
   | --- | --- | --- |
   | 1.20.x | `v0.12.0` … `v0.15.1` | **Yes** |
   | 1.22 | `v0.15.2` … `v0.15.4` | No |
   | 1.23.5 | `v0.16.0` … `v0.16.1` | No |
   | 1.25.1 | `v0.16.2` | No |
   | 1.26.2 | `v0.17.0` … (current) | No |

2. That instance was **still running in ~2025 or later** — the bug only produces
   a mismatch for modules that declare `go >= 1.24`, which do not exist before the
   Go 1.24 release (Feb 2025). An instance retired or upgraded to `v0.15.2`+
   before then holds no poison.

3. You proxied **public** modules that declare `go >= 1.24` and ship files under a
   `vendor/` directory nested below the module root (or a root
   `vendor/modules.txt`) — e.g. `github.com/onsi/ginkgo/v2@v2.32.0`.

Symptom your users saw: a `checksum mismatch` (or unexpected `404`) for specific
module versions that download and verify fine directly from `proxy.golang.org`.

### What your users must do

- **Clients using the default checksum database** were never silently poisoned;
  they just got errors. Once you purge (so Athens serves canonical bytes) they
  recover automatically on the next `go` command. No action needed.
- **Clients that disabled checksum verification** (`GONOSUMDB` / `GOSUMDB=off` /
  `GOPRIVATE`) for the affected module may have committed the wrong hash into
  their `go.sum`. After the fix they must refresh it: delete the stale line and
  run `go mod tidy` (or `GOFLAGS=-mod=mod`).

## Running it

`-verify-storage` runs as a one-shot process using the same image and config as
your deployment, and **exits without starting the server**. It is safe to run
against a live deployment: it is a separate process, and it only ever deletes
versions it can prove are wrong (and therefore re-fetchable).

Report only (deletes nothing):

```bash
athens-proxy -verify-storage -config_file=/config/config.toml
```

Verify and delete the mismatches:

```bash
athens-proxy -verify-storage -purge -config_file=/config/config.toml
```

`-purge` on its own is rejected — it only ever acts on the mismatches found by
the verify pass, never as a general cache purge.

### As a Kubernetes Job

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: athens-verify-storage
spec:
  template:
    spec:
      restartPolicy: Never
      containers:
        - name: verify
          image: gomods/athens:latest # same image as your Deployment
          args: ["-verify-storage", "-config_file=/config/config.toml"]
          envFrom:
            - secretRef:
                name: athens-storage-credentials # same storage config as the Deployment
          volumeMounts:
            - { name: config, mountPath: /config }
      volumes:
        - { name: config, configMap: { name: athens-config } }
```

Add `-purge` to `args` once you have reviewed the report. For the **disk**
storage driver, run it where the storage volume is mounted (mount the same PVC in
the Job, or `kubectl exec` into the running pod).

## Prerequisites and limitations

- **Checksum-DB access.** The sweep needs outbound access to your configured
  `SumDBs` (default `https://sum.golang.org`). Air-gapped with no such access →
  it verifies nothing and safely does nothing.
- **Private modules are skipped.** Anything matching `NoSumPatterns`
  (`ATHENS_GONOSUM_PATTERNS`) is skipped *before* any lookup — both because there
  is no public checksum to compare against and to avoid leaking private module
  names. Private modules therefore cannot be auto-verified; if you suspect one is
  poisoned, delete that version manually so it refetches.
- **Prefer this over emptying storage.** Emptying storage also destroys private
  modules whose upstream may no longer exist — Athens is often their last copy.
  The targeted sweep only removes what it can prove wrong *and* replaceable.
