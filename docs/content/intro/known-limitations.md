---
title: "Known Limitations"
date: 2018-11-01T13:58:58-07:00
weight: 4
LastModifierDisplayName: "Marwan"
LastModifierEmail: "marwan.sameer@gmail.com"

---

Go Modules is still behind an experimental flag and therefore it's important to know how Athens can help you *now* until all issues in the upstream of Go modules are resolved. 

#### What Athens can do well _now_ 

- Athens works will within a CI/CD system that installs your modules without needing a vendor folder.
- Athens works well inside your own company without exposing it to unlimited public traffic. 

### Known Limitations

#### Unable to download a module from the proxy by a commit hash
Go Modules work nicely with go get. You can `go get` a version from Athens by doing something like: `GOPROXY=<athens-url> go get github.com/pkg/errors@v0.8.0`.

However, sometimes you just want to install a package by its branch or commit sha, such as `go get github.com/pkg/errors@master` or `go get github.com/pkg/errors@059132a15dd08d6704c67711dae0cf35ab991756`

This works well without using a GOPROXY. But there's a [known issue](https://github.com/golang/go/issues/27947) that prevents 
 `GOPROXY=<athens-url> go get pkg@commit-hash` from working.
 
 Until this issue is resolved and Athens adapts a solution for it as well, we recommend that you don't use Athens to get a module by a revision or branch. 


##### Workaround: 

When you want to include a new library in your program by "branch", just do `go get github.com/my/pkg@master`. This will put a real semver pseudo version inside your go.mod file. From that point on, you can use the GOPROXY to get the same exact version. 