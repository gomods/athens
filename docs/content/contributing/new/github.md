---
title: "Github"
date: 2018-09-20T13:58:58-07:00
weight: 2
LastModifierDisplayName: "Robbie"
LastModifierEmail: "hello@robloranger.ca"

---

### Using GitHub

We use [GitHub](https://github.com) to host the remote copy of our
[repository](https://github.com/gomods/athens) and both track our
[issues](https://github.com/gomods/athens/issues) and manage our
[pull requests](https://github.com/gomods/athens/pulls). If you haven't signed
up before, take a minute to go do that now. We'll be right here when you get
back.

#### Issues

On GitHub we use the concept of an issue to record every change needed in our
code base. This means ideas, bugs, features, support, even just discussions.

Never hesitate to open an issue if you have question about the code or think you
may have found a bug. Sometimes people will find part of the documentation or
instructions difficult to follow, let us know if this happens.

In particular for our project on GitHub, we have a template for Bug Reports and
Feature Requests with one for Proposals .. proposed. When you click on 'New
Issue' you will be given a choice of which template to start with, if they do
not seem to fit your need there is an option to 'Open a regular issue' - which
is blank.

![GitHub Issue Tracker](/github-issue-header.png)
![GitHub Issue Template Selection](/github-issue-templates.png)

#### Forks

In order to keep things a bit tidy in our main repository, we all do work on
what is called a fork before creating a pull request. A fork is essentially a
copy of a repository that is owned by **your** GitHub account. You can create as
many branches as you like there and don't be shy with creating commits. We will
squash them all into one before we merge, more on that below.

But first take a minute to read this
[awesome post](https://kbroman.org/github_tutorial/pages/fork.html) on creating
and maintaining a fork of the project. It even goes into adding other
collaborators forks as remotes which you will find useful as you start working
more and more with other contributors.

#### Pull requests

[Brian Ketelsen](https://twitter.com/bketelsen) created an awesome video on
making your first open source pull request, for an overview of the process go
[watch that now](https://www.youtube.com/watch?v=bgSDcTyysRc). There's also a
[great video series](https://egghead.io/courses/how-to-contribute-to-an-open-source-project-on-github)
on contributing to an Open Source project by [Kent C. Dodds](https://twitter.com/kentcdodds).

When we receive a new pull request, it may take time for a maintainer or other
contributor to get to it, but when we do a few things will happen.

You can expect at least a few comments, don't let them discourage you though. We
are all trying to help one another write the best code and documentation
possible, all criticism should be considered constructive.

> If you feel like it is
> not, please do not hesitate to reach out to one of the maintainers. We take our
> [code of conduct](https://www.contributor-covenant.org) very seriously.

Most likely one or more contributors and maintainers will leave a review on your
pull request. You can discuss the changes requested in the pull request itself,
or if you need help with something in particular you can reach out to us in the
`#athens` channel on [Gophers Slack](https://gophers.slack.com).

After all requested changes are resolved a maintainer will give it a final look
and as long as our continuous integration passed they will merge it with the
master branch.

#### Project process

Let's just go over a quick general guideline on contributing to Athens.

##### Find an issue

When you see an issue you would like to work on, if no one else has expressed
interest please comment on the issue with something like:

- I will take this
- I would like to work on this

This let's other contributors know someone is working on it, we all know free
time is hard to get and don't want anyone wasting theirs.

If you don't see an issue for your bug or feature, please open one to discuss
with the community. Some things may not be in the best interest of the project,
or may have already been discussed.

If your issue is something very small, like a typo or broken link, you may skip
straight the pull request.

##### Open a pull request

After you have created a branch on your fork, and made the changes. Please make
sure all tests still pass, see [DEVELOPMENT.md](https://github.com/gomods/athens/blob/master/CONTRIBUTING.md#verify-your-work) for details. Then after you push all changes
up to your fork, head over to [Athens](https://github.com/gomods/athens) to open a pull request. Usually,
right after you have pushed a new branch and you visit the original repository,
GitHub will prompt you to open a new pull request. Otherwise you can do so from
the `Pull Requests` tab on the repository.
