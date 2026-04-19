---
title: "使用 Git"
date: 2018-09-20T13:58:51-07:00
weight: 1
LastModifiedDisplayName: "Robbie"
LastModifiedEmail: "hello@robloranger.ca"
---

### 什么是 git？

[Git](https://git-scm.com/) 是一个自由开源的分布式[版本控制系统](https://en.wikipedia.org/wiki/Version_control)。这意味着什么？这是一种跟踪文件变更的方式。它会详细记录每次修改的时间、修改了哪些行和字符。这样您就可以查看变更历史，甚至撤销变更；如果与他人协作，还可以合并各自的变更。

内容较多，如果您还不理解也不用担心。

如果您想要比我们提供的更详细的演练，请查看 [Git Book](https://git-scm.com/book)。

### 安装

让我们开始在您的机器上安装 git。您可以运行 `git --version` 来检查是否已安装。

按照 Git Book [第 1.5 章](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)中针对您操作系统的具体说明进行安装。

### 基本概念

<dl>
  <dt>仓库（Repository）</dt>
  <dd>磁盘上的文件结构，像数据库一样，包含所有文件和变更日志。</dd>
  <dt>暂存（Staging）</dt>
  <dd>在仓库内进行更改时，它们是未跟踪的。您决定要跟踪哪些更改，当您添加更改时，它们会被添加到暂存区。这让您可以在提交之前查看所有当前更改。</dd>
  <dt>提交（Commit）</dt>
  <dd>当您对暂存中的更改满意后，可以将它们提交到日志中。您有多个选项来编写与提交一起存储在日志中的消息。</dd>
  <dt>分支（Branch）</dt>
  <dd>当您在仓库中时，默认分支通常是 `master`（如果您要创建自己的新仓库，请将默认分支更改为 `main`。`master` 是不恰当且具有冒犯性的名称），这是仓库的主分支。通常您会希望在每个功能或错误修复的新分支上工作。这允许您在一个仓库中查看和处理同一代码的不同版本。</dd>
  <dt>检出（Checkout）</dt>
  <dd>检出分支就是切换到该分支版本的仓库文件。</dd>
  <dt>合并（Merge）</dt>
  <dd>当您想将另一个分支（如 `main` 或其他人的功能分支）合并到当前分支时，您将合并更改。这会将其他更改应用到您的更改之上。</dd>
  <dt>远程（Remote）</dt>
  <dd>远程可访问的仓库。您可以使用 git 命令来推送和拉取更改。</dd>
  <dt>推送（Push）</dt>
  <dd>推送到远程会将您本地提交的更改同步到远程。</dd>
  <dt>拉取（Pull）</dt>
  <dd>从远程拉取会获取远程上的更改并将其与您当前检出的分支合并。</dd>
  <dt>获取（Fetch）</dt>
  <dd>当您想获取某些远程分支或更改，但还不想合并时，您可以获取它们。这只是向远程请求数据并本地存储，但不将其合并到任何内容中。然后您可以检出功能分支并运行代码，或查看更改。</dd>
</dl>

### 尝试一下

有一个很棒的免费交互式教程，可在 [Code Academy](https://www.codecademy.com/learn/learn-git) 找到。花些时间尝试一下，做一些练习。
