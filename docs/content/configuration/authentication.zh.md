---
title: 私有仓库认证
description: 配置 Athens 认证
weight: 2
---

## 认证

## SVN 私有仓库

1.  Subversion 会在以下位置创建认证结构

        ~/.subversion/auth/svn.simple/<hash>

2.  为了正确创建 SVN 服务器的身份验证文件，您需要先对服务器进行身份验证，然后让 SVN 生成正确的哈希文件。

        $ svn list http://<domain:port>/svn/<somerepo>
        Authentication realm: <http://<domain> Subversion Repository
        Username: test
        Password for 'test':

3.  身份验证成功后，我们需要将 .subversion 目录共享给 Athens 代理服务器，以便重用这些凭据。下面我们将其作为卷挂载到代理容器中。

    **Bash**

    ```bash
    export ATHENS_STORAGE=~/athens-storage
    export ATHENS_SVN=~/.subversion
    mkdir -p $ATHENS_STORAGE
    docker run -d -v $ATHENS_STORAGE:/var/lib/athens \
      -v $ATHENS_SVN:/root/.subversion \
      -e ATHENS_DISK_STORAGE_ROOT=/var/lib/athens \
      -e ATHENS_STORAGE_TYPE=disk \
      --name athens-proxy \
      --restart always \
      -p 3000:3000 \
      gomods/athens:latest
    ```

    **PowerShell**

    ```PowerShell
    $env:ATHENS_STORAGE = "$(Join-Path $pwd athens-storage)"
    $env:ATHENS_SVN = "$(Join-Path $pwd .subversion)"
    md -Path $env:ATHENS_STORAGE
    docker run -d -v "$($env:ATHENS_STORAGE):/var/lib/athens" `
      -v "$($env:ATHENS_SVN):/root/.subversion" `
      -e ATHENS_DISK_STORAGE_ROOT=/var/lib/athens `
      -e ATHENS_STORAGE_TYPE=disk `
      --name athens-proxy `
      --restart always `
      -p 3000:3000 `
      gomods/athens:latest
    ```

## Bazaar(bzr) 私有仓库

* * Athens 自带的 Dockerfile 不支持 Bazaar，但相关说明对启用了 Bazaar 的自定义 Athens 构建仍然适用。*

1. Bazaar 配置文件位于

- Unix

      ~/.bazaar/

- Windows

      C:\Documents and Settings\<username>\Application Data\Bazaar\2.0

- 您可以使用以下命令检查您的位置

      bzr version

2. 有三个典型的配置文件

- bazaar.conf
  - 默认配置选项
- locations.conf
  - 针对特定分支的覆盖规则和/或设置
- authentication.conf
  - 远程服务器的凭证信息

3. 配置文件语法

- \# 这是注释
- [header] 这表示章节标题
- 章节选项位于标题章节中，包含选项名称、等号和值

  - 示例：

        [DEFAULT]
        email = John Doe <jdoe@isp.com>

4. 认证配置

   允许指定远程服务器的凭证。
   这可用于所有支持的传输协议以及 bzr 中需要认证的任何部分（如 smtp）。
   该语法遵循与其他语法相同的规则，只是其中的可选策略（option policies）不适用。

   示例：

   [myprojects]
   scheme=ftp
   host=host.com
   user=joe
   password=secret

   # hobby.net 上的个人项目

   [hobby]
   host=r.hobby.net
   user=jim
   password=obvious1234

   # 家庭服务器

   [home]
   scheme=https
   host=home.net
   user=joe
   password=lessobV10us

   [DEFAULT]

   # 我们的本地用户是 barbaz，在所有远程站点上我们称为 foobar

   user=foobar

   注意：使用 sftp 时，协议是 ssh，不支持密码，您应该使用 PPK

   [reference code]
   scheme=https
   host=dev.company.com
   path=/dev
   user=user1
   password=pass1

   # 开发服务器上的开发分支

   [dev]
   scheme=ssh # bzr+ssh 和 sftp 在此可用
   host=dev.company.com
   path=/dev/integration
   user=user2

   #代理
   [proxy]
   scheme=http
   host=proxy.company.com
   port=3128
   user=proxyuser1
   password=proxypass1

5. 认证成功后，我们需要将 bazaar 配置目录共享给 Athens 代理服务器，以便重用这些凭证。下面我们将其作为卷挂载到代理容器中。

   **Bash**

   ```bash
   export ATHENS_STORAGE=~/athens-storage
   export ATHENS_BZR=~/.bazaar
   mkdir -p $ATHENS_STORAGE
   docker run -d -v $ATHENS_STORAGE:/var/lib/athens \
     -v $ATHENS_BZR:/root/.bazaar \
     -e ATHENS_DISK_STORAGE_ROOT=/var/lib/athens \
     -e ATHENS_STORAGE_TYPE=disk \
     --name athens-proxy \
     --restart always \
     -p 3000:3000 \
     gomods/athens:latest
   ```

   **PowerShell**

   ```PowerShell
   $env:ATHENS_STORAGE = "$(Join-Path $pwd athens-storage)"
   $env:ATHENS_BZR = "$(Join-Path $pwd .bazaar)"
   md -Path $env:ATHENS_STORAGE
   docker run -d -v "$($env:ATHENS_STORAGE):/var/lib/athens" `
     -v "$($env:ATHENS_BZR):/root/.bazaar" `
     -e ATHENS_DISK_STORAGE_ROOT=/var/lib/athens `
     -e ATHENS_STORAGE_TYPE=disk `
     --name athens-proxy `
     --restart always `
     -p 3000:3000 `
     gomods/athens:latest
   ```

## Atlassian Bitbucket 和基于 SSH 保护的 Git VCS

本节最初用于描述如何配置 Athens 的 Git 客户端，使其通过 SSH（而非 HTTP）从本地部署的 Atlassian Bitbucket 实例中拉取特定的 Go 导入包。稍作调整后，本节也可为配置 Athens 代理以实现对托管版 Bitbucket 及其他基于 SSH 安全的版本控制系统（VCS）的认证访问提供参考。如果你的开发工作流要求通过 SSH 克隆、推送和拉取 Git 仓库，并且你希望 Athens 以相同方式运行，请继续阅读。。

作为 example.com 的开发人员，假设您的应用程序有一个依赖项，托管在 Bitbucket 上，导入语句如下：

```go
import "git.example.com/golibs/logo"
```

进一步假设您会手动克隆此包，如下所示：

```bash
$ git clone ssh://git@git.example.com:7999/golibs/logo.git
```

`go-get` 客户端（如 Athens 所调用的）会通过查找此输出中的 `go-import` meta 标签来[开始解析](https://golang.org/cmd/go/)此依赖项：

```bash
$ curl -s https://git.example.com/golibs/logo?go-get=1
<?xml version="1.0"?>
<!DOCTYPE html>
<html lang="en">
   <head>
      <meta charset="utf-8">
         <meta name="go-import" content="git.example.com/golibs/logo git https://git.example.com/scm/golibs/logo.git"/>
         <body/>
      </meta>
   </head>
</html>
```

其中说明 Go 导入的实际内容位于 `https://git.example.com/scm/golibs/logo.git` 。将此 URL 与通过 SSH 克隆项目的 URL（上述）进行比较，可以看出这种 [Git 全局配置](https://git-scm.com/docs/git-config) http 到 ssh 重写规则：

```
[url "ssh://git@git.example.com:7999"]
	insteadOf = https://git.example.com/scm
```

因此，为了通过 SSH 获取 `git.example.com/golibs/logo` 依赖项以填充其存储缓存，Athens 最终会调用 git。根据上述重写规则，git 需要对应的 SSH 私钥——该私钥的公钥必须绑定到执行克隆操作的开发者或 Bitbucket 上的服务账号。这本质上与 github.com 的 SSH 模型相同。我们至少需要为 Athens 提供两样东西：一份 SSH 私钥，以及一条将 HTTP 重写为 SSH 的 git 规则，并将它们挂载到 Athens 容器内供 root 用户使用：

```bash
$ mkdir -p storage
$ ATHENS_STORAGE=storage
$ docker run --rm -d \
    -v "$PWD/$ATHENS_STORAGE:/var/lib/athens" \
    -v "$PWD/gitconfig/.gitconfig:/root/.gitconfig" \
    -v "$PWD/ssh-keys:/root/.ssh" \
    -e ATHENS_DISK_STORAGE_ROOT=/var/lib/athens -e ATHENS_STORAGE_TYPE=disk --name athens-proxy -p 3000:3000 gomods/athens:canary
```

`$PWD/gitconfig/.gitconfig` 包含 http 到 ssh 的重写规则：

```
[url "ssh://git@git.example.com:7999"]
	insteadOf = https://git.example.com/scm
```

`$PWD/ssh-keys` 包含上述私钥和一个最小的 ssh-config：

```bash
$ ls ssh-keys/
config		id_rsa
```

我们还提供了一个 ssh config 来绕过主机 SSH 密钥验证，并展示如何为不同主机绑定不同的 SSH 密钥：

`$PWD/ssh-keys/config` 包含：

```
Host git.example.com
Hostname git.example.com
StrictHostKeyChecking no
IdentityFile /root/.ssh/id_rsa
```

现在，通过 Athens 代理执行的构建应该能够通过 SSH 认证的方式克隆 `git.example.com/golibs/logo` 依赖项。

### `SSH_AUTH_SOCK` 和 `ssh-agent` 支持

作为无密码 SSH 密钥的替代方案，可以使用 [`ssh-agent`](https://en.wikipedia.org/wiki/Ssh-agent)。如果`ssh-agent` 设置的 `SSH_AUTH_SOCK` 环境变量包含有效的 Unix 套接字路径（解引用符号链接后），该变量将传递给 go mod download 命令。

因此，如果在一个可用的 ssh agent（且 shell 中已设置 `SSH_AUTH_SOCK`）环境下运行，并在按照上一节所述配置好 `gitconfig` 之后，就可以按如下方式在 Docker 中运行 Athens：

```bash
$ mkdir -p storage
$ ssh-add .ssh/id_rsa_something
$ ATHENS_STORAGE=storage
$ docker run --rm -d \
    -v "$PWD/$ATHENS_STORAGE:/var/lib/athens" \
    -v "$PWD/gitconfig/.gitconfig:/root/.gitconfig" \
    -v "${SSH_AUTH_SOCK}:/.ssh_agent_sock" \
    -e "SSH_AUTH_SOCK=/.ssh_agent_sock" \
    -e ATHENS_DISK_STORAGE_ROOT=/var/lib/athens -e ATHENS_STORAGE_TYPE=disk --name athens-proxy -p 3000:3000 gomods/athens:canary
```

## GitHub Apps

除了使用 GitHub 的机器用户之外，还可以通过创建 GitHub App 并借助其进行身份验证。

在 **Settings > Developer settings > GitHub Apps** 中创建 GitHub App 并安装。需要从 App 中获取 AppID/ClientID、Installation ID 和私钥。

将 [GitHub App Git Credential Helper](https://github.com/bdellegrazie/git-credential-github-app) 设置到您的 `$PATH` 中。Athens Docker 镜像已预装此工具。

按如下方式配置您的 [全局 Git 配置](https://git-scm.com/docs/git-config)：

```
[credential "https://github.com/your-org"]
    helper = "github-app -username <app-name> -appId <app-id> -privateKeyFile <path-to-private-key> -installationId <installation-id>"
    useHttpPath = true

[credential "https://github.com"]
    helper = "cache --timeout=3600"

[url "https://github.com"]
    insteadOf = ssh://git@github.com
```

这指示 Git 使用 GitHub App 进行认证，并将结果缓存 3600 秒（认证令牌有效期为 1 小时）。

现在，通过 Athens 代理执行的构建应该能够通过 GitHub Apps 克隆 `github.com/your-org/your-repo` 依赖项。

### GitHub Enterprise 自托管

要针对自托管的 GitHub Enterprise 进行认证，除了 Git 配置应包含您的域名外，说明与 GitHub 托管 Apps 相同：

```
[credential "https://github.example.com/your-org"]
    helper = "github-app -username <app-name> -appId <app-id> -privateKeyFile <path-to-private-key> -installationId <installation-id> -domain github.example.com"
    useHttpPath = true

[credential "https://github.example.com"]
    helper = "cache --timeout=3600"

[url "https://github.example.com"]
    insteadOf = ssh://git@github.com
```
