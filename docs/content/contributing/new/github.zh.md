---
title: "Github"
date: 2018-09-20T13:58:58-07:00
weight: 2
LastModifierDisplayName: "Robbie"
LastModifierEmail: "hello@robloranger.ca"

---

### 使用 GitHub

我们使用 [GitHub](https://github.com) 来托管仓库的远程副本，跟踪 [issues](https://github.com/gomods/athens/issues) 并管理 [Pull Requests](https://github.com/gomods/athens/pulls)。如果您还没有注册，请现在就花一分钟去注册。我们会在您回来时在这里等您。

#### issue

在 GitHub 上，我们用"issue"来记录代码库中需要的每项变更。包括想法、错误、功能请求、支持请求，甚至只是讨论。

如果您对代码有疑问或认为可能发现了错误，请随时提交问题。有时人们会觉得部分文档或说明难以遵循，如果发生这种情况也请告诉我们。

对于我们的 GitHub 项目，有错误报告和功能请求的模板，还有一个提案模板。当您点击"新建问题"时，系统会提示您选择模板。如果这些模板都不符合您的需求，也可以选择"打开常规问题"——即空白问题。

![GitHub 问题跟踪器](/github-issue-header.png)
![GitHub 问题模板选择](/github-issue-templates.png)

#### Fork

为保持主仓库整洁，我们要求在创建 Pull Request 之前都在所谓的 fork 上进行工作。fork 本质上是一个由**您的** GitHub 账户拥有的仓库副本。您可以在那里创建任意数量的分支，尽管放心大胆地提交。我们会在合并前将它们全部压缩成一个，下文会有更多说明。

请先花一分钟阅读这篇关于创建和维护项目 fork 的[帖子](https://kbroman.org/github_tutorial/pages/fork.html)。它还介绍了如何将其他协作者的 fork 添加为远程仓库，这在您开始与其他贡献者合作时会发现很有用。

#### Pull requests

[Brian Ketelsen](https://twitter.com/bketelsen) 制作了一个关于制作您的第一个开源 Pull Request 的精彩视频，要了解流程概述请[观看](https://www.youtube.com/watch?v=bgSDcTyysRc)。还有[很棒的视频系列](https://egghead.io/courses/how-to-contribute-to-an-open-source-project-on-github)，由 [Kent C. Dodds](https://twitter.com/kentcdodds) 讲解如何为开源项目做贡献。

当我们收到新的 Pull Request 时，可能需要一些时间让维护者或其他贡献者处理。以下是处理时会发生的事情。

你可能会看到不少评论，但别让它们打击你的信心。我们所有人都在努力帮助彼此写出最优秀的代码和文档，任何批评都应该当作建设性的意见来看待。

> 如果您觉得不是这样，请随时联系维护者之一。我们非常认真地对待我们的[行为准则](https://www.contributor-covenant.org)。

很可能一位或多位贡献者和维护者会在您的 Pull Request 上留下审查意见。您可以在 Pull Request 本身中讨论这些更改，或者如果您需要特定方面的帮助，可以在 [Gophers Slack](https://gophers.slack.com) 的 `#athens` 频道联系我们。

在所有请求的更改都得到解决后，维护者会进行最终检查。只要持续集成通过，他们就会将其合并到 `main` 分支。

#### 项目流程

下面简要介绍向 Athens 贡献的通用指南。

##### 找到问题

当您看到想处理的问题时，如果没有人表示有兴趣，请用类似以下内容评论该问题：

- 我来处理这个
- 我想处理这个

这让其他贡献者知道有人正在处理它。我们都知道空闲时间很难得，不想让任何人浪费时间。

如果您没有看到关于您的错误或功能的问题，请提交一个问题与社区讨论。有些事情可能不符合项目的最佳利益，或者可能已经讨论过了。

如果问题很小，比如拼写错误或断开的链接，您可以直接提交 Pull Request。

##### 开启 Pull Request

在 fork 上创建分支并完成更改后，请确保所有测试通过，详情请参阅 [DEVELOPMENT.md](https://github.com/gomods/athens/blob/main/CONTRIBUTING.md#verify-your-work)。然后将所有更改推送到您的 fork 仓库后，前往 [Athens](https://github.com/gomods/athens) 开启 Pull Request。通常在您推送新分支并访问原始仓库后，GitHub 会提示您打开新的 Pull Request。或者您可以从仓库的 `Pull Requests` 选项卡执行此操作。
