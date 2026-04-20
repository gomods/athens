---
title: 立即体验！
description: 如何查看 Athens 的实际运行
---

## 立即体验 Athens

要快速查看 Athens 的实际运行，请按照以下步骤操作：

首先，确保您已安装 [Go v1.12+](https://gophersource.com/setup/)，GOPATH/bin 在您的路径上，并且已启用 [Go 模块](https://github.com/golang/go/wiki/Modules) 功能。

**Bash**
```bash
export GO111MODULE=on
```

**PowerShell**
```powershell
$env:GO111MODULE = "on"
```

接下来，使用 git 和 Go 在后台进程中安装并运行 Athens 代理。

```console
$ git clone https://github.com/gomods/athens
$ cd athens/cmd/proxy
$ go install
$ proxy &
[1] 37186
INFO[0000] Exporter not specified. Traces won't be exported
INFO[0000] Starting application at http://127.0.0.1:3000
```

接下来，您需要配置 Go 使用 Athens 代理！

**Bash**
```bash
export GOPROXY=http://127.0.0.1:3000
```

**PowerShell**
```powershell
$env:GOPROXY = "http://127.0.0.1:3000"
```

现在，当您构建和运行此示例应用程序时，**go** 将通过 Athens 获取依赖项！

```console
$ git clone https://github.com/athens-artifacts/walkthrough.git
$ cd walkthrough
$ go run .
go: finding github.com/athens-artifacts/samplelib v1.0.0
handler: GET /github.com/athens-artifacts/samplelib/@v/v1.0.0.info [200]
handler: GET /github.com/athens-artifacts/samplelib/@v/v1.0.0.mod [200]
go: downloading github.com/athens-artifacts/samplelib v1.0.0
handler: GET /github.com/athens-artifacts/samplelib/@v/v1.0.0.zip [200]
The 🦁 says rawr!
```

`go run .` 的输出包括尝试查找 **github.com/athens-artifacts/samplelib** 依赖项的尝试。由于代理是在后台运行的，您还应该看到 Athens 的输出，表明它正在处理该依赖项的请求。

您应该了解了使用 Athens 的感觉！
