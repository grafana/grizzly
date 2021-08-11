---
date: "2021-06-28T00:00:00+00:00"
title: "Hidden Elements"
---

> **NOTE**: This functionality is deprecated. It was necessary to support the
formats within [Monitoring Mixins](https://github.com/monitoring-mixins/docs).
However, [k8s-style](/k8s-style/) resources provide a cleaner and language
agnostic way to describe resources and their metadata.
 
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

