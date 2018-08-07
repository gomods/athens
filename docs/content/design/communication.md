---
title: "Communication"
date: 2018-02-11T15:57:56-05:00
---

## Communication flow

This is the story of a time long ago. A time of myth and legend when the ancient Gods were petty and cruel and they plagued build with irreproducibility.
Only one project dared to challenge their power...Athens. Athens possessed a strength the world had never seen. Filling its cache.

### Clean plate

At the beginning, there's theoretical state when a cache of proxy and registry is empty.

When User makes a request at this ancient time, it works as described on the flow below.

- User contacts proxy asking for module M in version v1.0
- Proxy checks whether or not it has this module in its storage. It has not.
-So it responds with a redirect to Olympus and schedules its own job to check with Olympus in few units of time, giving Olympus space to fill its cache.
- User receives redirect to Olympus and asks it for module M in version v1.0
- Olympus is cache free as well, so it asks the underlying VCS (e.g github.com) for a module.
- After it receives all the bits, it stores it into its own cache and serves it to the User.
- User receives module and is happy.
- Sometimes around this time proxy asks Olympus for module M in version v1.0 as well. Olympus, now aware of this module, serves it so proxy can fill its own cache.

![Communication flow for clear state](/athens-clear-scenario.png)

### New proxy joins the party

At this point, we have 1 proxy and 1 registry, each of them aware about module M. Now new proxy joins with an empty cache.

We can see the flow is very similar.

- User contacts new proxy, which checks internal storage to find out it is missing the module.
    - Redirects to Olympus,
    - Schedules new cache fill job.
- User contacts Olympus aware of the module and receives the response right away.
- Proxy, after some time, contacts Olympus and fills its cache.

![Communication flow for new proxy](/athens-new-proxy-old-olympus-scenario.png)


### Happy path

Now we have all proxies and Olympus aware of module M. So when a new user asks for M in version v1.0 it is served right away. Without proxy bothering Olympus nor VCS.

![Communication flow for new proxy](/athens-proxy-filled.png)


### Asking about private things

There are times when you do not want the mighty gods of Olympus to know about your desires. E.g:
- You are requesting private module,
- Communication is just disabled.

In this case
- User contacts proxy asking for a private module.
- Proxy detects this repo is private and checks its storage. It does not find it there.
- Proxy contacts VCS directly.
- VCS responds with a module which is then stored in a cache Synchronously.
- The module is served to the User.

![Communication flow for new proxy](/athens-private-repo-scenario.png)
