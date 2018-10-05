---
title: "From VCS to the User"
date: 2018-02-11T15:56:56-05:00
---

You read about proxy, communication and then opened a codebase and thought to yourself: This is not as simple as described in the docs.

If you feel lost how modules get from VCS to the user, which components it needs to visit on its way at which point it gets stored, read on my friend.

From [Communication](./communication.md) you know, that when a module is not backed up in the storage it gets downloaded from VCS (such as github.com) and then it is served to the user. You also know that this whole process is synchronous. But when you open a code you see module fetchers, download protocol stashers and you struggle to figure out what's what and how they differ. It's complicated, but it's not once you know what's going on.

## Components

Let's start with describing all the components you will see along the way. There's no better way to get a clear picture of the connected part than with a clear picture.

![Architecture chart of components](/from-vcs-to-user.png)

As you can see there is a lot of layers and wrappers. Let's start from the innermost as it is the first thing you will see in the code.

The two innermost components are:
- Storage and
- Fetcher

### Storage
Storage is what it sounds like. Storage instance created in `proxy/storage.go`'s `GetStorage` function.
Based on storage type passed as an ENV variable it will create in-memory, filesystem, mongo... storage.
This is where modules live. Once there, always there.

### Fetcher
Fetcher is the first interesting component on our way. As we can guess from the name, Fetcher (`pkg/module/fetcher.go`) is responsible for fetching the sources from VCS.
For this, it needs two things: Go Binary and `afero.FileSystem`. These are injected as shown in a snippet below.

```go
mf, err := module.NewGoGetFetcher(goBin, fs)
if err != nil {
    return err
}
```
*_app_proxy.go_

When a request for a new module comes, Fetch function is invoked.

```go
Fetch(ctx context.Context, mod, ver string) (*storage.Version, error)
```
*_fetch function_

Then Fetcher:
- creates a temp directory using an injected FileSystem
- in this temp dir, it constructs barebone go project so `Go CLI` can be used.
- invokes `go module download {module}`

This command downloads the module into the cache directory within a created temp directory.
Once the download is completed, `Fetch` function reads the module bits from FileSystem and returns them to the caller.

### Stash
As it is important for us to keep components small and readable, we did not want to bloat Fetcher with storing functionality. For storing modules into a storage we use `Stash`er. This is the single responsibility of a simple Stasher.

`Stash`er consists of a `Fetcher` and a `storage.Backend`.

```go
New(f module.Fetcher, s storage.Backend, wrappers ...Wrapper) Stasher
```
*_stasher.go_

As you can see in `pkg/stash/stasher.go` it does not really do much:
- invokes Fetcher to get module bits
- stores the bits using a `storage`

If you read carefully you noticed wrappers passed into a basic `Stash`er implementation.
These wrappers add more advanced logic and help to keep components clean.

The new method then returns a `Stasher` which is a result of wrapping basic `Stash`er with wrappers.

```go
for _, w := range wrappers {
    st = w(st)
}
```
*_stasher.go_

### Stash wrapper - Pool
As downloading a module is resource heavy (memory) operation, `Pool` (pkg/stash/with_pool.go) helps us to control simultaneous downloads.

It uses N-worker patter which spins up the specified number of workers which then waits for a job to complete. Once they complete their job, they return the result and are ready for the next one.

A job, in this case, is a call to Stash function on a backing `Stash`er.

### Stash wrapper - SingleFligt
We know that module fetching is a resource-heavy operation and we just put a limit on a number of parallel downloads. To help us save more resources we wanted to avoid processing the same module multiple times.

SingleFlight wrapper (pkg/stash/with_singleflight.go) takes care of that.
Internally it keeps track of currently running downloads using a map.
If a job arrives and `map[moduleVersion]` is empty, it initiates it with a callback channel and invokes a job on a backing `Stasher`.

```go
s.subs[mv] = []chan error{subCh}
go s.process(ctx, mod, ver)
```

if there is an entry for the requested module, SingleFligh will subscribe for a result

```go
s.subs[mv] = append(s.subs[mv], subCh)
```

and once the job is complete, the module is served one level up.

### Download protocol
The outer most level is a download protocol.

```go
dpOpts := &download.Opts{
    Storage: s,
    Stasher: st,
    Lister: lister,
}
dp := download.New(dpOpts, addons.WithPool(protocolWorkers))
```
It contains two components we already mentioned: `Storage`, `Stasher`
and one more additional: `Lister`.

`Lister` is used in `List` and `Latest` funcs to look upstream for the list of versions available.

`Storage` is here again. We saw it in a `Stasher` before, used for saving.
In _Download protocol_ it is used to check whether or not the module is already present. If it is, it is served directly from `storage`.

Otherwise, _Download protocol_ uses `Stasher` to download module, store it into a `storage` and then it serves it back to the user.

You can also see `addons.WithPool` in a code snippet above. This addon is something similar to `Stash wrapper - Pool`. It controls the number of concurrent requests proxy can handle.