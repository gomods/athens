---
title: "配置 Athens"
date: 2018-10-16T12:14:01-07:00
weight: 3
---

## 配置 Athens

本文档介绍如何使用各种配置方案来配置 Athens 应用程序。

>本文档涵盖了一些常用的配置，还有很多其他配置！如果想查看所有可设置的配置，我们已在 [此配置文件](https://github.com/gomods/athens/blob/main/config.dev.toml) 中完整记录。

### 认证

作为开发者，我们有多种版本控制系统可用。本节概述如何通过提供各种格式的凭证来使用它们。

 - [认证](/configuration/authentication)

### 存储

Athens 支持多种存储选项。本节描述如何配置这些存储选项。

 - [存储](/configuration/storage)

### 上游代理

本节描述如何配置上游代理，以便从 Go 模块仓库（如 [GoCenter](https://gocenter.io)、[Go 模块镜像](https://proxy.golang.org) 或另一个 Athens 服务器）获取所有模块。

  - [上游](/configuration/upstream)

### 代理 Checksum 数据库

本节描述如何代理 Checksum 数据库，详见 https://go.googlesource.com/proposal/+/master/design/25530-sumdb.md

- [Checksum](/configuration/sumdb)
