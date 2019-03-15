---
title: "下载协议"
date: 2018-02-11T16:58:56-05:00
weight: 3
---

Athens 在 Go 命令行接口的基础上建立了一些端点，用来于外部提供模块的代理通信。我们称这些端点为_下载协议_

vgo 在下载协议上的原始调研报告可以在这里找到：https://research.swtch.com/vgo-module

每个端点都对应一个顶层模块。让我们假设模块 `htp` 是由 `acidburn` 编写的。

因此，我们下面提到的端点都假设位于 `acidburn/htp/@v/{endpoint}`（例如：`acidburn/htp/@v/list`）

在下面的例子中，`$HOST` 和 `$PORT` 都是 Athens 服务的主机和端口的占位符。

## 版本列表

这个端点返回 Athens 中模块 `acidburn/htp` 的版本列表。下面的列表由换行符分割：

```HTTP
GET $HOST:$PORT/github.com/acidburn/htp/@v/list
```

```HTML
v0.1.0
v0.1.1
v1.0.0
v1.0.1
v1.2.0
```

## 版本信息

```HTTP
GET $HOST:$PORT/github.com/acidburn/htp/@v/v1.0.0.info
```

这会以 JSON 格式返回关于 v1.0.0 的信息。它看起来像：

```json
{
    "Name": "v1.0.0",
    "Short": "v1.0.0",
    "Version": "v1.0.0",
    "Time": "1972-07-18T12:34:56Z"
}
```

## 文件 Go.mod

```HTTP
GET $HOST:$PORT/github.com/acidburn/htp/@v/v1.0.0.mod
```

这会返回文件 go.mod 的版本 v1.0.0.如果 $HOST:$PORT/github.com/acidburn/htp 的 `v1.0.0` 版本没有依赖，
那么响应就会像这样：

```
module github.com/acidburn/htp
```

## 模块源

```HTTP
GET $HOST:$PORT/github.com/acidburn/htp/@v/v1.0.0.zip
```

显而易见——它会把该模块的 v1.0.0 版本的源码以 zip 格式返回。

## Latest

```HTTP
GET $HOST:$PORT/github.com/acidburn/htp/@latest
```

这个端点会返回对应模块的最新版本。如果没有 latest 标签，它会根据最后一次提交的哈希值去找到对应的版本。
