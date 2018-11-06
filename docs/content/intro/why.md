---
title: "Why Does It Matter?"
date: 2018-11-06T13:58:58-07:00
weight: 4
LastModifierDisplayName: "Marwan"
LastModifierEmail: "marwan.sameer@gmail.com"

---

### Immutability

Previously, the Go community has had lots of problems with libraries disappearing or changing without warning. It's easy for package maintainers to make changes to their code that can break yours - and much of the time it's an accident! Could your build break if one of your dependencies did this?

- Commit `abdef` was deleted
- Tag `v0.1.0` was force pushed
- The repository was deleted altogether

 Since your app's dependencies come directly from GitHub, any of those above cases can happen to you and your builds can break when they do - oh no! Athens solves these problems by copying code from VCS's into _immutable_ storage.

 This way, you don't need to upload anything manually to Athens storage. The first time Go asks Athens for a dependency, Athens will go get it from VCS (github, bitbucket etc). But once that module has been retrieved, it will be forever persisted in its storage backend and will never go back to VCS or that same version again. This is how Athens achieves module immutability. Keep in mind, you are in charge of that storage backend. 


### Logic 

The fact that the Go command line can now ping _your own_ server to download dependencies, that means you can program whatever logic you want around providing such dependencies. Things like Access Control (discussed below), adding custom versions, custom forks, custom packages etc. 


### Performance 

Downloading stored dependencies from Athens is _significantly_ faster than downloading dependencies from Version Control Systems. For example, it takes 10 seconds to download the CockroachDB source tree as a zip file from GitHub but almost four minutes to git clone it. 

### Access Control 

Worse than packages disappeaing, packages can be malicious. Therefore, you can make sure that no program inside your company or team will ever install github.com/some-user/malicious-package. This is because Go will come to your server and ask for that same exact package, and instead of fetching it from github, you can just return a 500 to the Go command line causing the build to fail. With Athens, you can achieve this through the [filter file](/configuration/filter.md). 