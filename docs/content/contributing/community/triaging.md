---
title: "Triaging Pull Requests"
date: 2018-08-24T17:01:56-07:00
weight: 4

---

Hi, Gopher! We're glad you're interested in getting into PR triaging. This page details how to do that. Let's get started!

# TL;DR

We're trying to all work together to make sure all of our [pull requests](https://github.com/gomods/athens/pulls) (PRs) get reviewed and merged efficiently. So, we set up an easy way for anyone to "triage" pull requests on any Monday, Wednesday or Friday. 

PR triaging means looking at [older PRs](https://github.com/gomods/athens/pulls?q=is%3Apr+is%3Aopen+sort%3Aupdated-asc) and do either or both of these things, as appropriate:

- Prompting reviewers to come back and re-review
- Prompting submitters to come back and address reviews  

_Absolutely anyone can do triaging, and this is a great way to get involved with the community._

**Sign up for triaging [here](https://docs.google.com/spreadsheets/d/1EVUSJc7xm1hXXatzCmp9e8XFsJuW8Uiui5MNkt6ijvw/edit?usp=sharing).**


# Intro

The Athens community all works together to keep up to date on [issues](https://github.com/gomods/athens/issues) and [pull requests](https://github.com/gomods/athens/pulls). For issues, we take some time each week to review issues in the next milestone and others that folks are interested in.

We try to keep PR reviews moving a little bit faster and more efficiently, so we look at those 3 times a week.

PR reviews are asynchronous:

- A PR gets submitted
- You leave feedback in your review
- The submitter reads and addresses it (e.g. change code or respond to the comment) your feedback sometime later
- You come back and re-review sometime after that

I personally love the asynchronous workflow, but life happens - people forget, people get busy, go on vacation, etc... - as they should! We're all human and we need breaks like that.

The problem is that PR reviews can get stalled. So, it's important to make sure that a PR doesn't sit idle for too long.

We're getting a person to come check in three times a week to make sure that older PRs are still getting attention.

# Triage Schedule

Since we don't have a super huge volume of PRs, we're looking for folks to do the following on a triage day:

- Look at [PRs not updated in the last 3 days](https://github.com/gomods/athens/pulls?q=is%3Apr+is%3Aopen+sort%3Aupdated-asc)
- Add a comment to prompt reviewers and the submitter to come back to the PR:
    - If new commits have been added to the PR since 1 or more reviewer has done a review, please prompt those reviewers to come back and re-review
    - If there are still comments pending and the submitter hasn't addressed them, please prompt the submitter to look at the new comments
- If you see a PR that hasn't been updated in more than 10 days, write this in the PR comments and we'll come figure out what's going on (probably contact someone directly or close the PR):
    ```@gomods/maintainers this PR is really old!```

If you need to prompt someone in your triage, do it by [mentioning](https://blog.github.com/2011-03-23-mention-somebody-they-re-notified/) someone on GitHub like this: `@arschles can you look at this again?`. If you notice that someone has been `@mentioned` already, you can try pinging them on Slack. If you ping them, be nice and remember that they might be busy with other things though :)

# How do I Sign Up?

Anyone, regardless of background, experience, familiarity with the project, time zone, or pretty much anything else. This is a wonderful way to get involved with the project. Triages generally take 15 minutes or less if you've done a few before (see the bottom of this section if you haven't).

If you'd like to do triaging on a particular day, please add your name to the [triaging spreadsheet](https://docs.google.com/spreadsheets/d/1EVUSJc7xm1hXXatzCmp9e8XFsJuW8Uiui5MNkt6ijvw/edit?usp=sharing).

If you haven't done a triage before and would like to get started, please [submit an issue](https://github.com/gomods/athens/issues/new?template=first_triage.md).

If any of this doesn't make sense, please contact us in the `#athens` channel in the [Gophers Slack](https://invite.slack.golangbridge.org/) and we'll clear it up and get you started.

# Can this be Automated?

Probably, yes! But we don't know if there are exact criteria on when PRs should be "prompted" and how a bot should do that. Maybe we'll learn those criteria here.

Even still, it's nice to have a human touch as a submitter and reviewer. It matches our [philosophy](https://github.com/gomods/athens/blob/master/PHILOSOPHY.md) very well.
