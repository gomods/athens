---
title: 代理校验和数据库 API
description: 如何配置 Athens 代理校验和数据库 API，以及为什么可能需要这样做。
weight: 4
---

如果您运行 `go get github.com/mycompany/secret-repo@v1.0.0`，而该模块版本尚不在您的 `go.sum` 文件中，Go 默认会向 `https://sum.golang.org/lookup/github.com/mycompany/secret-repo@v1.0.0` 发送请求。该请求将失败，因为 Go 工具需要校验和，但 `sum.golang.org` 无法访问您的私有代码。

结果是：*(1) 您的构建将失败*，*(2) 您的私有模块名称已被发送到您无法控制的公共服务器*。

>您可以在[此处](https://go.googlesource.com/proposal/+/master/design/25530-sumdb.md)阅读更多关于此 `sum.golang.org` 服务的信息。

## 代理校验和数据库

许多公司使用 Athens 托管私有代码，但 Athens 不仅是模块代理。它还是一个校验和数据库代理。这意味着您公司内的任何人都可以配置 `go` 将这些校验和请求发送到 Athens，而不是公共的 `sum.golang.org` 服务器。

如果 Athens 服务器配置了校验和过滤器，则可以防止这些问题。

如果您使用 Go 1.13 或更高版本运行以下命令：

```bash
$ GOPROXY=<athens-url> go build .
```

... 那么 Go 工具会自动将所有校验和请求发送到 `<athens-url>/sumdb/sum.golang.org`，而不是 `https://sum.golang.org`。

默认情况下，当 Athens 收到 `/sumdb/...` 请求时，它会自动将其代理到 `https://sum.golang.org`，即使它是 `sum.golang.org` 不知道也无法知道的私有模块。因此，如果您使用私有模块，您会希望更改默认行为。

>如果您希望 Athens _不_ 将某些模块名称发送到全局校验和数据库，请在 `config.toml` 的 `NoSumPatterns` 值中设置这些模块名称，或使用 `ATHENS_GONOSUM_PATTERNS` 环境变量。

以下部分将更详细地介绍校验和数据库如何工作、Athens 如何适应，以及这一切如何影响您的工作流程。

## 如何完成所有设置

在开始之前，您需要使用告诉它不代理某些模块的配置值来运行 Athens。如果您使用 `config.toml`，请使用以下配置：

```toml
NoSumPatterns = ["github.com/mycompany/*", "github.com/secret/*"]
```

如果您使用环境变量，请使用以下配置：

```bash
$ export ATHENS_GONOSUM_PATTERNS="github.com/mycompany/*,github.com/secret/*"
```

>您可以在这些环境变量中使用任何与 [`path.Match`](https://pkg.go.dev/path?tab=doc#Match) 兼容的字符串

使用此配置启动 Athens 后，所有以 `github.com/mycompany` 或 `github.com/secret` 开头的模块的校验和请求都不会被转发，Athens 将向 `go` CLI 工具返回一个错误。

此行为将确保您的私有模块名称不会泄露到公共互联网，但您的构建仍会失败。要解决此问题，请在运行 `go` 命令的机器上设置另一个环境变量：

```bash
$ export GONOSUMDB="github.com/mycompany/*,github.com/secret/*"
```

现在，您的构建将正常工作，并且您不会将有关私有代码库的信息发送到互联网。

## 我很困惑，为什么这很难？

当 Go 工具必须下载不在项目 `go.sum` 文件中的_新_代码时，它会尽力从其信任的服务器获取校验和，并将其与实际下载的代码中的校验和进行比较。它这样做是为了确保_来源_，即确保您刚下载的代码未被篡改。

可信的校验和都存储在 `sum.golang.org` 中，该服务器是集中控制的。

>只有在尝试获取_尚不在_您 `go.sum` 文件中的模块版本时，才会发生这些构建失败和潜在的隐私泄露。

Athens 在尊重和使用可信校验和的同时，尽力确保您的私有名称不会泄露到公共服务器。在某些情况下，它必须选择是让您的构建失败还是泄露信息，因此它选择让构建失败。这就是为什么每个使用该 Athens 服务器的人都需要设置 `GONOSUMDB` 环境变量。

我们相信，凭借良好的文档——我们希望这就是！——我们已经在便利性和隐私性之间取得了正确的平衡。
