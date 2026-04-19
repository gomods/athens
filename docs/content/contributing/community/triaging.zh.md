---
title: "Pull Request 分类"
date: 2018-08-24T17:01:56-07:00
weight: 4

---

您好，Gopher！我们很高兴您有兴趣参与 PR 分类。本文详细介绍如何操作，让我们开始吧！

# 摘要

我们共同努力确保所有 [Pull Requests](https://github.com/gomods/athens/pulls)（PR）得到高效审查和合并。因此我们建立了一种简单的方式，让任何人在星期一、星期三或星期五对 Pull Request 进行"分类"。

PR 分类意味着查看[较旧的 PR](https://github.com/gomods/athens/pulls?q=is%3Apr+is%3Aopen+sort%3Aupdated-asc)，并根据情况执行以下一项或两项：

- 提示审查者回来重新审查
- 提示提交者回来处理审查

_任何人都可以做到这一点，这是参与社区的好方式。_

**请在[此处](https://docs.google.com/spreadsheets/d/1EVUSJc7xm1hXXatzCmp9e8XFsJuW8Uiui5MNkt6ijvw/edit?usp=sharing)报名参加分类。**


# 介绍

Athens 社区共同努力跟进 [问题](https://github.com/gomods/athens/issues) 和 [Pull Requests](https://github.com/gomods/athens/pulls)。对于问题，我们每周花时间审查下一个里程碑中的问题和其他人们感兴趣的问题。

我们努力让 PR 审查更快、更高效，所以我们每周审查三次。

PR 审查是异步的：

- 提交 PR
- 您在审查中留下反馈
- 提交者稍后阅读并处理您的反馈（例如更改代码或回复评论）
- 那之后您回来重新审查

我个人喜欢异步工作流程，但生活中总会有意外——人们会忘记、会忙、会休假等等……这完全正常！我们都是人，需要休息。

问题是 PR 审查可能会停滞。因此，确保 PR 不会闲置太久很重要。

我们让一个人每周三次来检查，确保较旧的 PR 仍然受到关注。

# 分类日程

由于我们的 PR 数量不是特别多，我们在寻找人们在分类日执行以下操作：

- 查看[最近 3 天未更新的 PR](https://github.com/gomods/athens/pulls?q=is%3Apr+is%3Aopen+sort%3Aupdated-asc)
- 添加评论以提示审查者和提交者回到 PR：
    - 如果自一个或多个审查者完成审查后 PR 中添加了新提交，请提示那些审查者回来重新审查
    - 如果仍有待处理的评论且提交者尚未解决，请提示提交者查看新评论
- 如果您看到超过 10 天未更新的 PR，请在 PR 评论中写这个，我们会来弄清楚发生了什么（可能会直接联系某人或关闭 PR）：
    ```@gomods/maintainers this PR is really old!```

如果您需要在分类中提示某人，请在 GitHub 上像这样[提及其人](https://blog.github.com/2011-03-23-mention-somebody-they-re-notified/)：@arschles can you look at this again?。如果您注意到已经有人被 @过了，您可以在 Slack 上尝试提醒他们。如果您提醒他们，请友善一点，请记住他们可能正忙于其他事情 :)

# 如何报名？

任何人，无论背景、经验、对项目熟悉程度、时区或其他任何因素，都可以参与。这是参与项目的好方式。

如果您想在特定日期进行分类，请将您的姓名添加到[分类电子表格](https://docs.google.com/spreadsheets/d/1EVUSJc7xm1hXXatzCmp9e8XFsJuW8Uiui5MNkt6ijvw/edit?usp=sharing)中。

如果您以前从未做过分类并想要开始，请[提交一个问题](https://github.com/gomods/athens/issues/new?template=first_triage.md)。

如果这些都不清楚，请在 [Gophers Slack](https://invite.slack.golangbridge.org/) 的 `#athens` 频道联系我们，我们会弄清楚并让您开始。

# 这可以自动化吗？

可能可以！但我们不确定 PR 应该何时"被提示"以及机器人应该怎么做。也许我们能通过这个过程了解到这些标准。

尽管如此，有人情味还是好的。这与我们的[理念](https://github.com/gomods/athens/blob/main/PHILOSOPHY.md)非常吻合。
