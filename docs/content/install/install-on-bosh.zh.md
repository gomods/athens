# 在 BOSH 上安装 Athens

## 前提条件

在开始之前，请确保您已经具备以下条件：

- 已安装并配置好 BOSH CLI
- 已部署 BOSH Director
- 已准备好 BOSH 环境
- 已下载 Athens 的 BOSH 发行版

## 部署步骤

1. 创建 BOSH 部署清单文件：

```yaml
---
name: athens
releases:
- name: athens
  version: latest
instance_groups:
- name: athens
  instances: 1
  vm_type: default
  stemcell: default
  azs: [z1]
  networks:
  - name: default
  jobs:
  - name: athens
    release: athens
    properties:
      port: 3000
      storage:
        type: disk
        disk:
          path: /var/vcap/store/athens
```

2. 使用 BOSH CLI 部署 Athens：

```bash
bosh -e my-env -d athens deploy athens.yml
```

3. 验证部署状态：

```bash
bosh -e my-env -d athens instances
```

## 配置说明

- **port**: Athens 服务监听的端口号
- **storage.type**: 存储类型，支持 disk 和 memory
- **storage.disk.path**: 当使用 disk 存储时的存储路径

## 访问 Athens

部署完成后，您可以通过以下方式访问 Athens：

```bash
http://<athens-vm-ip>:3000
```

## 维护操作

- 查看日志：

```bash
bosh -e my-env -d athens logs
```

- 重启服务：

```bash
bosh -e my-env -d athens restart
```

- 更新版本：

```bash
bosh -e my-env -d athens deploy athens.yml --recreate
```

## 故障排除

1. 如果部署失败，请检查 BOSH Director 日志
2. 确保网络配置正确
3. 验证存储路径权限
4. 检查防火墙设置，确保端口 3000 可访问