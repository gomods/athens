---
title: "Why Does It Matter?"
date: 2018-11-06T13:58:58-07:00
weight: 4
LastModifierDisplayName: "Marwan"
LastModifierEmail: "marwan.sameer@gmail.com"

---

### Immutability

The Go community has had lots of problems with libraries disappearing or changing without warning. It's easy for package maintainers to make changes to their code that can break yours - and much of the time it's an accident! Could your build break if one of your dependencies did this?

- Commit `abdef` was deleted
- Tag `v0.1.0` was force pushed
- The repository was deleted altogether

 Since your app's dependencies come directly from a VCS (Version Control System, such as GitHub), any of those above cases can happen to you and your builds can break when they do - oh no! Athens solves these problems by copying code from VCS's into _immutable_ storage.

 This way, you don't need to upload anything manually to Athens storage. The first time Go asks Athens for a dependency, Athens will go get it from VCS (github, bitbucket etc). But once that module has been retrieved, it will be forever persisted in its storage backend and the proxy will never go back to VCS for that same version again. This is how Athens achieves module immutability. Keep in mind, you are in charge of that storage backend. 

### Logic 

The fact that the Go command line can now ping _your own_ server to download dependencies, that means you can program whatever logic you want around providing such dependencies. Things like Access Control (discussed below), adding custom versions, custom forks, and custom packages. For example, Athens provides a [Validation Hook](https://github.com/gomods/athens/blob/master/config-example.toml#L87) that will get called for every module download to determine whether a module should be downloaded or not. Therefore, you can extend Athens with your own logic such as scanning a module path or code for red flags etc. 


### Performance 

Downloading stored dependencies from Athens is _significantly_ faster than downloading dependencies from Version Control Systems. This is because `go get` by default uses VCS to download modules such as `git clone`, while `go get` with GOPROXY enabled will use HTTP to download zip archives. Therefore, depending on your computer and internet connection speed, it takes 10 seconds to download the CockroachDB source tree as a zip file from GitHub but almost four minutes to git clone it. 

### Access Control 

Worse than packages disappearing, packages can be malicious. To make sure no such malicious package is ever installed by your team or company, you can have your proxy server return a 500 when the Go command line asks for an excluded module. This will cause the build to fail because Go expects a 200 HTTP response code. With Athens, you can achieve this through the [filter file]({{< ref "/configuration/filter.md" >}}).


### Vendor Directory Becomes Optional
With immutability, performance, and a robust proxy server, there's no longer an absolute need for each repository to have its vendor directory checked in to its version control. The go.sum file ensures that no package is manipulated after the first install. Furthermore, your CI/CD can install all of your dependencies on every build with little time. 
