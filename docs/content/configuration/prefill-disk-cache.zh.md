# 预填充磁盘缓存

Athens 提供了预填充磁盘缓存的功能，以提高性能并确保模块的可用性。

## 配置

要预填充磁盘缓存，请在 Athens 配置中添加以下内容：

```yaml
PreFill:
  Enabled: true
  Modules:
    - github.com/gomods/athens
    - github.com/gin-gonic/gin
```

## 使用

当 Athens 启用预填充功能启动时，它将：
1. 下载指定的模块
2. 将它们存储在磁盘缓存中
3. 使它们立即可供客户端使用

这在以下情况下特别有用：
- 减少首次下载延迟
- 确保关键依赖项始终可用
- 提高隔离环境中的性能