# 日志配置

Athens 提供了灵活的日志配置，帮助您监控和调试应用程序。以下是主要配置选项：

## 日志级别
- `DEBUG`: 用于调试的详细信息
- `INFO`: 常规操作信息
- `WARNING`: 表示潜在问题
- `ERROR`: 需要关注的重要问题
- `CRITICAL`: 严重故障

## 配置示例
```yaml
logging:
  level: INFO
  format: "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
  file: /var/log/athens.log
  max_size: 10485760  # 10MB
  backup_count: 5
```

## 可用选项
- `level`: 要捕获的最低日志级别
- `format`: 日志消息格式
- `file`: 日志文件路径（可选）
- `max_size`: 日志文件轮转前的最大大小
- `backup_count`: 保留的备份文件数量