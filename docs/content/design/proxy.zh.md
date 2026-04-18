---
title: "代理"
date: 2018-02-11T15:59:56-05:00
---

## Athens 代理

Athens 代理有两个主要用途：

- 内部部署
- 公共镜像部署

本文档详细介绍了 Athens 代理的功能，您可以使用这些功能来实现任一用途。

## Athens 代理的角色

将代理主要部署在企业内部可以：

- 托管私有模块
- 限制对公共模块的访问
- 存储公共模块

重要的是，Athens 代理并不打算作为上游代理的完整镜像。对于公共模块，其角色是在本地存储模块并提供访问控制。

## 当公共模块未存储时会发生什么

当用户从代理请求模块 `MxV1`，而 Athens 代理的存储中没有 `MxV1` 时，它首先确定 `MxV1` 是私有模块还是非私有模块。

如果是私有模块，它立即从内部 VCS 将模块存储到代理存储中。

如果不是私有模块，Athens 代理会查询其排除列表以获取非私有模块（见下文）。如果 `MxV1` 在排除列表上，Athens 代理返回 404 并且终止其他流程。如果 `MxV1` 不在排除列表上，Athens 代理执行以下算法：

```
upstreamDetails := lookUpstream(MxV1)
if upstreamDetails == nil {
	return 404 // if the upstream doesn't have the thing, just bail out
}
return upstreamDetails.baseURL
```

这个算法的重要部分是 `lookUpstream`。该函数查询上游代理上的一个端点：

- 如果在其存储中没有 `MxV1`，则返回 404
- 如果其存储中有 `MxV1`，则返回 MxV1 的 base URL

_在项目的更高版本中，我们可能会在代理上实现事件流，任何其他代理都可以订阅并监听其关心的模块的删除/弃用信息_

## 排除列表和私有模块过滤器

为了适应私有（企业）部署，Athens 代理维护两个重要的访问控制机制：

- 私有模块过滤器
- 公共模块排除列表

### 私有模块过滤器

私有模块过滤器是字符串通配符模式，告诉 Athens 代理什么是私有模块。例如，字符串 `github.internal.com/**` 告诉 Athens 代理：

- 永远不要向公共互联网（即上游代理）发出关于此模块的请求
- 从 VCS 的 `github.internal.com` 下载模块代码（在其存储机制中）

### 公共模块排除列表

公共模块排除列表也是通配符模式，告诉 Athens 代理它永远不应该从任何上游代理下载这些模块。例如，字符串 `github.com/arschles/**` 告诉 Athens 代理始终向客户端返回 `404 Not Found`。

## 目录端点

代理提供了一个 `/catalog` 服务端点，用于获取本地存储中包含的所有模块及其版本。该端点接受一个分页Token和一个页面大小参数来分页查询。

查询格式为：

`https://proxyurl/catalog?token=foo&pagesize=47`

其中 token 是可选的分页参数，pagesize 是返回页大小。
首次调用时不需要 `token` 参数，它用于分页查询。

结果是以下结构的 json：

```
{"modules": [{"module":"github.com/athens-artifacts/no-tags","version":"v1.0.0"}],
 "next":""}'
```

如果没有返回 `next` token，则表示当前是最后一页。默认分页大小为 1000。
