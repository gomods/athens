# SumDB 配置

SumDB 是 Athens 用于验证模块完整性的校验和数据库。

## 配置选项

- `SUMDB_URL`: SumDB 服务器地址 (默认: `https://sum.golang.org`)
- `SUMDB_PUBLIC_KEY`: 用于验证 SumDB 签名的公钥
- `SUMDB_READ_TIMEOUT`: SumDB 读取操作超时时间
- `SUMDB_WRITE_TIMEOUT`: SumDB 写入操作超时时间

## 配置示例

```yaml
sumdb:
  url: "https://sum.golang.org"
  public_key: "sum.golang.org+033de0ae+Ac4zctda0e5eza+HJyk9SxEdh+s3Ux18htTTAD8OuAn8"
  read_timeout: "5s"
  write_timeout: "5s"
```