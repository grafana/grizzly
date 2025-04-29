---
date: "2021-06-28T00:00:00+00:00"
title: "Setup and Configuration"
---

Grizzly can be configured in two ways: using environment variables
and through Grizzly's own 'context' support (much like `kubectl` contexts).

Environment variables are suitable for use against a single system, and best suited
to CI and automation scenarios.

Grizzly contexts allow easy switching between multiple configurations and are best
suited to workstation use.

# Using Grizzly Contexts

By default, the `default` context is used. We will discuss how to use the default
context first, then explain how to use multiple contexts thereafter.

> **NOTE**: If you have used Grizzly previously with environment variables, you can
  initialise the `default` context from your environment simply with `grr config import`.
  Also note, after importing, remember to unset your environment variables, as they
  take precendence over contexts and will thus prevent access to your context configurations.

Settings can be configured via Grizzly itself, using `grr config set`. They are stored
in an OS specific location.

## Authenticate with a Grafana instance

Grizzly interacts with Grafana via its REST API. For this, you will need to
establish authentication credentials.

The minimum requirement is to set the URL of the Grafana instance to be used:
```sh
grr config set grafana.url http://localhost:3000 # URL for the root of your Grafana instance
```

Optionally, set the following field(s), depending our your [authentication method with the given Grafana instance](https://grafana.com/docs/grafana/latest/setup-grafana/configure-security/configure-authentication/):
- A [token](#token-or-password-optional) if using a [Grafana service account](https://grafana.com/docs/grafana/latest/administration/service-accounts) (recommended)
- A [username](#username-optional) and [password](#token-or-password-optional) if using basic authentication

Next, consider setting a [context](#using-grizzly-contexts) to save this configuration.

Once you have configured your authentication method, you are ready to use the [Grizzly server](./server.md) to view and/or edit resources.

### Token or password (optional)

To authenticate with a Grafana instance using a service account token, first [create a service account in the
Grafana instance](https://grafana.com/docs/grafana/latest/administration/service-accounts/#to-create-a-service-account).
Before clicking `Create` to create the service account, ensure that you have added the role that
best suits your use case. A role with `Viewer` permissions would be enough
to view resources. Use `Editor` or `Admin` to be able to edit resources.
Then,
[add a token to the new service account](https://grafana.com/docs/grafana/latest/administration/service-accounts/#add-a-token-to-a-service-account-in-grafana).
Copy the token and set it with Grizzly:

```sh
grr config set grafana.token abcd12345 # Service account token (or basic auth password)
```

Alternatively, set this field to a password if using [basic authentication](https://grafana.com/docs/grafana/latest/setup-grafana/configure-security/configure-authentication/#basic-authentication).

### Username (optional)

Optionally, set a username if using [basic authentication](https://grafana.com/docs/grafana/latest/setup-grafana/configure-security/configure-authentication/#basic-authentication).

```sh
grr config set grafana.user admin # (Optional) Username if using basic auth
```

## Authenticate with hosted Prometheus

To interact with [hosted Prometheus / Mimir](./prometheus.md) resources, use these settings:

```sh
grr config set mimir.address https://mimir.example.com # URL for Mimir instance or Grafana Cloud Prometheus instance
grr config set mimir.tenant-id myTenant # Tenant ID for your Grafana Cloud Prometheus account
grr config set mimir.api-key abcdef12345 # Authentication token (if you are using Grafana Cloud)
```

**Notes**
* Be sure to set `api-key` when you need to interact with Grafana Cloud.

## Authenticate with Grafana Synthetic Monitoring
To interact with Grafana Synthetic Monitoring, you must configure the below settings:

```sh
grr config set synthetic-monitoring.access-token abcdef123456
grr config set synthetic-monitoring.url https://synthetic-monitoring-api.grafana.net
```

You can find the URL and access token in the Synthetic Monitoring plugin's config page in Grafana.

## Configuring Targets
Grizzly supports a number of resource types (`grr providers` will list those supported). Often, however, we do not
wish to use all of these types. It is possible to set a list of "target" resource types that Grizzly should interact
with:

```
grr config set targets Dashboard,DashboardFolder
```

These can be overriden on the command line with the `-t` or `--target` flag.

## Configuring Output Formats
Grizzly, when retrieving resources from Grafana, can present them in a range of formats. Currently, it supports
YAML and JSON. Default is YAML. It can be configured in contexts:

```
grr config set output-format json
```

This can be overridden on the command line with `-o` or `--output`.

Also, Grizzly wraps resources into an "envelope" that provides a consistent way of specifying typing and metadata,
following Kubernetes' lead. This envelope can be removed with the `only-spec` setting:

```
grr config set only-spec true
```

This can be overridden on the command line with `-s` (to only include the spec component) or `--only-spec=false` to
disable this setting (if currently set in the context).

# Contexts
Grizzly supports multiple contexts allowing easy swapping between instances. By default, Grizzly uses the `default`
context.

[Optional] If you have environment variables configured, you can import these into the `default` context like so:

```sh
grr config import
```

Create a new context with:
```sh
grr config create-context production
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

To see the config for a context, these options are available:
```sh
grr config get
grr config get grafana
grr config get grafana.url
```

After selecting a different context, all future `grr` invocations will use the credentials and settings in this
new context, whether `grr apply` to apply resources or `grr config set` to set configuration values.

# Configuring Grizzly with environment variables

In some circumstances (e.g. when used within automated pipelines) it makes sense to configure Grizzly directly
with environment variables as opposed to contexts. Environment variables, when set, take precedence over
Grizzly contexts as described above. Below are the variables that can be used for this.

| Name            | Description                                           | Required | Default   |
|-----------------|-------------------------------------------------------|----------|-----------|
| `GRAFANA_URL`   | Fully qualified domain name of your Grafana instance. | true     | -         |
| `GRAFANA_USER`  | Basic auth username if applicable.                    | false    | `api_key` |
| `GRAFANA_TOKEN` | Basic auth password or API token.                     | false    | -         |

See Grafana's [Authentication API
docs](https://grafana.com/docs/grafana/latest/http_api/auth/) for more info.

## Grafana Cloud Prometheus
To interact with Grafana Cloud Prometheus, you must have these environment variables set:

| Name               | Description                                         | Required |
|--------------------|-----------------------------------------------------|----------|
| `MIMIR_ADDRESS`    | URL for Grafana Cloud Prometheus instance           | true     |
| `MIMIR_TENANT_ID`  | Tenant ID for your Grafana Cloud Prometheus account | true     |
| `MIMIR_API_KEY`    | Authentication token/api key                        | false    |
| `MIMIR_AUTH_TOKEN` | Authorization Bearer Token                          | false    |

Note, this will also work with other Mimir installations, alongside Grafana Cloud Prometheus.

## Grafana Synthetic Monitoring
To interact with Grafana Synthetic Monitoring, you must have these environment variable set:

| Name                      | Description                                                           | Required |
|---------------------------|-----------------------------------------------------------------------|----------|
| `GRAFANA_SM_ACCESS_TOKEN` | Authentication token/api key (must have MetricsPublisher permissions) | true     |
| `GRAFANA_SM_STACK_ID`     | Grafana instance/stack ID                                             | true     |
| `GRAFANA_SM_LOGS_ID`      | Logs instance ID                                                      | true     |
| `GRAFANA_SM_METRICS_ID`   | Metrics instance ID                                                   | true     |
| `GRAFANA_SM_URL`          | Synthetic Monitoring instance URL                                     | true     |

Your stack ID is the number at the end of the url when you view your Grafana instance details, ie. `grafana.com/orgs/myorg/stacks/123456` would be `123456`. Your metrics and logs ID's are the `User` when you view your Prometheus or Loki instance details in Grafana Cloud.
You can find your instance URL under your Synthetic Monitoring configuration.

# Grizzly configuration file
To get the path of the config file:
```sh
grr config path
```

# Other Configurations

## Timeouts

Grizzly has a 10 second timeout on some HTTP calls. To override this behavior, use the `GRIZZLY_HTTP_TIMEOUT=<seconds>` environment variable.

## HTTP PROXY
To use a proxy with Grizzly, you must have the following environment variable set:

| Name | Description | Required |
| --- | --- | --- |
| `HTTPS_PROXY` | This should be the full url/port of your proxy https://proxy:8080 | true |
