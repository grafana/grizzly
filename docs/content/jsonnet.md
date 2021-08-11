---
date: "2021-06-28T00:00:00+00:00"
title: "Using Jsonnet"
---

Jsonnet is by far the most powerful use-case for Grizzly. It allows resources
to be described programmatically, using all the common patterns programmers
are familiar with, e.g. variables, libraries and abstractions.

To understand the Jsonnet language, see the [Jsonnet tutorial](https://jsonnet.org/learning/tutorial.html).

> **Note**: The original approach to using Jsonnet in Grizzly involved hidden elements.
> For details of this deprecated approach, see [Hidden Elements](../hidden-elements/).

Here is an example of rendering Prometheus alert and recording rules with Jsonnet:

```
{
  local rule(name, namespace, spec) = {
    apiVersion: 'grizzly.grafana.com/v1alpha1',
    kind: 'PrometheusRuleGroup',
    metadata: {
      name: name,
      namespace: namespace,
    },
    spec: spec,
  },

  grizzly_alerts:: {
    rules: [{
      alert: 'PromScrapeFailed',
      expr: 'up != 1',
      'for': '1m',
      labels: {
        severity: 'critical',
      },
      annotations: {
        message: 'Prometheus failed to scrape a target {{ $labels.job }}  / {{ $labels.instance }}',
      },
    }],
  },

  grizzly_records:: {
    rules: [{
      record: 'job:up:sum',
      expr: 'sum by(job) (up)',
    }],
  },

  prometheus_rules: [
    rule('grizzly_alert_rules', 'grizzly_rules', $.grizzly_alerts),
    rule('grizzly_recording_rules', 'grizzly_rules', $.grizzly_records),
  ],
}

```
In this example, we first define a function `rule()` for creating resources. Then we define our
alert and recording rules separately. Finally, we use the `rule` function to expose these.

In Jsonnet, `::` signifies hidden, that is, elements defined with `::` won't be visible in the
output. Thus `grizzly_alerts` and `grizzly_records` are both internal to the script, and only
see the light of day because they are referenced within `prometheus_rules`.
