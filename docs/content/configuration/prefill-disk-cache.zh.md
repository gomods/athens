---
title: 手动添加模块
description: 如何手动添加模块到 Athens 存储中
weight: 5
---

Athens 的广受欢迎的特性之一是它可以完全脱离互联网运行。然而，在这种情况下它无法访问上游（例如版本控制系统或其他模块代理）来获取存储中缺失的模块。因此，我们需要将所需的依赖项手动放到 Athens 所使用的磁盘中。

本文将指导您如何打包一个名为 `github.com/my/module` 的单一模块，并将其添加到 Athens 的存储中。

# 首先，获取工具

你需要根据模块源代码生成以下资源：

- `source.zip` - Go 源代码，打包成 zip 文件
- `go.mod` - 模块中的 `go.mod` 文件
- `$VERSION.info` - 模块的元数据

`source.zip` 文件具有特定的目录结构，`$VERSION.info` 具有 JSON 结构。这两者的格式都必须正确，Athens 才能提供 Go 工具链所接受的正确依赖格式。

>不建议您自己创建这些资源。请使用 [pacmod](https://github.com/plexsystems/pacmod) 或 [gopack](https://github.com/alex-user-go/gopack)。

## 使用 pacmod

安装 `pacmod` 工具，运行 `go get` 命令：

```console
$ go get github.com/plexsystems/pacmod@v0.4.0
```

此命令会将 `pacmod` 二进制文件安装到 `$GOPATH/bin/pacmod` 目录中，请确保该目录在您的 `$PATH` 中。

**接下来，运行 `pacmod` 创建资源**

安装 `pacmod` 后，您需要获取要打包的模块源代码。在运行命令之前，将环境中的 `VERSION` 变量设置为生成资源的模块版本。

配置示例如下：

```console
$ export VERSION="v1.0.0"
```

>注意：确保 `VERSION` 变量以 `v` 开头

接下来，导航到模块源代码的顶级目录，运行 `pacmod`：

```console
$ pacmod pack github.com/my/module $VERSION .
```

命令完成后，您会注意到在与运行命令相同的目录中创建了三个新文件：

- `go.mod`
- `$VERSION.info`
- `$VERSION.zip`

## 使用 gopack

>使用此方法需要安装 docker-compose。

Fork gopack 项目并克隆到本地机器（或仅将文件下载到您的计算机）

编辑 <code>goget.sh</code>，添加您要下载的 Go 模块列表：

```bash
#!/bin/bash
go get github.com/my/module1;
go get github.com/my/module2;
```

运行：

```bash
docker-compose up --abort-on-container-exit
```

命令完成后，您会在 ATHENS_STORAGE 文件夹中看到所有模块，可以移动到 Athens 磁盘存储中。

# 接下来，将资源移动到 Athens 存储目录

现在您已经构建了资源，需要将它们移动到 Athens 磁盘存储的位置。下面的命令假设 `$STORAGE_ROOT` 是指向 Athens 用于磁盘存储的顶级目录的环境变量。

>如果您使用 `$ATHENS_DISK_STORAGE_ROOT` 环境变量配置了 Athens，则Athens 存储位置的根目录就是这个环境变量的值。使用 `export STORAGE_ROOT=$ATHENS_DISK_STORAGE_ROOT` 来为下面的命令准备环境。

首先创建要将资源移动目标目录：

```console
$ mkdir -p $STORAGE_ROOT/github.com/my/module/$VERSION
```

最后，确保您仍在模块源代码仓库根目录中（与运行 `pacmod` 命令时相同），并将三个新文件移动到您刚创建的新目录中：

```console
$ mv go.mod $STORAGE_ROOT/github.com/my/module/$VERSION/go.mod
$ mv $VERSION.info $STORAGE_ROOT/github.com/my/module/$VERSION/$VERSION.info
$ mv $VERSION.zip $STORAGE_ROOT/github.com/my/module/$VERSION/source.zip
```

>请注意，这里更改了 `.zip` 文件的名称。

# 最后，测试您的设置

此时，Athens 服务器的磁盘缓存应该已存储了 `github.com/my/module` 模块的 `$VERSION` 版本。下次请求此模块时，Athens 会在其磁盘存储中找到它，而不会尝试从上游代理获取。

可以通过运行以下 `curl` 命令快速测试，假设 Athens 服务器运行在 `http://localhost:3000` 上，并且已配置为存储了对应包的磁盘。

```console
$ curl localhost:3000/github.com/my/module/@v/$VERSION.info
```

运行此命令时，Athens 应该立即返回，而不会连接到任何其他网络服务。
