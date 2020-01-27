# Hugo Theme Techdoc

The Techdoc is a Hugo Theme for technical documentation.

![The Techdoc screenshot](/images/screenshot.png "screenshot")

## Features

* Modern, Simple layout
* Responsive web design
* Documentation menu
* Edit link to documentation repository
* Custom Shortcodes
* Analytics with Google Analytics, Google Tag Manager

## Getting Started

### Download Hugo theme

If you have git installed, you can do the following at the command-line-interface within the Hugo directory:

```
cd themes
git clone https://github.com/thingsym/hugo-theme-techdoc.git
```

For more information read [the Hugo documentation](https://gohugo.io/themes/installing-and-using-themes/).

### Configure

You may specify options in config.toml (or config.yaml/config.json) of your site to make use of this theme's features.

For an example of `config.toml`, [config.toml](https://github.com/thingsym/hugo-theme-techdoc/blob/master/exampleSite/config.toml) in exampleSite.

### Preview site

To preview your site, run Hugo's built-in local server.

```
hugo server -t hugo-theme-techdoc
```

Browse site on http://localhost:1313

## Development environment

```
cd /path/to/hugo-theme-techdoc
yarn install
gulp watch
```

## Contribution

### Patches and Bug Fixes

Small patches and bug reports can be submitted a issue tracker in Github. Forking on Github is another good way. You can send a pull request.

1. Fork [Hugo Theme Techdoc](http://thingsym.github.io/hugo-theme-techdoc/) from GitHub repository
2. Create a feature branch: git checkout -b my-new-feature
3. Commit your changes: git commit -am 'Add some feature'
4. Push to the branch: git push origin my-new-feature
5. Create new Pull Request

## Changelog

* Version 0.2.0
  * add screenshot images
  * add exampleSite
  * fix sub-menu for responsive
  * improve menu and pagination
* Version 0.1.0
  * initial release

## License

Licensed under the MIT License.

## Author

[thingsym](https://github.com/thingsym)

Copyright (c) 2017-2018 by thingsym
