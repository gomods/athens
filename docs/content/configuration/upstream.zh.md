---
title: 使用上游 Go 模块仓库（已弃用）
description: 如何配置 Athens 从上游模块仓库（如 GoCenter 或另一个 Athens 服务器）获取缺失的模块
weight: 7
---

>注意：本文档介绍的过滤文件已弃用。请参阅["使用下载模式文件进行过滤"](/configuration/download)获取有关如何在 Athens 中设置上游仓库的最新说明。

默认情况下，Athens 从 GitHub.com 等上游版本控制系统（VCS）获取模块代码，但可以配置为使用 Go 模块仓库（如 GoCenter 或另一个 Athens 服务器）。

1. 创建一个过滤文件（例如 ```/usr/local/lib/FilterForGoCenter```），第一行放置字母 `D`（代表"直接访问"）。更多详细信息，请参阅 [过滤模块](/configuration/filter) 文档。

    ```
    # 用于直接从上游获取模块的 FilterFile
    D
    ```
2. 如果您不使用配置文件，请创建一个新的配置文件（基于示例 config.dev.toml）并编辑值以匹配您的环境。此外，在当前或新配置文件中，按如下方式设置建议的参数：

    ```
    FilterFile = "/usr/local/lib/FilterForGoCenter"
    GlobalEndpoint = "https://<url_to_upstream>"
    # 例如，要使用 GoCenter，请将 <url_to_upstream> 替换为 gocenter.io
    # 您也可以使用 https://proxy.golang.org 来使用 Go 模块镜像
    ```
1. 使用更新后的配置文件重启 Athens。

    ```
     <path_to_athens>/proxy  -config_file <path-to updated  current or new configfile>
    ```
1. 使用["试用 Athens"文档](/try-out)中的步骤验证新配置，并完成相同的演练示例。
