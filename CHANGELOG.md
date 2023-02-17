# Changelog

## 0.2.1 (2023-01-12)
This release is resolves a number of security vulnerabilities.
* Updates to Golang 1.19.1 (#224) thanks @rhowe!
* Update Jsonnet/Jsonnet Bundler versions (#231)
* Revert to external `cortextool`, simplifying security patch maintenance (#231)
* Updated go dependencies (#234)
* Updates the Docker image to a newer version, bringing in updated versions of openssl and zlib

## 0.2.0 (2022-05-04)

This release provides fixes and improvements over 0.2.0-beta-3
* Add missing space in log message (#212)
* Improve jsonnet parser for alerts/rules (#211)
* Fix access to Grafana without any auth (#210)
* Update jsonnet to 0.18.0 (#198)
* Modify requests to send authorization header (#197)
* Fix docker arm build (#194)
* Documenting the use of a proxy with grizzly (#193)
* Docker: add jsonnet, jsonnetfmt and jsonnet-bundler (#192)

## 0.2.0-beta3 (2021-10-28)

This release provides fixes and improvements over 0.2.0-beta-2
* Set dashboard folder to a default value when it is undefined (#183)
* Add debug logging for most commands(#188)
* Correctly match synthetic monitoring locally-generated IDs to currently deployed checks (#189)

## 0.2.0-beta2 (2021-10-11)

This release provides fixes to problems introduced by 0.2.0-beta1.
* Improved error messages when the referenced Jsonnet file does not exist (#170)
* Previously, `lib` and `vendor` directories were looked for relative to the
  Jsonnet script. This makes it hard to work with larger Jsonnet codebases. This
  has changed to looking for these directories in the current directory. Ideally,
  we would have a concept of the "root" of a code tree, and locate `lib` and
  `vendor` directories there. In lieu of that, using the current directory is
  better. (#172)
* Documentation corrections for `grr pull` and Synthetic Monitoring (#174 and #173)
* Fixed `grr pull` so that it correctly pulls dashboards in the General folder.

## 0.2.0-beta1 (2021-09-23)

Much has changed in Grizzly, including the switch to Kubernetes style resources.

This version needs some testing (hence the beta label) and some work on the new
tests we've added.

Here's some of the things that are new:
* Targets now use dots consistently (#163)
* Handlers are applied in a consistent order, so that folders get created before
  dashboards (#162)
* Listen feature (temporarily) removed - Grafana API is not yet stable enough for
  this (#161)
* Switch Jsonnet examples to use a Grizzly library (#157)
* Addition of fuller documentation
* CortexTool no longer required for interacting with hosted Prometheus
* `watch` now watches directories recursively (#122)
* Grizzly can now list remote resources (`grr list -r`) (#120)
* Folders are now first-class citizens (#115)
* Grizzly can pull remote resources to local filesystem (`grr pull`) (#110)
* Grizzly can now be configured with YAML instead of Jsonnet (#100)
* Resources are now consumed as Kubernetes style objects (#94). Hidden resources
  in Jsonnet are now deprecated (#94)

## 0.1.1 (2021-08-02)

Add binary for Darwin Arm64 arch

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
