# 上游配置

Athens 支持配置上游仓库进行依赖解析。这允许您控制依赖项的获取来源。

## 配置选项

- `URL`: 上游仓库的基础 URL
- `Type`: 仓库类型 (例如：代理、直接)
- `Timeout`: 连接超时时间（秒）
- `Cache`: 启用/禁用依赖项缓存

## 配置示例

```yaml
upstream:
  url: https://proxy.golang.org
  type: proxy
  timeout: 30
  cache: true
```