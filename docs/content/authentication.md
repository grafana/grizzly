---
date: "2021-06-28T00:00:00+00:00"
title: "Authentication"
---

# Authentication and Configuration

Each system handles authentication differently. Authentication must be configured
if resources are provided for a system.

## Grafana Itself
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

## Grafana Cloud Prometheus
To interact with Grafana Cloud Prometheus, you must have these environment variables set:

| Name | Description | Required |
| --- | --- | --- |
| `CORTEX_ADDRESS` | URL for Grafana Cloud Prometheus instance | true |
| `CORTEX_TENANT_ID` | Tenant ID for your Grafana Cloud Prometheus account | true |
| `CORTEX_API_KEY` | Authentication token/api key | true |

Note, this will also work with other Cortex installations, alongside Grafana Cloud Prometheus.

## Grafana Synthetic Monitoring
To interact with Grafana Synthetic Monitoring, you must have these environment variable set:

| Name | Description | Required |
| --- | --- | --- |
| `GRAFANA_SM_TOKEN` | Authentication token/api key (must have MetricsPublisher permissions) | true |
| `GRAFANA_SM_STACK_ID` | Grafana instance/stack ID | true |
| `GRAFANA_SM_LOGS_ID` | Logs instance ID | true |
| `GRAFANA_SM_METRICS_ID` | Metrics instance ID | true |

Your stack ID is the number at the end of the url when you view your Grafana instance details, ie. `grafana.com/orgs/myorg/stacks/123456` would be `123456`. Your metrics and logs ID's are the `User` when you view your Prometheus or Loki instance details in Grafana Cloud.

# Timeouts

Grizzly has a 10 second timeout on some HTTP calls. To override this behavior, use the `GRIZZLY_HTTP_TIMEOUT=<seconds>` environment variable.

## HTTP PROXY
To use a proxy with Grizzly, you must have the following environment variable set:

| Name | Description | Required |
| --- | --- | --- |
| `HTTPS_PROXY` | This should be the full url/port of your proxy https://proxy:8080 | true |

