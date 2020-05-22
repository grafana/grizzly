# Grizzy: A Grafana Dashboard Manager

A utility for managing Jsonnet dashboards - uses the Grafana REST API.

It is designed to work with [monitoring mixins](https://github.com/monitoring-mixins/docs). It expects dashboards to be exposed by a libsonnet file as a grafanaDashboards object.

**Status: Alpha**. This is a proof of concept. It will have many holes. PRs welcome.

## Authentication and Configuration

This tool interacts with Grafana via its REST API. For this, you will need to
establish authentication credentials. These are provided to `grr` via
environment variables.

| Name | Description | Required | Default |
| --- | --- | --- | --- |
| GRAFANA\_URL | Fully qualified domain name of your Grafana instance. | true | - |
| GRAFANA\_USER | Basic auth username if applicable. | false | `api_key` |
| GRAFANA\_TOKEN | Basic auth password or API token. | false | - |

See Grafana's [Authentication API
docs](https://grafana.com/docs/grafana/latest/http_api/auth/) for more info.

## Commands

### grr get
Retrieves a dashboard from Grafana, via its UID:
```sh
$ grr get my-uid
```

### grr list
List dashboard keys from file.
```sh
$ grr list some-mixin.libsonnet
```

### grr show
Shows the mixin rendered as json.
```sh
$ grr show some-mixin.libsonnet
```

### grr diff
Compares each dashboard rendered by the mixin with the dashboard in Grafana
```sh
$ grr diff some-mixin.libsonnet
```

### grr apply
Uploads each dashboard rendered by the mixin to Grafana
```sh
$ grr apply some-mixin.libsonnet
```

## Flags

### `-t, --target strings`

The `show`, `diff`, and `apply` commands accept this flag. It allows the
targeting of dashboards by key. This can be useful if there are many dashboards
configured in the Jsonnet file you are working with. Run `grr list` to get a
list of the dashboard keys.

## Example

Create a file, called `mydash.libsonnet`, that contains this:

```jsonnet
{

  grafanaDashboardFolder: 'my-folder', // optional

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

**Note that this dashboard has a UID. UIDs are required for `grr` to function correctly.**

This file follows the standard Monitoring Mixin pattern as defined within the `prometheus-ksonnet` library.

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
$ grr get prod-overview
```
