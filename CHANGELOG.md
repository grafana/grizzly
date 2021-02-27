# Changelog

## 0.1.0 (2021-02-27)

This is a substantial release.

### Multiple 'providers'
The most significant change is that Grizzly now supports multiple 'providers'. It
can now target multiple observability systems, focused on those offered by
Grafana Labs. That is, you can now use Grizzly to manage:

 * Grafana dashboards
 * Grafana datasources
 * Prometheus rules with Grafana Cloud Metrics or Cortex Ruler
 * Blackbox checks with Grafana Synthetic Monitoring.

### `listen`
Grafana 7.3 has a feature that informs all browsers that have a dashboard open
when a user saves that dashboard. All instances of that dashboard are then
automatically reloaded. To use this feature, you must enable the `live` feature
flag on Grafana itself.

Grizzly takes advantage of this feature with its `grr listen` command. Here,
Grizzly will connect to Grafana, and wait for changes to a specific dashboard.
When that dashboard changes, it writes the JSON for that dashboard to a local
file. The Grizzly user can then commit that file to version control (e.g. Git)
and we have the beginnings of a version controlled workflow for Grafana
Dashboards.

### Jsonnet 0.17
Grizzly now uses [Jsonnet 0.17](https://github.com/google/jsonnet/releases/tag/v0.17.0) (up from 0.15.1).

### Adopting Kubernetes Style Naming
Naming of resources has been confusing in Grizzly for some time. Is it a 'dashboard',
a 'grafana-dashboard', etc. In this release, we have opted for following the Kubernetes
approach for naming resources, with there being a group, a version and a kind. At
present, all resource types (dashboards, prometheus rules, etc) are within a single
group and version (`grizzly.grafana.com/v1alpha1`).

Other improvements:
 * Better `grr show` pager, with more intuitive key bindings
 * Targets can now include globs, e.g. `-t Dashboard.*`
 * `diff` has switched to showing the whole resource to a unified diff

## 0.0.1 (2020-07-31)

Initial release of Grizzly. Although increasingly used, this is still alpha code.
Please help us test it, and provide us feedback.

Initial version provides support for interacting with Grafana and its dashboards.
