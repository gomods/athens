# 开发指南

欢迎来到 Athens 开发指南！本文档将帮助您开始为项目做出贡献。

## 先决条件

- 已安装 Go 1.20+
- 已安装 Docker
- 对 Go 和微服务有基本了解

## 环境设置

1. 克隆仓库：
```bash
git clone https://github.com/athens/athens.git
```

2. 安装依赖：
```bash
make deps
```

3. 启动开发环境：
```bash
make run
```

## 运行测试

运行所有测试：
```bash
make test
```

运行特定测试：
```bash
go test ./path/to/package
```

## 代码风格

- 遵循 Go 官方编码规范
- 使用 gofmt 进行格式化
- 编写清晰简洁的注释