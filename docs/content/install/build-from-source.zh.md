---
title: 从源代码构建版本化的Athens二进制文件
description: 从源代码构建版本化的Athens二进制文件
weight: 1
---
您只需执行以下几个命令即可轻松实现构建Athens二进制文件：

**Bash**
```bash
git clone https://github.com/gomods/athens
cd athens
make build-ver VERSION="0.2.0"
```

**PowerShell**
```PowerShell
git clone https://github.com/gomods/athens
cd athens
$env:GO111MODULE="on"
$env:GOPROXY="https://proxy.golang.org"
$version = "0.2.0"
$date = (Get-Date).ToUniversalTime()
go build -ldflags "-X github.com/gomods/athens/pkg/build.version=$version -X github.com/gomods/athens/pkg/build.buildDate=$date" -o athens ./cmd/proxy
```

这将生成一个名为`athens`的二进制文件. 你可以通过下列命令打印版本以及构建时间:
```console
 ./athens -version
```
which should return something like:
```console
Build Details:
        Version:        0.2.0
        Date:           2018-12-13-20:51:06-UTC
```
