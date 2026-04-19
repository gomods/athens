---
title: 日志配置
description: 为您想要的输出配置日志记录器
weight: 9
---

Athens 被设计为支持多种日志记录场景。

## 标准日志

标准结构化日志记录器可通过 `LogFormat` 或 `ATHENS_LOG_FORMAT` 配置为 `plain` 或 `json` 格式。此外，还可通过设置 `LogLevel` 或 `ATHENS_LOG_LEVEL` 来控制日志详细程度。为了使标准结构化日志记录器正常工作，`CloudRuntime` 和 `ATHENS_CLOUD_RUNTIME` 不应设置为有效值。

日志记录通过 [Logrus](https://github.com/sirupsen/logrus) 实现，日志配置选项的允许值由该项目决定。例如，`ATHENS_LOG_LEVEL` 可以是 `debug`、`info`、`warn`/`warning`、`error` 等。

## 运行时

Athens 可根据特定云提供商的运行时进行配置。**GCP** 运行时配置 Athens 重命名某些在 GCP 日志环境中可能被丢弃或覆盖的日志字段。此运行时可与 `LogLevel` 或 `ATHENS_LOG_LEVEL` 结合使用来控制日志详细程度。
