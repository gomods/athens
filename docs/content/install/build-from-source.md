---
title: Building a versioned Athens binary from source
description: Building a versioned Athens binary from source
weight: 1
---
You can do that easily with just few commands:

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
$version = "0.2.0"
$date = (Get-Date).ToUniversalTime()
go build -ldflags "-X github.com/gomods/athens/pkg/build.version=$version -X github.com/gomods/athens/pkg/build.buildDate=$date" -o athens ./cmd/proxy
```

This will give you a binary named `athens`. You can print the version and time information by running:
```console
 ./athens -version
```
which should return something like:
```console
Build Details:
        Version:        0.2.0
        Date:           2018-12-13-20:51:06-UTC
```
