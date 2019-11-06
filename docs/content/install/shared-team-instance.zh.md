---
title: 共享团队实例
description: 为开发团队安装Athens实例
weight: 2
---
当您按照[Walkthrough](/walkthrough)中的说明进行操作时，Athens最终使用的是本地存储空间。 这仅适用于短期试用Athens，因为您将很快耗尽内存，并且Athens在两次重启之间不会保留模块。 本指南将帮助您以一种更适合的方式运行Athens，以用于提供一个实例供开发团队共享的场景。

我们将使用Docker来运行Athens，因此首先请确保您已经[安装Docker](https://docs.docker.com/install/).

## 选择存储提供程序

Athens目前支持许多存储驱动程序。 对于本机使用，建议从使用本地磁盘作为存储提供程序开始使用。对于其他提供商，请参阅
 [the Storage Provider documentation](/configuration/storage).


## 使用本地磁盘作为存储安装Athens


为了使用本地磁盘存储来运行Athens，您接下来需要确定要将模块持久化的位置。 在下面的示例中，我们将在当前目录中创建一个名为`athens-storage`的新目录。现在您可以在启用磁盘存储的情况下运行Athen。 要启用本地磁盘存储，您需要在运行Docker容器时设置`ATHENS_STORAGE_TYPE`和`ATHENS_DISK_STORAGE_ROOT`环境变量。

为了简单起见，下面的示例使用`：latest` Docker标记，但是我们强烈建议您在环境启动并运行后切换到使用正式版本（例如`：v0.3.0`）。

**Bash**
```bash
export ATHENS_STORAGE=~/athens-storage
mkdir -p $ATHENS_STORAGE
docker run -d -v $ATHENS_STORAGE:/var/lib/athens \
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
md -Path $env:ATHENS_STORAGE
docker run -d -v "$($env:ATHENS_STORAGE):/var/lib/athens" `
   -e ATHENS_DISK_STORAGE_ROOT=/var/lib/athens `
   -e ATHENS_STORAGE_TYPE=disk `
   --name athens-proxy `
   --restart always `
   -p 3000:3000 `
   gomods/athens:latest
```

注意：如果您之前尚未使用Docker for Windows挂载此驱动器，则可能会提示您允许访问

Athens现在应该作为带有本地目录`athens-storage`的Docker容器运行。当Athens检索模块(module)时，它们将被存储在先前创建的目录中。首先，让我们确认雅典是否在运行：

```console
$ docker ps
CONTAINER ID        IMAGE                               COMMAND           PORTS                    NAMES
f0429b81a4f9        gomods/athens:latest   "/bin/app"        0.0.0.0:3000->3000/tcp   athens-proxy
```

现在，我们可以从安装了Go v1.12+的任何机器上使用Athens。 要验证这一点，请尝试以下示例：

**Bash**
```console
$ export GO111MODULE=on
$ export GOPROXY=http://127.0.0.1:3000
$ git clone https://github.com/athens-artifacts/walkthrough.git
$ cd walkthrough
$ go run .
go: downloading github.com/athens-artifacts/samplelib v1.0.0
The 🦁 says rawr!
```


**PowerShell**
```console
$env:GO111MODULE = "on"
$env:GOPROXY = "http://127.0.0.1:3000"
git clone https://github.com/athens-artifacts/walkthrough.git
cd walkthrough
$ go run .
go: downloading github.com/athens-artifacts/samplelib v1.0.0
The 🦁 says rawr!
```

我们可以通过检查Docker日志来验证Athens是否处理了此请求：

```console
$ docker logs -f athens-proxy
time="2018-08-21T17:28:53Z" level=warning msg="Unless you set SESSION_SECRET env variable, your session storage is not protected!"
time="2018-08-21T17:28:53Z" level=info msg="Starting application at 0.0.0.0:3000"
handler: GET /github.com/athens-artifacts/samplelib/@v/v1.0.0.info [200]
handler: GET /github.com/athens-artifacts/samplelib/@v/v1.0.0.mod [200]
handler: GET /github.com/athens-artifacts/samplelib/@v/v1.0.0.zip [200]
```

现在，如果查看`athens_storage`文件夹的内容，您将会看到与samplelib模块相关的新增文件。

**Bash**
```console
$ ls -lr $ATHENS_STORAGE/github.com/athens-artifacts/samplelib/v1.0.0/
total 24
-rwxr-xr-x  1 jeremyrickard  wheel    50 Aug 21 10:52 v1.0.0.info
-rwxr-xr-x  1 jeremyrickard  wheel  2391 Aug 21 10:52 source.zip
-rwxr-xr-x  1 jeremyrickard  wheel    45 Aug 21 10:52 go.mod
```

**PowerShell**
```console
$ dir $env:ATHENS_STORAGE\github.com\athens-artifacts\samplelib\v1.0.0\


    Directory: C:\athens-storage\github.com\athens-artifacts\samplelib\v1.0.0


Mode                LastWriteTime         Length Name
----                -------------         ------ ----
-a----        8/21/2018   3:31 PM             45 go.mod
-a----        8/21/2018   3:31 PM           2391 source.zip
-a----        8/21/2018   3:31 PM             50 v1.0.0.info
```


重新启动Athens后，它将在该位置提供模块（module），而无需重新下载。 为了验证这一点，我们需要首先删除Athens容器。

```console
docker rm -f athens-proxy
```

接下来，我们需要清除本地Go模块中的缓存。 这是必要的，以便您本地的Go命令行工具从Athens重新下载该模块。 以下命令将清除本地存储中的模块：

**Bash**
```bash
sudo rm -fr "$(go env GOPATH)/pkg/mod"
```

**PowerShell**
```powershell
rm -recurse -force $(go env GOPATH)\pkg\mod
```

现在，我们重新运行Athens容器

**Bash**
```console
docker run -d -v $ATHENS_STORAGE:/var/lib/athens \
   -e ATHENS_DISK_STORAGE_ROOT=/var/lib/athens \
   -e ATHENS_STORAGE_TYPE=disk \
   --name athens-proxy \
   --restart always \
   -p 3000:3000 \
   gomods/athens:latest
```

**PowerShell**
```console
docker run -d -v "$($env:ATHENS_STORAGE):/var/lib/athens" `
   -e ATHENS_DISK_STORAGE_ROOT=/var/lib/athens `
   -e ATHENS_STORAGE_TYPE=disk `
   --name athens-proxy `
   --restart always `
   -p 3000:3000 `
   gomods/athens:latest
```

当我们重新运行我们的Go示例时，Go cli将再次从Athens下载模块。然而，Athens不需要重新检索模块。它将从本地磁盘中获取。

**Bash**
```console
$ ls -lr $ATHENS_STORAGE/github.com/athens-artifacts/samplelib/v1.0.0/
total 24
-rwxr-xr-x  1 jeremyrickard  wheel    50 Aug 21 10:52 v1.0.0.info
-rwxr-xr-x  1 jeremyrickard  wheel  2391 Aug 21 10:52 source.zip
-rwxr-xr-x  1 jeremyrickard  wheel    45 Aug 21 10:52 go.mod
```

**PowerShell**
```console
$ dir $env:ATHENS_STORAGE\github.com\athens-artifacts\samplelib\v1.0.0\


    Directory: C:\athens-storage\github.com\athens-artifacts\samplelib\v1.0.0


Mode                LastWriteTime         Length Name
----                -------------         ------ ----
-a----        8/21/2018   3:31 PM             45 go.mod
-a----        8/21/2018   3:31 PM           2391 source.zip
-a----        8/21/2018   3:31 PM             50 v1.0.0.info
```

请注意文件的时间戳并没有更改

下一步:

* [通过helm在Kubernetes上运行Athens](/install/install-on-kubernetes)
* 查看Athens在生产环境上的最佳实践. [即将发布](https://github.com/gomods/athens/issues/531)
