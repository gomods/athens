# 在 Google App Engine (GAE) 上安装 Athens

## 前提条件
- 已安装 Google Cloud SDK
- 已创建 GCP 项目
- 已在 GCP 项目上启用结算功能

## 安装步骤

1. 克隆 Athens 仓库：
   ```bash
   git clone https://github.com/gomods/athens
   cd athens
   ```

2. 创建 app.yaml 配置文件：
   ```yaml
   runtime: go116
   service: athens
   env_variables:
     ATHENS_DISK_STORAGE_ROOT: /tmp
     ATHENS_STORAGE_TYPE: disk
   ```

3. 部署到 GAE：
   ```bash
   gcloud app deploy
   ```

4. 访问你的 Athens 实例：
   ```bash
   gcloud app browse
   ```

## 故障排除
- 查看日志：`gcloud app logs tail -s athens`
- 验证部署：`gcloud app versions list`