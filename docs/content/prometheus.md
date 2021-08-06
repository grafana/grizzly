---
date: "2021-06-28T00:00:00+00:00"
title: "With Prometheus"
---

## Which Prometheus' are supported?
Prometheus itself requires its configuration to be present in text files on
local disk. As Grizzly focuses on systems that can be managed via HTTP APIs,
Grizzly cannot (currently) work with Prometheus itself.

Various hosted Prometheus installations, such as Grafana Cloud Prometheus
are supported, as are systems running Cortex.

## Configuring Prometheus
Prometheus alert and recording rules are both created using the same `kind`:
`PrometheusRuleGroup`. Rule groups need to be placed into a `namespace` -
namespaces simply group these rule groups together.

## Prometheus Alerts

A sample Prometheus alert could be:
```
apiVersion: grizzly.grafana.com/v1alpha1
kind: PrometheusRuleGroup
metadata:
    name: grizzly_alerts
    namespace: grizzly_rules
spec:
    rules:
        - alert: PromScrapeFailed
          annotations:
            message: Prometheus failed to scrape a target {{ $labels.job }}  / {{ $labels.instance }}
          expr: up != 1
          for: 1m
          labels:
            severity: critical
```

## Prometheus Recording Rules

```
apiVersion: grizzly.grafana.com/v1alpha1
kind: PrometheusRuleGroup
metadata:
    name: grizzly_recording_rules
    namespace: grizzly_rules
spec:
    rules:
        - expr: sum by(job) (up)
          record: job:up:sum
```
