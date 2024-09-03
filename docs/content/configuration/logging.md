---
title: Logging
description: Configure the logger for your desired output
weight: 9
---

Athens is designed to support a myriad of logging scenarios.

## Standard

The standard structured logger can be configured in `plain` or `json` formatting via `LogFormat` or `ATHENS_LOG_FORMAT`. Additionally, verbosity can be controlled by setting `LogLevel` or `ATHENS_LOG_LEVEL`. In order for the standard structured logger to work, `CloudRuntime` and `ATHENS_CLOUD_RUNTIME` should not be set to a valid value.

The logging is via [Logrus](https://github.com/sirupsen/logrus), so the allowed values for logging config options are determined by that project. For example, `ATHENS_LOG_LEVEL` can be `debug`, `info`, `warn`/`warning`, `error`, etc.

## Runtimes

Athens can be configured according to certain cloud provider specific runtimes. The **GCP** runtime configures Athens to rename certain logging fields that could be dropped or overriden when running in a GCP logging environment. This runtime can be used with `LogLevel` or `ATHENS_LOG_LEVEL` to control the verbosity of logs.
