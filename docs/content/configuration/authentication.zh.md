# 认证配置

Athens 支持多种认证方式以保护您的私有模块代理服务。本文档将指导您如何配置不同的认证方式。

## 基本认证

要启用基本认证，请在配置文件中添加以下内容：

```yaml
auth:
  basic:
    username: your_username
    password: your_password
```

## API 密钥认证

Athens 支持通过 API 密钥进行认证：

```yaml
auth:
  api_key: your_api_key_here
```

## OAuth 认证

对于更复杂的认证需求，Athens 支持 OAuth 2.0：

```yaml
auth:
  oauth:
    client_id: your_client_id
    client_secret: your_client_secret
    token_url: https://your.oauth.provider/token
```

## 多认证方式

您可以同时配置多种认证方式，Athens 会按顺序尝试：

```yaml
auth:
  basic:
    username: user
    password: pass
  api_key: key123
```

## 安全建议

1. 始终使用 HTTPS
2. 定期轮换凭据
3. 使用强密码策略
4. 限制 API 密钥的访问范围

> 注意：生产环境中请勿使用示例中的凭据，这些仅用于演示目的。