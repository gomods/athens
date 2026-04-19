---
title: 演练
description: 了解 Athens 代理和 Go 模块
---

首先，确保您已安装 [Go v1.12+](https://gophersource.com/setup/) 并且 GOPATH/bin 在您的路径上。

## 不使用 Athens 代理

让我们回顾一下，在没有 Athens 代理的情况下，Go 中的一切是什么样子：

**Bash**
```console
$ git clone https://github.com/athens-artifacts/walkthrough.git
$ cd walkthrough
$ GO111MODULE=on go run .
go: downloading github.com/athens-artifacts/samplelib v1.0.0
The 🦁 says rawr!
```

**PowerShell**
```console
$ git clone https://github.com/athens-artifacts/walkthrough.git
$ cd walkthrough
$ $env:GO111MODULE = "on"
$ go run .
go: downloading github.com/athens-artifacts/samplelib v1.0.0
The 🦁 says rawr!
```

运行此命令的最终结果是 Go 下载了包源代码并将其打包成一个模块，将其保存在 Go 模块本地存储中。

现在我们已经看到了不使用 Athens 代理的 Go 模块工作方式，让我们看看 Athens 代理如何改变工作流程和输出。

## 使用 Athens 代理

使用最简单的安装方式，让我们一步步了解如何使用 Athens 代理，并找出每一步发生了什么。

在继续之前，让我们清除 Go 模块的本地文件，这样我们就可以在没有本地填充模块的情况下看到 Athens 代理的作用：

**Bash**
```bash
sudo rm -fr $(go env GOPATH)/pkg/mod
```

**PowerShell**
```powershell
rm -recurse -force "$(go env GOPATH)\pkg\mod"
```

现在在后台进程中运行 Athens 代理：

**Bash**
```console
$ mkdir -p $(go env GOPATH)/src/github.com/gomods
$ cd $(go env GOPATH)/src/github.com/gomods
$ git clone https://github.com/gomods/athens.git
$ cd athens
$ GO111MODULE=on go run ./cmd/proxy -config_file=./config.dev.toml &
[1] 25243
INFO[0000] Starting application at 127.0.0.1:3000
```

**PowerShell**
```console
$ mkdir "$(go env GOPATH)\src\github.com\gomods"
$ cd "$(go env GOPATH)\src\github.com\gomods"
$ git clone https://github.com/gomods/athens.git
$ cd athens
$ $env:GO111MODULE = "on"
$ $env:GOPROXY = "https://proxy.golang.org"
$ Start-Process -NoNewWindow go 'run .\cmd\proxy -config_file=".\config.dev.toml"'
[1] 25243
INFO[0000] Starting application at 127.0.0.1:3000
```

Athens 代理现在在后台运行，正在监听来自 localhost (127.0.0.1) 3000 端口的请求。

由于我们没有提供任何特定配置，Athens 代理正在使用内存存储，这仅适用于短时间体验 Athens 代理，因为您会很快耗尽内存，而且存储不会在重启之间持久化。

### 使用 Docker

有关在 Docker 中运行 Athens 的更多详细信息，请参阅[安装文档](/install/using-docker)

为了使用 Docker 运行 Athens 代理，我们首先需要创建一个用于存储持久化模块的目录。在下面的示例中，新目录名为 `athens-storage`，位于我们的用户空间中（即 `$HOME`）。然后，在运行 Docker 容器时需要设置 `ATHENS_STORAGE_TYPE` 和 `ATHENS_DISK_STORAGE_ROOT` 环境变量。

**Bash**
```bash
export ATHENS_STORAGE=$HOME/athens-storage
mkdir -p $ATHENS_STORAGE
docker run -d -v $ATHENS_STORAGE:/var/lib/athens \
   -e ATHENS_DISK_STORAGE_ROOT=/var/lib/athens \
   -e ATHENS_STORAGE_TYPE=disk \
   --name athens-proxy \
   --restart always \
   -p 3000:3000 \
   gomods/athens:latest
```

**PowerShell**
```PowerShell
$env:ATHENS_STORAGE = "$(Join-Path $HOME athens-storage)"
md -Path $env:ATHENS_STORAGE
docker run -d -v "$($env:ATHENS_STORAGE):/var/lib/athens" `
   -e ATHENS_DISK_STORAGE_ROOT=/var/lib/athens `
   -e ATHENS_STORAGE_TYPE=disk `
   --name athens-proxy `
   --restart always `
   -p 3000:3000 `
   gomods/athens:latest
```

接下来，您需要启用 [Go 模块](https://github.com/golang/go/wiki/Modules) 功能并配置 Go 使用 Athens 代理！

### 使用 Athens 代理

**Bash**
```bash
export GO111MODULE=on
export GOPROXY=http://127.0.0.1:3000
```

**PowerShell**
```powershell
$env:GO111MODULE = "on"
$env:GOPROXY = "http://127.0.0.1:3000"
```

`GO111MODULE` 环境变量仅在 Go 1.11 中控制 Go 模块功能。
可能的值有：

* `on`：始终使用 Go 模块
* `auto`（默认）：仅当存在 go.mod 文件或从 GOPATH 外部运行 go 命令时才使用 Go 模块
* `off`：从不使用 Go 模块

`GOPROXY` 环境变量告诉 `go` 二进制文件，在解析包依赖项时，不要直接与版本控制系统（如 github.com）通信，而是应该与代理通信。Athens 代理实现了 [Go 下载协议](/intro/protocol)，负责列出包的可用的版本，以及提供特定版本包的 zip 文件。

现在，当您构建和运行此示例应用程序时，`go` 将通过 Athens 获取依赖项！

```console
$ cd ../walkthrough
$ go run .
go: finding github.com/athens-artifacts/samplelib v1.0.0
handler: GET /github.com/athens-artifacts/samplelib/@v/v1.0.0.info [200]
handler: GET /github.com/athens-artifacts/samplelib/@v/v1.0.0.mod [200]
go: downloading github.com/athens-artifacts/samplelib v1.0.0
handler: GET /github.com/athens-artifacts/samplelib/@v/v1.0.0.zip [200]
The 🦁 says rawr!
```

`go run .` 的输出包括尝试查找 **github.com/athens-artifacts/samplelib** 依赖项的尝试。由于代理是在后台运行的，您还应该看到 Athens 的输出，表明它正在处理该依赖项的请求。

让我们分解一下这里发生的事情：

1. 在 Go 运行我们的代码之前，它检测到我们的代码依赖于 **github.com/athens-artifacts/samplelib** 包，而该包不在 Go 模块本地存储中。
2. 此时，Go 模块功能开始发挥作用，因为我们已将其启用。
   与其查看 GOPATH 中的包，Go 读取我们的 **go.mod** 文件并看到我们想要该包的特定版本 v1.0.0。

   ```
   module github.com/athens-artifacts/walkthrough

   require github.com/athens-artifacts/samplelib v1.0.0
   ```
3. Go 首先在 Go 模块本地存储（GOPATH/pkg/mod）中检查 **github.com/athens-artifacts/samplelib@v1.0.0**。如果该版本的包已在本地存储中，则 Go 将使用它并停止查找。但由于这是我们第一次运行，本地存储为空，所以 Go 继续查找。
4. Go 因为代理设置在 GOPROXY 环境变量中，所以从我们的代理请求 **github.com/athens-artifacts/samplelib@v1.0.0**。
5. Athens 代理在其自己的存储（在本例中是内存存储）中检查该包，但找不到。所以它从 github.com 下载它，然后保存以供后续请求使用。
6. Go 下载模块 zip 并将其放入 Go 模块本地存储 GOPATH/pkg/mod 中。
7. Go 将使用该模块并构建我们的应用程序！

对 `go run .` 的后续调用将不那么冗长：

```
$ go run .
The 🦁 says rawr!
```

不会打印额外输出，因为 Go 在 Go 模块本地存储中找到了 **github.com/athens-artifacts/samplelib@v1.0.0**，不需要从 Athens 代理请求它。

最后，退出 Athens 代理。这不能直接完成，因为我们是在后台启动 Athens 代理的，因此必须通过找到它的进程 ID 并手动终止它。

**Bash**
```bash
lsof -i @localhost:3000
kill -9 <<PID>>
```

**PowerShell**
```powershell
netstat -ano | findstr :3000 (local host Port number)
taskkill /PID typeyourPIDhere /F
```

## 下一步

现在您已经看到了 Athens 的实际应用：

* 了解如何安装具有持久化存储的[共享团队 Athens](/install/shared-team-instance)。
* 探索在生产环境中运行 Athens 的最佳实践。[即将推出/需要帮助](https://github.com/gomods/athens/issues/531)
