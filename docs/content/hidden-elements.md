---
date: "2022-04-17T00:00:00+00:00"
title: "Hidden Elements"
---

> **NOTE**: This functionality is deprecated. It was necessary to support the
formats within [Monitoring Mixins](https://github.com/monitoring-mixins/docs).
However, using explicit resources provides a cleaner and language
agnostic way to describe resources and their metadata. If you have existing
libaries that use hidden elements, it isn't too hard to write small functions
to wrap these into resources.

When working in Jsonnet, Grizzly can detect resources within standardised JSON
elements, following a convention of using hidden (`::`) elements for these.

For example, to specify a simple dashboard:

```
{
  grafanaDashboardFolder:: 'sample',
  grafanaDashboards+:: {
    'my-dash.json': {
      uid: 'prod-overview',
      title: 'Production Overview',
      tags: ['templated'],
      timezone: 'browser',
      schemaVersion: 17,
    },
  },
}
```

The `grafanaDashboardFolder` hidden element specifies the folder that all dashboards
will go into. (Using [k8s-style](/k8s-style/) resources, dashboards can be placed
individually into folders).

Within `grafanaDashboards`, each dashboard is added to a map with a key. This key
is ignored by Grizzly (it has previously been used as a filename when using
Grafana's file-based provisioning, which Grizzly does not use).

Prometheus alerts and recording rules can be defined too, for example:

```
{
  grizzly_alert:: {
    alert: 'PromScrapeFailed',
    expr: 'up != 1',
    'for': '1m',
    labels: {
      severity: 'critical',
    },
    annotations: {
      message: 'Prometheus failed to scrape a target {{ $labels.job }}  / {{ $labels.instance }}',
    },
  },

  prometheusAlerts+:: {
    grizzly_rules: {
      groups: [{
        name: 'grizzly_alert_rules',
        rules: [
          $.grizzly_alert,
        ],
      }],
    },
  },

  grizzly_record:: {
    record: 'job:up:sum',
    expr: 'sum by(job) (up)',
  },

  prometheusRules+:: {
    grizzly_rules: {
      groups: [{
        name: 'grizzly_recording_rules',
        rules: [
          $.grizzly_record,
        ],
      }],
    },
  },
}
```
Here, we first create an alert rule element (`grizzly_alert`), which we then add to namespace `grizzly_rules` inside `prometheusAlerts`. `PrometheusAlerts` is where Grizzly expects to find alerts for Prometheus configured when using 'hidden
elements'.

Likewise, we then create a recording rule element (`grizzly_record`), which we then add to
namespace `grizzly_rules` inside `prometheusRules`, which is where Grizzly expects to find Prometheus recording rules configured, again
when using hidden elements.

It also possible to load rules without specifying a namespace at all. In that case `grizzly_rules` namespace would be used. This is useful when loading alerts from monitoring-mixins:

```
{
  grizzly_alert:: {
    alert: 'PromScrapeFailed',
    expr: 'up != 1',
    'for': '1m',
    labels: {
      severity: 'critical',
    },
    annotations: {
      message: 'Prometheus failed to scrape a target {{ $labels.job }}  / {{ $labels.instance }}',
    },
  },

  prometheusAlerts+:: {
      // namespace removed
      groups: [{
        name: 'grizzly_alert_rules',
        rules: [
          $.grizzly_alert,
        ],
      }],
  },

  grizzly_record:: {
    record: 'job:up:sum',
    expr: 'sum by(job) (up)',
  },

  prometheusRules+:: {
      // namespace removed
      groups: [{
        name: 'grizzly_recording_rules',
        rules: [
          $.grizzly_record,
        ],
      }],
  },
}
```
