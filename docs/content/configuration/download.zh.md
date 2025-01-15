# 下载配置

## 概述
本节介绍如何在 Athens 中配置下载设置。

## 配置选项
- **下载路径**: 指定文件保存的目录
- **并发下载**: 设置最大同时下载数量
- **超时设置**: 配置连接和下载超时
- **重试策略**: 定义失败下载的重试方式

## 配置示例
```yaml
download:
  path: /var/downloads
  concurrent: 5
  timeout: 300
  retry:
    attempts: 3
    delay: 10
```