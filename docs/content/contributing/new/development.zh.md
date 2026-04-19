---
title: "如何贡献"
date: 2018-10-30T17:48:51-07:00
weight: 3
---
# Athens 开发指南

代理使用符合 Go 语言习惯的写法编写，使用标准工具。如果您了解 Go，您就能够阅读代码并运行服务器。

Athens 使用 [Go Modules](https://golang.org/cmd/go/#hdr-Modules__module_versions__and_more) 进行依赖管理。您需要 [Go v1.12+](https://golang.org/dl) 才能开始开发 Athens。

请参阅我们的[贡献指南](https://github.com/gomods/athens/blob/main/CONTRIBUTING.md)，了解提交 Pull Request 的注意事项。

### Go 版本

Athens 基于 Go v1.12+ 开发。

如需使用其他版本的 Go，请设置以下环境变量：
```
GO_BINARY_PATH=go1.12.X
# 或您想与 athens 一起使用的任何其他二进制文件
```

# 运行代理

如果您在 GOPATH 内，请确保 `GO111MODULE=on`；如果您在 GOPATH 外，Go Modules 默认开启。主包在 `cmd/proxy` 中，运行方式与普通 Go 项目相同：

```
cd cmd/proxy
go build
./proxy
```

服务器启动后，控制台会输出类似以下内容：

```console
Starting application at 127.0.0.1:3000
```

### 依赖项

# Athens 需要的服务

Athens 需要多个服务（如数据库等）才能正常运行。我们使用 [Docker](http://docker.com/) 镜像来配置和运行这些服务。**但是，Athens 默认不需要任何存储依赖项**。默认存储在内存中，您也可以选择使用 `fs`，同样无需依赖项。

如果您不熟悉 Docker，也没关系。我们已尽力简化启动和运行流程：

1. [下载并安装 docker-compose](https://docs.docker.com/compose/install/)（docker-compose 是一个用于轻松一次启动和停止多个服务的工具）
2. 从仓库根目录运行 `make dev`

完成！`make dev` 命令执行完毕后，所有服务都将启动并运行，您可以继续下一步。

如需停止所有内容，请运行 `make down`。

请注意，`make dev` 仅启动工作所需的最小依赖项。如需启动所有可能的依赖项，请运行 `make alldeps` 或直接运行 `docker-compose.yml` 文件中可用的服务。但请记住，`make alldeps` 不会启动 Athens 或 Olympus，只会启动它们的**依赖项**。

# 运行单元测试

运行单元测试前，必须先启动相关依赖服务：

```console
make alldeps
```

依赖服务启动后，即可运行单元测试：

```console
make test-unit
```

# 运行文档

我们提供了 Docker 镜像用于文档开发，运行 [Hugo](https://gohugo.io/) 来渲染文档。

```
make docs
docker run -it --rm \
        --name hugo-server \
        -p 1313:1313 \
        -v ${PWD}/docs:/src:cached \
        gomods/hugo
```

完成后请访问 [http://localhost:1313](http://localhost:1313/)。

# 代码检查

我们的 CI/CD 流程使用 govet，建议您提前在本地运行检查：

```
go vet ./...
```
