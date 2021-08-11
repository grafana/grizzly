---
date: "2021-06-28T00:00:00+00:00"
title: "Synthetic Monitoring"
---
## Grafana Cloud Synthetic Monitoring Checks
The only resource type currently configurable by Grizzly is a "check", which
tells Synthetic Monitoring to periodically check the status of an endpoint.

A synthetic monitoring check requires a name and a type to be specified within
its metadata. The type is required as metadata, when it is also included in
the spec because it is possible to reuse the same name across checks of
different types. The type is therefore required as metadata to make the check
UID unique.

```
apiVersion: grizzly.grafana.com/v1alpha1
kind: SyntheticMonitoringCheck
metadata:
    name: grafana-com
    type: http
spec:
    alertSensitivity: ""
    basicMetricsOnly: true
    enabled: true
    frequency: 60000
    job: grafana-com
    labels: []
    offset: 0
    probes:
        - Atlanta
        - Chicago
        - LosAngeles
        - Miami
        - Seattle
        - SanJose
        - Paris
        - Tokyo
        - Seol
        - NewYork
        - SanFrancisco
        - Amsterdam
        - Singapore
        - Frankfurt
        - Bangalore
        - Dallas
        - Newark
        - Toronto
        - London
        - Mumbai
        - Sydney
    settings:
        http:
            failIfNotSSL: false
            failIfSSL: false
            ipVersion: V4
            method: GET
            noFollowRedirects: false
    target: https://google.com/
    timeout: 2500
```

### Working with Probes
The HTTP API for Synthetic Monitoring requires numeric probe IDs to be
specified under the `probes` element. Working with numerical IDs is not
easy, so as a convenience for the user, Grizzly first calls the `probes`
API within Synthetic Monitoring and converts names to numerical IDs, or
visa versa.
