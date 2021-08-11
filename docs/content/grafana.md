---
date: "2021-06-28T00:00:00+00:00"
title: "With Grafana"
---

## Dashboards
In [What is Grizzly?](../what-is-grizzly/) we saw an example of how to manage
a Grafana dashboard.

When representing a dashboard, the JSON that is downloaded from Grafana should
be placed into the `spec` element. If using YAML, the JSON should be converted
to YAML before doing so.

## Folders
Grafana dashboard folders are probably the simplest resources you can manage
with Grizzly:

```
apiVersion: grizzly.grafana.com/v1alpha1
kind: DashboardFolder
metadata:
  name: sample
spec:
  title: Special Sample Folder
```

A folder simply has a name and a title.

### Placing Dashboards in Folders
Dashboards can be placed into folders using the `folder` metadata field. Here, a
dashboard is placed into the folder defined above:

```
apiVersion: grizzly.grafana.com/v1alpha1
kind: Dashboard
metadata:
    folder: sample
    name: prod-overview
spec:
    schemaVersion: 17
    tags:
        - templated
    timezone: browser
    title: Production Overview
    uid: prod-overview
```
> **Note:** The 'general' folder is a special case, and can be assumed to exist.
> You cannot manage it directly with Grizzly. However, you can place dashboards
> in the General folder simply by specifying `folder: general` in the metadata
> for the dashboard.

## Datasources
To describe a Grafana datasource, use something like the following:

```
apiVersion: grizzly.grafana.com/v1alpha1
kind: Datasource
metadata:
    name: prometheus
spec:
    access: proxy
    url: http://localhost/prometheus/
```
