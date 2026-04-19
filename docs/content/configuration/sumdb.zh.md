---
title: 代理 checksum 数据库 API
description: 如何配置 Athens 代理 checksum 数据库 API，以及为什么可能需要这样做。
weight: 4
---

如果你运行 `go get github.com/mycompany/secret-repo@v1.0.0`，而该模块版本还不在你的 `go.sum` 文件中，Go 默认会向 `https://sum.golang.org/lookup/github.com/mycompany/secret-repo@v1.0.0` 发送请求。该请求会失败，因为 Go 工具需要校验和 ，但 `sum.golang.org` 无法访问你的私有代码。

结果是：*(1) 你的构建将失败*，*(2) 你的私有模块名称已被发送到你无法控制的公共服务器*。

>你可以在[此处](https://go.googlesource.com/proposal/+/master/design/25530-sumdb.md)阅读更多关于此 `sum.golang.org` 服务的信息。

## 代理 checksum 数据库

许多公司使用 Athens 托管私有代码，但 Athens 不仅是模块代理。它还是一个 checksum 数据库代理，这意味着你公司内的任何人都可以配置 `go 工具`将这些 checksum 请求发送到 Athens，而不是公共的 `sum.golang.org` 服务器。

如果 Athens 服务器配置了 checksum 过滤器，则可以防止这些问题。

如果你使用 Go 1.13 或更高版本运行以下命令：

```bash
$ GOPROXY=<athens-url> go build .
```

那么 `Go 工具`会自动将所有 checksum 请求发送到 `<athens-url>/sumdb/sum.golang.org`，而不是 `https://sum.golang.org`。

默认情况下，当 Athens 收到 `/sumdb/...` 请求时，它会自动将其代理到 `https://sum.golang.org`，即使它是 `sum.golang.org` 不知道也无法知道的私有模块。因此，如果使用私有模块，需要更改此默认行为。

>如果你希望 Athens _不_ 将某些模块名称发送到全局 checksum 数据库，请在 `config.toml` 的 `NoSumPatterns` 值中设置这些模块名称，或使用 `ATHENS_GONOSUM_PATTERNS` 环境变量。

以下部分将更详细地介绍 checksum 数据库如何工作、Athens 如何适配，以及这一切如何影响你的工作流。

## 如何完成所有设置

在开始之前，你需要配置 Athens 不代理某些模块 checksum 请求。如果你使用 `config.toml`，请使用以下配置：

```toml
NoSumPatterns = ["github.com/mycompany/*", "github.com/secret/*"]
```

如果你使用环境变量，请使用以下配置：

```bash
$ export ATHENS_GONOSUM_PATTERNS="github.com/mycompany/*,github.com/secret/*"
```

>你可以在这些环境变量中使用任何与 [`path.Match`](https://pkg.go.dev/path?tab=doc#Match) 兼容的通配符

使用此配置启动 Athens 后，所有以 `github.com/mycompany` 或 `github.com/secret` 开头的模块的 checksum 请求都不会被转发，Athens 将向 `go` CLI 工具返回一个错误。

这确保你的私有模块名称不会泄露到公共互联网，但你的构建仍会失败。要解决此问题，请在运行 `go` 命令的机器上设置另一个环境变量：

```bash
$ export GONOSUMDB="github.com/mycompany/*,github.com/secret/*"
```

现在，你的构建将正常工作，并且你不会将有关私有代码库的信息发送到互联网。

## 我很困惑，为什么这么难？

当 Go 工具必须下载未项目 `go.sum` 文件中的 _新_ 代码时，它会尽量从其信任的服务器获取 checksum ，并将其与实际下载的代码中的 checksum 进行比较。它这样做是为了确保 _来源_ ，即确保你刚下载的代码未被篡改。

可信的 checksum 都存储在 `sum.golang.org` 中，该服务器是集中控制的。

>只有在尝试获取 _尚不在_ 你 `go.sum` 文件中的模块版本时，才会发生这些构建失败和潜在的隐私泄露。

Athens 在尊重和使用可信 checksum 的同时，尽力确保你的私有名称不会泄露到公共服务器。某些情况下，它需要在构建失败和泄露信息之间二选一，最终它选择让构建失败。这就是为什么每个使用 Athens 服务器的人都需要设置 `GONOSUMDB` 环境变量。

我们相信，配合一份优秀的文档（我们希望本手册正是如此），我们已经在便利性和隐私之间找到了恰当的平衡点。
