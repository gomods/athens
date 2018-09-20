---
title: "Using Git"
date: 2018-09-20T13:58:51-07:00
LastModifiedDisplayName: "Robbie"
LastModifiedEmail: "hello@robloranger.ca"
---

### What is git?

[Git](https://git-scm.com/) is a free and open source distributed [version control system](https://en.wikipedia.org/wiki/Version_control). What does
that really mean? It is a way to track changes to files on your computer. This
is like keeping a detailed log of every time you change a file, what lines and
characters were changed. So you could look at the log and see what changed, or
undo those changes if you wanted or if you are working with others you can
merge your changes together.

It's a lot to take in, so don't worry if you aren't following yet.

If you want a more detailed walk through than we provide here, have a look at
the [Git Book](https://git-scm.com/book).

### Installing

Let's start by getting git installed on you machine. You can check if it's
already installed by running `git --version` from the command line.

Follow your operating system specific instructions in [Chapter
1.5](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git) of the Git
Book.

### Basic concepts

<dl>
  <dt>Repository</dt>
  <dd>This is a file structure on disk, like a database, it contains all files
  and the log of changes.</dd>
  <dt>Staging</dt>
  <dd>When you make changes inside a repository, they are untracked. You
  decide which changes to track, as you add changes they are added to the
  staging area. This let's you see all current changes before committing
  them.</dd>
  <dt>Commit</dt>
  <dd>After you are happy with the changes tracked in staging, you can
  commit them to the log we mentioned. You have a few options for writing a
  message that will be stored with the commit in the log, more on that later.</dd>
  <dt>Branch</dt>
  <dd>When you are in the repository the default is usually called master,
  the main branch of the repository. Typically you will want to do your work
  on a new branch for each feature or bug. This allows you to see and work on
  different versions of the same code in one repository.</dd>
  <dt>Checkout</dt>
  <dd>To check out a branch, is to switch to view that branches version of the
  files in the repository.</dd>
  <dt>Merge</dt>
  <dd>When you want to incorporate another branch, master or someone else's
  feature for example, into your current branch you will merge the changes. This will apply
  the other changes on top of yours.
  <dt>Remote</dt>
  <dd>This is just a repository, that is accessible remotely. You can use the
  git command to push and pull changes to.</dd>
  <dt>Push</dt>
  <dd>Pushing to a remote will synchronize your locally committed changes to the
  remote.</dd>
  <dt>Pull</dt>
  <dd>Pulling from a remote will both fetch and merge the changes on the remote
  with the branch you have currently checked out.</dd>
  <dt>Fetch</dt>
  <dd>When you want to get some remote branch or changes, but not merge them
  yet, you can fetch them. Just ask the remote for the data and store it locally
  but not incorporate it into anything. You could then checkout the feature
  branch and run the code, or read over the changes.</dd>
</dl>

### Try it out

There is a great interactive tutorial, for free, available at [Code Academy](https://www.codecademy.com/learn/learn-git). Take some time to play with it and try out some of the commands.
