---
title: 常见问题
description: 常见问题解答
---

### Athens 只是一个代理吗？还是一个注册表？

_简而言之："注册表"并不能描述 Athens 在这里想要做的事情。这意味着世界上只有一个为所有人提供 Go 模块的服务。Athens 并不打算成为那样的存在。相反，Athens 试图成为一个联邦式模块代理小组的一部分。_

注册表通常由一个实体运行，是一个逻辑服务器，提供认证（有时还提供溯源），几乎是依赖项的事实上唯一来源。有时它由营利性公司运行。

这绝对不是我们在 Athens 社区中要去实现的目标，如果真的走那条路，那将损害我们的社区。

首先，Athens 是 [Go 模块下载 API](/intro/protocol) 的一个 _实现_ 。不仅标准 Go 工具链支持该 API 的任何实现，而且 Athens 代理还设计为可以与任何其他实现该 API 的服务器通信。这使得 Athens 能够与社区中的其他代理进行通信。

最后，我们有意识地以这种方式构建这个项目——并与工具链人员合作——让每个想要编写代理的人都能参与进来。

### Athens 是否与 go 工具链集成？

Athens 目前由 [Go v1.12+](https://golang.org/dl) 工具链通过[下载协议](/intro/protocol/)支持。

关于该协议的简要说明，它是一个 REST API，允许 go 工具链（即 go get）查看版本列表并获取特定版本的源代码。

Athens 是一个实现该协议的服务器。它、该协议和工具链（如您所知）都是开源的。

### Athens 提供的包是否不可变？

_简而言之：Athens 确实将代码存储在 CDN 中，并可以选择将代码存储在其他持久化数据存储中。_

更长的版本：

当源代码来自 Github 时，几乎不可能确保不可变构建。长期以来，我们一直被这个问题困扰着。Go 模块下载协议是解决这个问题的绝佳机会。Athens 代理在高层的工作方式非常简单：

1. `go get github.com/my/module@v1` 发生
1. Athens 在其数据存储中查找，找不到
1. Athens 从 Github 下载 `github.com/my/module@v1`（它在后台也使用 go get）
1. Athens 将模块存储在其数据存储中
1. Athens 从其数据存储中永久提供 `github.com/my/module@v1`

重复一下，"数据存储"意味着 CDN（我们目前支持 Google Cloud Storage、Azure Blob Storage 和 AWS S3）或其他数据存储（我们支持 MongoDB、磁盘和其他一些）。

### Athens 代理能否对私有仓库进行认证？

_简而言之：可以，通过在 Athens 代理主机上定义适当的认证配置。_

当在客户端设置 GOPROXY 环境变量时，Go 1.11+ CLI 不会尝试通过类似 `https://example.org/pkg/foo?go-get=1` 的请求来请求元数据。

Athens 在内部使用 `go get`（更准确地说是 `go mod download`），但没有设置 `GOPROXY` 环境变量，这样 `go` 工具将使用其支持的标准认证机制来请求元数据。因此，如果 v1.11 之前的 `go` 对您有效，那么使用 GOPROXY 的 go 1.11+ 也应该有效，前提是 Athens 代理主机配置了适当的认证。

### 我可以完全排除一个模块吗？

可以，这是可能的。代理提供了一个配置文件，允许用户指定哪些模块根本不应该被获取。[过滤模块配置](/configuration/filter/)提供了有关配置文件以及如何排除某些模块的详细信息。

### 我可以指定从上游代理获取模块而不是本地存储吗？

可以，这是可能的。[过滤模块配置](/configuration/filter/)提供了有关配置文件以及如何排除某些模块的详细信息。

### 代理是否支持监控和可观测性？

目前，我们为代理提供结构化日志。此外，我们还添加了链路追踪功能，以帮助开发人员识别关键代码路径并调试延迟问题。虽然日志不需要配置，但链追踪需要安装一些软件。我们目前支持使用 [Jaeger](https://www.jaegertracing.io/)、[GCP Stackdriver](https://cloud.google.com/stackdriver/) 和 [Datadog](https://docs.datadoghq.com/tracing/)（未经测试）导出链路追踪。其他导出器的支持正在进展中。

要试用 Jaeger 追踪，请执行以下操作：

- 将环境设置为开发（否则追踪将被采样）
- 运行 `docker-compose up -d`（该文件位于 athens 源代码根目录），以初始化所需的服务
- 运行演练教程
- 打开 `http://localhost:16686/search`

  可观测性不是 Athens 代理的硬性要求。因此，如果基础设施没有正确设置，它将失败并记录信息日志。例如，如果 Jaeger 没有运行或提供了错误的导出器 URL，代理将继续运行。但是，当导出器后端不可用时，它不会收集任何追踪或指标。

### Athens 支持哪些 VCS 服务器？

Athens 在后台使用 `go mod download`，因此支持 `go mod` 支持的任何内容。

目前包括：

- git
- svn
- hg
- bzr
- fossil

### 我什么时候应该使用 vendor 目录，什么时候应该使用 Athens？

在模块代理（如 Athens）出现之前，Go 社区长期以来一直使用 vendor 目录，因此协作代码的每个小组应该自己决定是想使用 vendor 目录、Athens，还是两者兼用！

（在不使用代理的情况下）使用 vendor 目录的价值在于：

- CI/CD 系统无法访问 Athens（即使它是内部的）
- 当 vendor 目录非常小，以至于从仓库检出比从服务器拉取 zip 文件更快时
- 如果您来自 glide/dep 或其他利用 vendor 目录的依赖管理系统

（在不使用 vendor 目录的情况下）Athens 的价值在于：

- 您有一个新项目
- 您正在将 Go 项目升级为使用 Go 模块
- 您的团队要求您使用 Athens（例如，出于隔离或依赖审计的目的）
- 您的 vendor 目录很大，导致检出缓慢，而从 Athens 下载可以加快构建速度
  - 对于开发人员来说，缓慢的检出不会像 CI 工具那样成为大问题，因为 CI 工具经常需要检出项目的全新副本
- 您想从项目中移除 vendor 目录以：
  - 减少 pull 请求中的噪音
  - 减少在项目中进行模糊文件搜索的难度
