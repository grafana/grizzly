⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️

**WARNING:** This repo is no longer maintained. It's archived and read-only.

⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️⚠️

# Cupper

[![Netlify Status](https://api.netlify.com/api/v1/badges/bc8c4e51-37ee-419d-ad4f-b378010ee546/deploy-status)](https://app.netlify.com/sites/cupper-hugo-theme/deploys)

An accessibility-friendly Hugo theme, ported from the [original Cupper](https://github.com/ThePacielloGroup/cupper) project.

## Table of contents

<!-- toc -->

- [Demo](#demo)
- [Minimum Hugo version](#minimum-hugo-version)
- [Installation](#installation)
- [Updating](#updating)
- [Run example site](#run-example-site)
- [Configuration](#configuration)
- [Nav Title or Logo](#nav-title-or-logo)
- [Favicons](#favicons)
- [Shortcodes](#shortcodes)
- [Syntax highlighting](#syntax-highlighting)
- [Enable Table of Contents for a Blog Post](#enable-table-of-contents-for-a-blog-post)
- [Localization](#localization)
- [Custom CSS and JS](#custom-css-and-js)
- [Default to Dark Theme](#default-to-dark-theme)
- [Enable utterances](#enable-utterances)
- [Non-Git Repo](#non-git-repo)
- [Getting help](#getting-help)
- [Credits](#credits)

<!-- tocstop -->

## Demo

https://cupper-hugo-theme.netlify.app/

## Minimum Hugo version

Hugo version `0.81.0` or higher is required. View the [Hugo releases](https://github.com/gohugoio/hugo/releases) and download the binary for your OS.

## Installation

From the root of your site:

```
git submodule add https://github.com/zwbetz-gh/cupper-hugo-theme.git themes/cupper-hugo-theme
```

## Updating

From the root of your site:

```
git submodule update --remote --merge
```

## Run example site

From the root of `themes/cupper-hugo-theme/exampleSite`:

```
hugo server --themesDir ../..
```

## Configuration

Copy `config.yaml` from the [`exampleSite`](https://github.com/zwbetz-gh/cupper-hugo-theme/tree/master/exampleSite), then edit as desired.

## Nav Title or Logo

- The `navTitleText` param will be checked in your config file. **If** this param exists, the text value will be used as the nav title
- **Otherwise**, a logo will be used as the nav title. Place your **SVG** logo at `static/images/logo.svg`. If you don't provide a logo, then the default theme logo will be used

## Favicons

Upload your image to [RealFaviconGenerator](https://realfavicongenerator.net/) then copy-paste the generated favicon files under `static`.

## Shortcodes

See the [full list of supported shortcodes](https://cupper-hugo-theme.netlify.com/cupper-shortcodes/).

## Syntax highlighting

Syntax highlighting is provided by [Prism](https://prismjs.com/). See this [markdown code fences example](https://cupper-hugo-theme.netlify.com/cupper-shortcodes/#syntax-highlighting).

By default, only a few languages are supported. If you want to add more, follow these steps:

1. Select the languages you want from <https://prismjs.com/download.html>
1. Download the JS file, then copy it to `static/js/prism.js`
1. Download the CSS file, then copy it to `static/css/prism.css`

## Enable Table of Contents for a Blog Post

Set `toc` to `true`. For example:

```
---
title: "My page with a few headings"
toc: true
---
```

## Localization

The strings in the templates of this theme can be localized. Make a copy of `<THEME_BASE_FOLDER>/i18n/en.yaml` to `<YOUR_SITE_FOLDER>/i18n/<YOUR_SITE_LANGUAGE>.yaml`, and translate one by one, changing the `translation` field.

[Here is a tutorial that goes more in depth about this.](https://regisphilibert.com/blog/2018/08/hugo-multilingual-part-2-i18n-string-localization/)

## Custom CSS and JS

You can provide an optional list of custom CSS files, which must be placed inside the `static` dir. These will load after the theme CSS loads. So, `static/css/custom_01.css` translates to `css/custom_01.css`.

You can provide an optional list of custom JS files, which must be placed inside the `static` dir. These will load after the theme JS loads. So, `static/js/custom_01.js` translates to `js/custom_01.js`.

See the [example site config file](https://github.com/zwbetz-gh/cupper-hugo-theme/blob/master/exampleSite/config.yaml) for sample usage.

## Default to Dark Theme

In the site config file set the param `defaultDarkTheme` to true.

E.g. for `config.yaml`
```yaml
params:
  defaultDarkTheme: true
```

Note that the default of light or dark theme only applies to the first visit to a site using this theme. Once the site is visited the choice of dark or light theme is stored in 'local storage' in the browser.

To reset to a 'first visit' scenario (e.g. for testing), one needs to either browse in private mode (aka Incognito/InPrivate/etc) or delete 'local storage' for this site. The easiest way to do that, but which affects other sites as well, is to use the 'Clear History' feature of the browser.

Check your browser's help or documentation for details.

## Enable utterances

`utterances` is a lightweight comments widget built on GitHub issues.

Firstly, choose the repository utterances will connect to:
1. Make sure the repo is public, otherwise your readers will not be able to view the issues/comments.
2. Make sure the [utterances app](https://github.com/apps/utterances) is installed on the repo, otherwise users will not be able to post comments.
3. If your repo is a fork, navigate to its settings tab and confirm the issues feature is turned on.

Secondly,In the site config file set the param `utterances.repo` to enable it.

E.g. for `config.yaml`
```yaml
params:
  utterances:
    repo: username/username.github.io
    issueTerm: title
    theme: github-light
```

Refer to [utterances](https://utteranc.es/) for more information!

## Non-Git Repo

If your site is **not** a git repo, then set `enableGitInfo` to `false` in your config file.

## Getting help

If you run into an issue that isn't answered by this documentation or the [`exampleSite`](https://github.com/zwbetz-gh/cupper-hugo-theme/tree/master/exampleSite), then visit the [Hugo forum](https://discourse.gohugo.io/). The folks there are helpful and friendly. **Before** asking your question, be sure to read the [requesting help guidelines](https://discourse.gohugo.io/t/requesting-help/9132).

## Credits

Thank you to [Heydon Pickering](http://www.heydonworks.com) and [The Paciello Group](https://www.paciellogroup.com/) for creating the original Cupper project.
