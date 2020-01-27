---
title: "Installation"
date: 2017-10-17T15:26:15Z
draft: false
weight: 10
---

## Download Hugo theme

If you have git installed, you can do the following at the command-line-interface within the Hugo directory:

```
cd themes
git clone https://github.com/thingsym/hugo-theme-techdoc.git
```

For more information read [the Hugo documentation](https://gohugo.io/themes/installing-and-using-themes/).

## Configure

You may specify options in config.toml (or config.yaml/config.json) of your site to make use of this theme's features.

For an example of `config.toml`, see [config.toml](https://github.com/thingsym/hugo-theme-techdoc/blob/master/exampleSite/config.toml) in exampleSite.

See [the Configuration documentation](../configuration/).

## Preview site

To preview your site, run Hugo's built-in local server.

```
hugo server -t hugo-theme-techdoc
```

Browse site on http://localhost:1313
