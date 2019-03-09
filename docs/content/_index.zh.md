---
title: "介绍"
date: 2018-12-07T11:44:36+00:00
---

<img src="/banner.png" width="600" alt="Athens Logo"/>

## Athens 是你的 Go Packages 服务器

欢迎，Gophers! 我们期待把 Athens 介绍给你..

在这个站点上，我们详细地记录了 Athens 的细节。我们将会告诉您它都做了什么，它为什么有意义，你可以用它来做什么，以及你如何运行它。下面是简单的概要。

#### Athens 做了什么？

Athens 为你运行 [Go 模块](https://github.com/golang/go/wiki/Modules) 提供服务。它可以为你提供公有和私有的代码，因此，你不需要直接从像 GitHub 或 GitLab 等版本控制系统（VCS）上拉取。

#### 它为什么有意义？

你需要代理服务器（如安全性和性能）的原因有很多。[看一下](/zh/intro/why)这里的描述的。

#### 我如何使用它？

你自己可以轻易地运行起来 Athens.我们给你几个选项： 

- 可以在你的系统上以二进制的方式运行
    - 稍后会有相关指令
- 你可以用 [Docker](https://www.docker.com/) 镜像的方式来运行(查看[这里](./install/shared-team-instance/)对如何做的介绍)
- 你可以在 [Kubernetes](https://kubernetes.io) 上运行它(查看[这里](./install/shared-team-instance/)对如何做的介绍)

我们还运行了一个体验版本的 Athens,因此你什么都不需要安装也能开始。为此，你需要设置环境变量 `GOPROXY="https://athens.azurefd.net"`.

**[喜欢你听到的吗？现在尝试一下 Athens 吧！](/zh/try-out)**

## 还没有准备好尝试 Athens?

这里有一些其他的参与方法：

- 阅读完整的[指南](/walkthrough)，设置、运行并测试 Athens 代理，进行深入地探索。
* 加入我们的[开发者周例会](/contributing/community/developer-meetings/)！这是一个很好的方法，与大家见面、提问或者只是旁听。我们欢迎任何人加入并参与。
* 查看我们的问题列表中的 [good first issues](https://github.com/gomods/athens/issues?q=is%3Aopen+is%3Aissue+label%3A%22good+first+issue%22)
* 在 [Gophers Slack](https://invite.slack.golangbridge.org/) 上的 `#athens` 频道中加入我们

---
Athens banner attributed to Golda Manuel
