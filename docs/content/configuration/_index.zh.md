---
title: "配置 Athens"
date: 2018-10-16T12:14:01-07:00
weight: 3
---

## 配置 Athens
这里我们将介绍如何利用各种配置场景来配置 Athens 应用程序。

> 本节涵盖了一些更常用的配置变量，但还有更多！如果你想查看所有可以设置的配置变量，我们已经在[这个配置文件](https://github.com/gomods/athens/blob/main/config.dev.toml)中记录了它们。

### 认证
作为开发人员，我们可以使用许多版本控制系统。在本节中，我们将概述如何通过为 Athens 项目提供各种格式的所需凭证来使用它们。

 - [认证](/configuration/authentication)
 
### 存储
在 Athens 中，我们支持多种存储选项。在本节中，我们将描述如何配置它们

 - [存储](/configuration/storage)

### 上游代理
在本节中，我们将描述如何配置上游代理以从 Go 模块仓库（如 [GoCenter](https://gocenter.io)、[Go 模块镜像](https://proxy.golang.org) 或其他 Athens 服务器）获取所有模块。

  - [上游](/configuration/upstream)

### 代理校验和数据库
在本节中，我们将描述如何代理校验和数据库，如 https://go.googlesource.com/proposal/+/master/design/25530-sumdb.md 所述

- [校验和](/configuration/sumdb)