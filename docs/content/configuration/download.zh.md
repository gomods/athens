---
title: 下载模式文件
description: 当模块不在存储中的处理方式
weight: 1
---

Athens 接受一个 [HCL](https://github.com/hashicorp/hcl) 格式的配置文件，该文件定义了当某个模块的某个版本在本地存储中找不到时，Athens 应如何响应。通过此功能，Athens 可以灵活配置以满足各类组织需求。该文件最常见的用途包括：

- 配置 Athens 永不下载或代理某个模块或一组模块
- 将某个模块或一组模块重定向到其他代理

本文档概述如何使用此`下载模式文件`来完成这些任务及其他更多功能。

>请参阅下面的"用例"部分，了解更多关于如何启用这些行为及其他功能的详细信息。

## 配置

创建`下载模式配置文件`后，通过在 `config.toml` 文件中设置 `DownloadMode` 配置参数，或设置 `ATHENS_DOWNLOAD_MODE` 环境变量，告诉 Athens 使用该文件。您可以将此配置值设置为以下两个值之一：

1. 设置为 `file:$FILE_PATH`，其中 `$FILE_PATH` 是 HCL 文件的路径
2. 设置为 `custom:$BASE_64`，其中 `$BASE_64` 是 base64 编码的 HCL 文件

>除了上述两个值外，您也可以将此配置设置为 `sync`、`async`、`none`、`redirect` 或 `async_redirect`。这样下载模式将作为全局设置，而非针对特定的子模块。参见下方了解每个值的含义。

## 关键字

当 Athens 收到对模块 `github.com/pkg/errors` 版本 `v0.8.1` 的请求，而其存储中没有该模块该版本时，它会查阅`下载模式文件`获取具体操作指令：

1. **`sync`**：通过 `go mod download` 从 VCS 同步下载模块，将其保存到 Athens 存储，然后立即返回给用户，这是默认行为。
2. **`async`**：向客户端返回 404，并异步下载 module@version 到 Athens 存储。
3. **`none`**：返回 404，不执行任何操作。
4. **`redirect`**：重定向到上游代理（如 proxy.golang.org），之后不执行任何操作。
5. **`async_redirect`**：重定向到上游代理，同时异步下载 module@version 到 Athens 存储。

Athens 期望这些关键字与模块通配符使用（例如 `github.com/pkg/*`）。您可以将关键字和通配符组合起来，为特定一组模块指定行为。

>Athens 使用 Go 的 [path.Match](https://golang.org/pkg/path/#Match) 函数来解析模块通配符。

下面是下载模式文件的一个示例：

```javascript
downloadURL = "https://proxy.golang.org"

mode = "async_redirect"

download "github.com/gomods/*" {
    mode = "sync"
}

download "golang.org/x/*" {
    mode = "none"
}

download "github.com/pkg/*" {
    mode = "redirect"
    downloadURL = "https://gocenter.io"
}
```

前两行描述了针对所有模块的 _默认_ 行为。此行为在下面的特定一组模块中被覆盖。本例中默认行为为：

- 立即将所有请求重定向到 `https://proxy.golang.org`
- 在后台从版本控制系统（VCS）下载模块并存储

文件的其余部分包含 `download` 块，用于覆盖特定一组模块的默认行为。

第一个块指定匹配 `github.com/gomods/*` 的模块（如 `github.com/gomods/athens`）将从 GitHub 下载、保存，然后返回给用户。

第二个块指定匹配 `golang.org/x/*` 的模块（如 `golang.org/x/text`）始终返回 HTTP 404 响应码。此行为确保 Athens 永远不会存储或提供任何以 `golang.org/x` 开头的模块。

如果用户的 `GOPROXY` 环境变量设置为逗号分隔的列表，`go` 命令行工具将持续尝试列表中的下一个选项。例如，如果用户将 `GOPROXY` 设置为 `https://athens.azurefd.net,direct`，然后运行 `go get golang.org/x/text`，他们仍然会下载 `golang.org/x/text` 到本地，只是该模块不会来自 Athens。

最后一个块指定匹配 `github.com/pkg/*` 的模块（如 `github.com/pkg/errors`）始终将 `go` 工具重定向到 https://gocenter.io。这种情况下，Athens 永远不会将给定模块保存到存储中。

## 用例

下载模式文件用途广泛，允许您以多种方式配置 Athens。以下是一些常见用法。

## 阻止某些模块

如果运行 Athens 为 Go 开发团队服务，您可能希望确保团队不使用特定一组模块（例如，由于许可或安全问题）。

这种情况下，您可以在文件中写入以下内容：

```hcl
download "bad/module/repo/*" {
    mode = "none"
}
```

### 防止存储溢出

如果您使用空间有限的 [存储后端](/configuration/storage) 运行 Athens，可能希望阻止 Athens 存储占用大量空间的某些模块。为了避免耗尽存储，同时确保用户仍能访问这些模块，可以使用 `redirect` 指令：

```hcl
download "very/large/*" {
    mode = "redirect"
    url = "https://reliable.proxy.com"
}
```

>使用 `redirect` 模式时，请确保指定的`url`指向一个可靠的代理。
