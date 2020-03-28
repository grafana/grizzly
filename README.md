# Grafana Dash: A Dashboard Manager

A simple utility for managing Jsonnet dashboards - uses the Grafana REST API.

It is designed to work with [monitoring mixins](https://github.com/monitoring-mixins/docs). It expects dashboards to be exposed by a libsonnet file as a grafanaDashboards object.

**Status: Alpha**. This is a proof of concept. It will have many holes. PRs welcome.

## Building
Makefile needs sorting out. Build with `go build cmd/g/g.go`.

## Authentication and Configuration
This tool interacts with Grafana via its REST API. For this, you will need to establish authentication
credentials. These are provided to `dash` via environment variables.

You can either provide a `GRAFANA_URL`, which can include authentication details, or you scan specify separately:

* `GRAFANA_PROTOCOL`: Defaults to `https`. You can set this to `http` if you Grafana instance does not use SSL.
* `GRAFANA_USER`: Either basic auth username, or `apikey` if using an API token.
* `GRAFANA_TOKEN`: Either an API token, or a password for Basic auth. To retrieve a token in Grafana, select the configuration option on the left (the cog), then 'API Keys'. From there you should be able add a new key, with 'Editor' permissions.
* `GRAFANA_HOST`: The hostname for your Grafana installation.
* `GRAFANA_PATH`: If the Grafana instance is not hosted at the root of the domain, you can add specify a path such as `grafana`. Do not specify the initial slash.

## Basic Commands
### dash get
Retrieves a dashboard from Grafana, via its UID:
```sh
$ dash get my-uid
```

### dash show
Shows the mixin rendered as json.
```sh
$ dash show some-mixin.libsonnet
```

### dash diff
Compares each dashboard rendered by the mixin with the dashboard in Grafana
```sh
$ dash diff some-mixin.libsonnet
```

### dash apply
Uploads each dashboard rendered by the mixin to Grafana
```sh
$ dash apply some-mixin.libsonnet
```

## Example

Create a file, called `mydash.libsonnet`, that contains this:

```jsonnet
{
  grafanaDashboards+:: {
     "my-dash.json": {
        "uid": "prod-overview",
        "title": "Production Overview",
        "tags": [ "templated" ],
        "timezone": "browser",
        "schemaVersion": 16,
     },
  },
}
```

**Note that this dashboard has a UID. UIDs are required for `dash` to function correctly.**

This file follows the standard Monitoring Mixin pattern as defined within the `prometheus-ksonnet` library.

Now, we can see this rendered as a JSON dashboard with:

```sh
$ dash show mydash.libsonnet
```

Next, we need to interact with Grafana. You will need to export two environment variables: `GRAFANA_TOKEN`, an API token created within Grafana itself, and `GRAFANA_HOST`, the hostname for your Grafana installation. If the Grafana instance is not hosted at the root of the domain, you can add specify a hostname such as `admin.example.com/grafana`. If your Grafana does not use SSL, you can set `GRAFANA_PROTOCOL` to `http`.

Now, let's push our dashboard to Grafana:

```sh
$ dash apply mydash.libsonnet
```

At this point, you should find your dashboard has been uploaded and
will be visible within the Grafana UI.

As a next step, we shall make a change to our dashboard. In your `mydash.libsonnet`, change the title to `Production Overview mk2`. Your jsonnet will now look like this:

```jsonnet
{
  grafanaDashboards+:: {
    "my-dash.json": {
        "uid": "prod-overview",
        "title": "Production Overview mk2",
        "tags": [ "templated" ],
        "timezone": "browser",
        "schemaVersion": 16,
    },
  },
}
```

Then:

```sh
$ dash diff mydash.libsonnet
```
will show us the single change we have made to our dashboard.

And next:

```sh
$ dash apply mydash.libsonnet
```
will push our change up to Grafana. Check your Grafana dashboard list - you should see your dashboard show with its new name.

You can retrieve a dashboard from Grafana via its UID:

```sh
$ dash get prod-overview
```
