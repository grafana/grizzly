# Docs

This is a gohugo based docs site.

It will be automatically built by github actions, then deployed to the `gh-pages` branch of this repo.

You can view the rendered docs at [https://grafana.github.io/grizzly/](https://grafana.github.io/grizzly/)

To run a development server while editing docs, you will need to install the latest Hugo, then do the following:

```
git submodule init
git submodule update
hugo server -D -s docs
```

You should then be able to view the docs at http://localhost:1313/grizzly/

Changes will instantly be rendered, without page reloads being required. Neat.