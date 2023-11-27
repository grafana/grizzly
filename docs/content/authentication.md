---
date: "2021-06-28T00:00:00+00:00"
title: "Authentication"
---

# Authentication and Configuration

## Introducing Contexts
Grizzly supports 'contexts' (much like `kubectl` contexts), allowing users to maintain
configuration for multiple Grafana (and Mimir and Synthetic Monitoring) systems.

By default, the `default` context is used. We will discuss how to use the default
context first, then explain how to use multiple contexts thereafter.

## Settings

> **NOTE**: If you have used Grizzly previously with environment variables, you can
  initialise the `default` context from your environment simply with `grr config import`.

Settings can be configured via Grizzly itself, using `grr config set`. They are stored
in an OS specific location.

## Grafana Itself

This tool interacts with Grafana via its REST API. For this, you will need to
establish authentication credentials.

```sh
grr config set grafana.url http://localhost:3000 # URL for the root of your Grafana instance
grr config set grafana.user admin # Optional: Username if using basic auth
grr config set grafana.token abcd12345 # Service account token (or basic auth password)
```

## Grafana Cloud Prometheus
To interact with Grafana Cloud Prometheus (aka Mimir), use these settings:

```sh
grr config set mimir.address https://mimir.example.com # URL for Grafana Cloud Prometheus instance
grr config set mimir.tenant-id 1234567 # Tenant ID for your Grafana Cloud Prometheus account
grr config set mimir.api-key abcdef12345 # Authentication token
```

Note, this will also work with other Cortex installations, alongside Grafana Cloud Prometheus/Mimir.

## Grafana Synthetic Monitoring
To interact with Grafana Synthetic Monitoring, you must configure the below settings:

```sh
grr config set synthetic-monitoring token abcdef123456 # API key (must have MetricsPublisher permissions)
grr config set synthetic-monitoring stack-id # Grafana stack ID
grr config set synthetic-monitoring metrics-id # Logs instance ID
grr config set synthetic-monitoring logs-id # Metrics instance ID
```
Your stack ID is the number at the end of the url when you view your Grafana instance details, ie. `grafana.com/orgs/myorg/stacks/123456` would be `123456`. Your metrics and logs ID's are the `User` when you view your Prometheus or Loki instance details in Grafana Cloud.

# Contexts
Grizzly supports multiple contexts allowing easy swapping between instances. By default, Grizzly uses the `default`
context.

[Optional] If you have environment variables configured, you can import these into the `default` context like so:

```sh
grr config import
```

Create a new context with:
```sh
grr config create production
```

To list existing contexts:
```sh
grr config get-contexts
```

To show the currently selected context:
```sh
grr config current-context
```

To switch to a different context:
```sh
grr config use-context staging
```

After selecting a different context, all future `grr` invocations will use the credentials and settings in this
new context, whether `grr apply` to apply resources or `grr config set` to set configuration values.

# Other Configurations

## Timeouts

Grizzly has a 10 second timeout on some HTTP calls. To override this behavior, use the `GRIZZLY_HTTP_TIMEOUT=<seconds>` environment variable.

## HTTP PROXY
To use a proxy with Grizzly, you must have the following environment variable set:

| Name | Description | Required |
| --- | --- | --- |
| `HTTPS_PROXY` | This should be the full url/port of your proxy https://proxy:8080 | true |
