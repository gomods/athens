---
title: "从 VCS 到用户"
date: 2018-02-11T15:56:56-05:00
---

您阅读了[代理](/design/proxy)、[通信](/design/communication/)文档，然后打开代码库并对自己说：这并不像文档描述的那么简单。

Athens 有一系列架构组件，负责处理 Go 模块从 VCS 进入存储再到用户的整个过程。如果您对这些组件如何协同工作感到困惑，请继续阅读！

从[通信](/design/communication/)中，您知道当模块没有在存储中找到时，它会从 VCS（如 github.com）下载，然后提供给用户。您还知道整个过程是同步的。但阅读源码时，您会遇到模块获取器和下载协议暂存器等组件，难以区分其功能。本文将帮助您理解整个流程。

## 组件

请参阅下图了解各组件架构：

![组件架构图](/from-vcs-to-user.png)

如图所示，存在多层包装器。您在代码中首先遇到的组件是 Storage 和 Fetcher，下面我们首先从这两个组件开始介绍。

### Storage 存储

Storage 名称即其功能描述。通过 `proxy/storage.go` 的 `GetStorage` 函数创建存储实例。

基于传入的存储类型环境变量，可创建内存、文件系统、MongoDB 等多种存储。

模块存储于此。

### Fetcher 获取器

`Fetcher` 是我们介绍的第一个组件。从名称可以推断，`Fetcher`（`pkg/module/fetcher.go`）负责从 VCS 获取源代码。

为此，需要两项要素：`go` 二进制文件和 `afero.FileSystem`，在初始化期间传递给 `Fetcher`。

```go
mf, err := module.NewGoGetFetcher(goBin, fs)
if err != nil {
    return err
}
```
_app\_proxy.go_

当请求新模块时，会调用 `Fetch` 函数。

```go
Fetch(ctx context.Context, mod, ver string) (*storage.Version, error)
```
_fetch 函数_

`Fetcher` 的工作流程如下：

- 使用注入的 `FileSystem` 创建一个临时目录
- 在临时目录中构建一个虚拟的 Go 项目，包含简单的 `main.go` 和 `go.mod`，以便使用 `go` CLI
- 调用 `go mod download -json {module}`

此命令将模块下载到存储中。下载完成后：
- `Fetch` 函数从存储中读取模块数据并返回给调用者
- `go mod` 命令在返回的 `JSON` 响应中会包含模块文件的确切路径。

### Stash 暂存

为保持组件精简和可读，我们不愿将存储功能放到 `Fetcher`中使其膨胀。对于将模块保存到存储中，我们使用 `Stasher`，这是 Stasher 的唯一职责。

我们认为保持组件小而正交很重要，所以 `Fetcher` 和 `storage.Backend` 不相互交互。相反，`Stasher` 将它们组合在一起，并协调获取代码并存储的过程。

New 方法接受 `Fetcher` 和 `storage.Backend` 以及一组包装器（稍后解释）。

```go
New(f module.Fetcher, s storage.Backend, wrappers ...Wrapper) Stasher
```
_stasher.go_

`pkg/stash/stasher.go` 中的代码并不复杂，但很重要。其主要完成两项工作：

- 调用 `Fetcher` 获取模块数据
- 使用 `storage` 存储数据

仔细阅读代码，您会注意到传递给基本 `Stasher` 实现的包装器。这些包装器添加了更高级的逻辑，有助于保持组件整洁。

新方法返回一个包装器包装后的 `Stasher` 。

```go
for _, w := range wrappers {
    st = w(st)
}
```
_stasher.go_

### Stash wrapper - Pool 池

由于下载模块是资源密集型（内存）操作，`Pool`（pkg/stash/with_pool.go）帮助我们控制并发下载数量。

它使用 N-worker 模式，启动指定数量的 worker，然后等待任务完成。Worker 完成任务后返回结果，等待下一个任务。

在这种情况下，一个任务就是对底层 `Stasher` 的 Stash 函数调用。

### Stash wrapper - SingleFlight

我们知道模块获取是资源密集型操作，我们刚刚限制了并行下载的数量。为了帮助我们节省更多资源，我们希望避免多次处理同一个模块。

`SingleFlight` 包装器（pkg/stash/with_singleflight.go）在内部使用 map 跟踪当前下载，避免重复处理。

如果任务到来且 `map[moduleVersion]` 为空，用回调通道初始化它，并在底层 `Stasher` 上开启一个 Stash 任务。

```go
s.subs[mv] = []chan error{subCh}
go s.process(ctx, mod, ver)
```

如果请求的模块已有条目，`SingleFlight` 将订阅结果：

```go
s.subs[mv] = append(s.subs[mv], subCh)
```

一旦任务完成，模块被传递至上一层 `download protocol`（或可能包装的 `stasher`）。

### Download protocol 下载协议

最外层是 `download protocol` 下载协议。

```go
dpOpts := &download.Opts{
    Storage: s,
    Stasher: st,
    Lister: lister,
}
dp := download.New(dpOpts, addons.WithPool(protocolWorkers))
```

它包含两个组件：`Storage`、`Stasher` 和一个额外的：`Lister`。

`Lister` 用于在 `List` 和 `Latest` 函数中用于在上游代理中查找可用版本。

`Storage` 在这里又出现了，之前在 `Stasher` 中用于保存。在 _Download protocol_ 中，其用于检查模块是否已存在。如果已存在，则直接从 `storage` 获取。

否则，_Download protocol_ 使用 `Stasher` 下载模块，将其存储到 `storage`，然后返回给用户。

您还可以在上面的代码片段中看到 `addons.WithPool`。这个 addon 类似于 `Stash wrapper - Pool`。它控制代理可以处理的并发请求数量。
