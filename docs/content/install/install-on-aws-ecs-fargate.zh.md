# 在 AWS ECS Fargate 上安装 Athens

本文档将指导您如何在 AWS ECS Fargate 上设置 Athens 代理。

## 先决条件

- 已安装并配置 AWS CLI
- 已创建 ECS 集群
- 已创建 ECR 仓库
- 已创建 VPC 和安全组

## 步骤 1: 构建并推送 Docker 镜像

1. 克隆 Athens 仓库：
   ```bash
   git clone https://github.com/gomods/athens
   cd athens
   ```

2. 构建 Docker 镜像：
   ```bash
   docker build -t athens .
   ```

3. 标记并推送镜像到 ECR：
   ```bash
   docker tag athens:latest <your-ecr-repo-url>/athens:latest
   docker push <your-ecr-repo-url>/athens:latest
   ```

## 步骤 2: 创建 ECS 任务定义

1. 在 AWS 控制台中，导航到 ECS > 任务定义 > 创建新任务定义
2. 选择 Fargate 启动类型
3. 配置任务定义：
   - 任务定义名称：athens
   - 任务角色：选择适当的 IAM 角色
   - 网络模式：awsvpc
   - 任务内存：1GB
   - 任务 CPU：0.5 vCPU
4. 添加容器：
   - 容器名称：athens
   - 镜像：<your-ecr-repo-url>/athens:latest
   - 端口映射：3000
   - 环境变量：
     - ATHENS_DISK_STORAGE_ROOT：/var/lib/athens
     - ATHENS_STORAGE_TYPE：disk
5. 创建任务定义

## 步骤 3: 创建 ECS 服务

1. 在 ECS 集群中，点击"创建服务"
2. 配置服务：
   - 启动类型：FARGATE
   - 任务定义：选择之前创建的 athens 任务定义
   - 集群：选择您的集群
   - 服务名称：athens-service
   - 任务数量：1
3. 配置网络：
   - VPC：选择您的 VPC
   - 子网：选择公共子网
   - 安全组：选择允许端口 3000 的安全组
4. 创建服务

## 步骤 4: 验证安装

1. 获取服务的公共 IP 地址
2. 在浏览器中访问 http://<public-ip>:3000
3. 您应该看到 Athens 欢迎页面

## 后续步骤

- 配置负载均衡器以实现高可用性
- 设置自动扩展策略
- 配置监控和日志记录

## 故障排除

- 检查 ECS 任务日志
- 验证安全组规则
- 确保 IAM 角色具有适当权限