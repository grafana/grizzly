---
date: "2021-06-28T00:00:00+00:00"
title: "Home"
---

# Grizzly: A tool for Managing Observability Systems

A utility for managing various observability resources with Jsonnet. Currently supported
are:

 * Grafana dashboards/dashboard folders
 * Grafana datasources
 * Grafana Cloud Prometheus recording rules/alerts
 * Grafana Synthetic Monitoring checks

It is designed to work with existing [monitoring mixins](https://github.com/monitoring-mixins/docs).

> **Status: Alpha**. This is a proof of concept. It will have many holes. PRs welcome.
> **Release / Install** Please see the release page https://github.com/grafana/grizzly/releases/

## Authentication and Configuration

Grizzly has a 10 second timeout on some HTTP calls. To override this behavior, use the `GRIZZLY_HTTP_TIMEOUT=<seconds>` environment variable

Each system handles authentication differently.

### Grafana Itself
This tool interacts with Grafana via its REST API. For this, you will need to
establish authentication credentials. These are provided to `grr` via
environment variables.

| Name | Description | Required | Default |
| --- | --- | --- | --- |
| `GRAFANA_URL` | Fully qualified domain name of your Grafana instance. | true | - |
| `GRAFANA_USER` | Basic auth username if applicable. | false | `api_key` |
| `GRAFANA_TOKEN` | Basic auth password or API token. | false | - |

See Grafana's [Authentication API
docs](https://grafana.com/docs/grafana/latest/http_api/auth/) for more info.

### Grafana Cloud Prometheus
To interact with Grafana Cloud Prometheus, you must have the `cortextool` binary
available on your path (download it [here](https://github.com/grafana/cortex-tools/releases)),
and must have these environment variables set:

| Name | Description | Required |
| --- | --- | --- |
| `PROMETHEUS_ADDRESS` | URL for Grafana Cloud Prometheus instance | true |
| `PROMETHEUS_TENANT_ID` | Tenant ID for your Grafana Cloud Prometheus account | true |
| `PROMETHEUS_TOKEN` | Authentication token/api key | true |

### Grafana Synthetic Monitoring
To interact with Grafana Synthetic Monitoring, you must have these environment variable set:

| Name | Description | Required |
| --- | --- | --- |
| `GRAFANA_SM_TOKEN` | Authentication token/api key (must have MetricsPublisher permissions) | true |
| `GRAFANA_SM_STACK_ID` | Grafana instance/stack ID | true |
| `GRAFANA_SM_LOGS_ID` | Metrics instance ID | true |
| `GRAFANA_SM_METRICS_ID` | Logs instance ID | true |

Your stack ID is the number at the end of the url when you view your Grafana instance details, ie. `grafana.com/orgs/myorg/stacks/123456` would be `123456`. Your metrics and logs ID's are the `User` when you view your Prometheus or Loki instance details in Grafana Cloud.

## Commands

### grr get
Retrieves a resource from the remote system, via its UID. Its UID will be two parts separated by a dot, `<resource-type>.<resource-id>`. A dashboard might be `dashboard.mydash`:

```sh
$ grr get Dashboard.my-uid
```

### grr list
List all resources found after executing Jsonnet file.
```sh
$ grr list my-lib.libsonnet
```

### grr show
Shows the resources found after executing Jsonnet, rendered as expected for each resource type:

```sh
$ grr show my-lib.libsonnet
```

### grr diff
Compares each resource rendered by Jsonnet with the equivalent on the remote system:

```sh
$ grr diff my-lib.libsonnet
```

### grr apply
Uploads each dashboard rendered by the mixin to Grafana
```sh
$ grr apply my-lib.libsonnet
```

### grr watch
Watches a directory for changes. When changes are identified, the
jsonnet is executed and changes are pushed to remote systems.
The directory is watched recursively (i.e. all subdirectories are watched too),
but if new subdirectories are added, watch command needs to be re-started,
as new directories will not be picked up automatically.

This example watches the current directory for changes, then executes
`my-lib.libsonnet` when changes are noticed:

```sh
$ grr watch . my-lib.libsonnet
```

### grr listen
The opposite to `watch`, when supported, this listens for changes on a remote
system. When a change is noticed, the raw resource is downloaded and saved to
a local named file.

```sh
$ grr listen dashboard.my-uid my-dash.json
```

### grr export
Renders Jsonnet and saves resources as files directory which is specified with
the second argument.

These resources can then be deployed This can be used with
[Grafana Provisioning](https://grafana.com/docs/grafana/latest/administration/provisioning/)
to provision dashboards that can be picked up immediately by Grafana.

```sh
$ grr export some-mixin.libsonnet my-provisioning-dir
```

### grr preview
When a backend supports preview functionality, this renders Jsonnet and
uploads previews to endpoint systems.

At present, only Grafana dashboards are supported. With Grafana, it produces
dashboard snapshots. It then prints out links for each snapshot that was uploaded.

```sh
$ grr preview my-lib.libsonnet
```
Grafana snapshots by default do not expire. Expiration can be set via the
`-e, --expires` flag which takes a number of seconds as an argument.

## Flags

### `-t, --target strings`

It allows the targeting of resources by key, where key is in the form `<type>/<uid>`.

Run `grr list` to get a list of resource keys in your code.

### `-J, --jpath`

It allows the targeting folder containing jsonnet library to include, should be repeated multiple times.

If not specified it include `vendor`, `lib` and local dir (`.`) folders by default.

## Grafana Dashboard Example

Create a file, called `mydash.libsonnet`, that contains this:

```jsonnet
{

  grafanaDashboardFolder:: 'my-folder', // optional

  grafanaDashboards+:: {
    'my-dash.json': {
      uid: 'prod-overview',
      title: 'Production Overview',
      tags: ['templated'],
      timezone: 'browser',
      schemaVersion: 16,
    },
  },
}
```

> **Note that this dashboard has a UID. Dashboard UIDs are required for `grr` to function correctly.**

This file follows the standard Monitoring Mixin pattern, where resources are added
to hidden maps at the root of the JSON output.

Now, we can see this rendered as a JSON dashboard with:

```sh
$ grr show mydash.libsonnet
```

Next, we need to interact with Grafana. You will need to export two environment variables: `GRAFANA_TOKEN`, an API token created within Grafana itself, and `GRAFANA_HOST`, the hostname for your Grafana installation. If the Grafana instance is not hosted at the root of the domain, you can add specify a hostname such as `admin.example.com/grafana`. If your Grafana does not use SSL, you can set `GRAFANA_PROTOCOL` to `http`.

Now, let's push our dashboard to Grafana:

```sh
$ grr apply mydash.libsonnet
```

At this point, you should find your dashboard has been uploaded and
will be visible within the Grafana UI.

As a next step, we shall make a change to our dashboard. In your `mydash.libsonnet`, change the title to `Production Overview mk2`.

Then:

```sh
$ grr diff mydash.libsonnet
```
will show us the single change we have made to our dashboard.

And next:

```sh
$ grr apply mydash.libsonnet
```
will push our change up to Grafana. Check your Grafana dashboard list - you should see your dashboard show with its new name.

You can retrieve a dashboard from Grafana via its UID:

```sh
$ grr get Dashboard.prod-overview
```
