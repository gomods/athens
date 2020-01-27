---
title: "Configuration"
date: 2017-10-17T15:26:15Z
draft: false
weight: 20
---

You may specify options in config.toml (or config.yaml/config.json) of your site to make use of this theme’s features.

For an example of `config.toml`, see [config.toml](https://github.com/thingsym/hugo-theme-techdoc/blob/master/exampleSite/config.toml) in exampleSite.

## Params

    # Souce Code repository section
    description = "put your description"
    github_repository = "https://github.com/thingsym/hugo-theme-techdoc"
    version = "0.2.0"

    # Documentation repository section
    # documentation repository (set edit link to documentation repository)
    github_doc_repository = "https://github.com/thingsym/hugo-theme-techdoc"

    # Analytic section
    google_analytics_id = "" # Your Google Analytics tracking id
    tag_manager_container_id = "" # Your Google Tag Manager container id
    google_site_verification = "" # Your Google Site Verification

    # Theme settings section
    dateformat = "" # default "2 Jan 2006"

    # path name excluded from document menu
    menu_exclusion = ["archives", "blog", "entry", "post", "posts"]

#### `description`

The document summary

default: `put your description`

#### `github_repository`

URL of souce code repository

default: `https://github.com/thingsym/hugo-theme-techdoc`

#### `version`

The version of souce code

default: `0.2.0`

#### `github_doc_repository`

URL of documentation repository for editting

default: `https://github.com/thingsym/hugo-theme-techdoc`

#### `google_analytics_id`

ID of Google Analytics

default: `""`

Container ID of Google Tag Manager

#### `tag_manager_container_id`

default: `""`

#### `google_site_verification`

Content value in meta tag `google-site-verification` for Google Search Console

```
<meta name="google-site-verification" content="e7-viorjjfiihHIoowh8KLiowhbs" />
```

default: `""`

#### `dateformat`

default: `""` as `2 Jan 2006`

#### `menu_exclusion`

Path name excluded from documentation menu

By default, we exclude commonly used folder names in blogs.

default: `["archives", "blog", "entry", "post", "posts"]`
