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

## Library Elements

Library Elements (currently Panels and Variables) are structured like this:

```yaml
apiVersion: grizzly.grafana.com/v1alpha1
kind: LibraryElement
metadata:
    name: bcb6ec8a-8a64-4fa6-a490-68b8e73bee16
spec:
    kind: 1 # 1 = Panel, 2 = Variable
    model:
        datasource:
            type: prometheus
            uid: grafanacloud-prom
        description: ""
        gridPos:
            h: 8
            w: 12
            x: 0
            "y": 0
        id: 1
        options:
            code:
                language: plaintext
                showLineNumbers: false
                showMiniMap: false
            content: Example Content
            mode: markdown
        title: Example Panel
        type: text
    name: Example Panel
    orgId: 1
    type: text
```

## AlertRuleGroup

AlertRuleGroups are sets of rules evaluated at the same interval.
The easiest way to build alert rules is using the "Modify export" functionality in Grafana.

The resulting resource looks like this:

```yaml
apiVersion: grizzly.grafana.com/v1alpha1
kind: AlertRuleGroup
metadata:
    name: fee4037a-b193-4e28-9330-2cc9028b048c.d
spec:
    folderUid: fee4037a-b193-4e28-9330-2cc9028b048c
    interval: 180 # in seconds
    rules:
        - condition: B
          data:
            - datasourceUid: grafanacloud-demoinfra-prom
              model:
                datasource:
                    type: prometheus
                    uid: grafanacloud-demoinfra-prom
                editorMode: code
                expr: weather_temp_c{location="Vienna"}
                instant: true
                intervalMs: 1000
                legendFormat: __auto
                maxDataPoints: 43200
                range: false
                refId: A
              refId: A
              relativeTimeRange:
                from: 600
            - datasourceUid: __expr__
              model:
                conditions:
                    - evaluator:
                        params:
                            - 30
                        type: gt
                      operator:
                        type: and
                      query:
                        params:
                            - C
                      reducer:
                        params: []
                        type: last
                      type: query
                datasource:
                    type: __expr__
                    uid: __expr__
                expression: A
                intervalMs: 1000
                maxDataPoints: 43200
                refId: B
                type: threshold
              refId: B
              relativeTimeRange:
                from: 600
          execErrState: Error
          folderUID: fee4037a-b193-4e28-9330-2cc9028b048c
          for: 5m0s
          id: 3
          noDataState: NoData
          orgID: 1
          ruleGroup: d
          title: Temperature high
          uid: d4231da1-2456-4741-8a81-527167a96b69
    title: Alert Group Europe
```

## Contact Points

To provision contact points, use the following structure:

```yaml
apiVersion: grizzly.grafana.com/v1alpha1
kind: AlertContactPoint
metadata:
    name: eaae236a-7be9-4748-a08e-54b92ffb2e60
spec:
    name: grafana-default-email
    settings:
        addresses: <example@email.com>
        singleEmail: false
    type: email
    uid: eaae236a-7be9-4748-a08e-54b92ffb2e60
```

If the contact point contains credentials, grizzly will always report a change
as Grafana will not expose the credentials via the API.

## Notification Policy

As the Notification Policy is stored as a single resource in Grafana, you can
only have a single `AlertNotificationPolicy` specified. It needs to have the
name `global`:


```yaml
apiVersion: grizzly.grafana.com/v1alpha1
kind: AlertNotificationPolicy
metadata:
    name: global
spec:
    group_by:
        - grafana_folder
        - alertname
    receiver: grafana-default-email
    routes:
        - group_by:
            - region
          object_matchers:
            - - foo
              - =
              - bar
          receiver: grafana-oncall
```
