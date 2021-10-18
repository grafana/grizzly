---
date: "2021-06-28T00:00:00+00:00"
title: "What is Grizzly?"
---
Grafana Grizzly is a command line tool that allows you to manage your
observability resources with code.

Now you can define your dashboards, alerting and recording rules, etc, all
from within your codebase.

You can build Grizzly into your continuous deployment pipelines, meaning
whenever you deploy your application, your observability is updated too.

## Getting Started
The simplest way to achieve this is using YAML resource
descriptions. (Programmatic description is explained in the section on 
[Jsonnet](../jsonnet/)).

We will first explore how to use Grizzly with Grafana.

## Authentication
You will need to set at least the `GRAFANA_URL` environment variable. This should
be to the root of your Grafana instance. If your Grafana instance requires authentication,
then you are recommended to acquire an API key (`Configuration` / `API keys`). Set the
`GRAFANA_TOKEN` to this value. For more this, see [authentication](../authentication/).

## Managing a Simple Dashboard

Below is an example YAML representation of a Grafana dashboard (the format is inspired by
Kubernetes):

#### **`dashboard.yaml`**:
```
apiVersion: grizzly.grafana.com/v1alpha1
kind: Dashboard
metadata:
    name: prod-overview
    folder: general
spec:
    schemaVersion: 17
    tags:
        - templated
    timezone: browser
    title: Production Overview
    uid: prod-overview
```

This has a number of components:
 * **kind**: the type of object being described
 * **apiVersion**: the specific version of that type of object
 * **metadata**: various descriptors for the resource
 * **spec**: the actual resource that will be pushed to the
   end point.

Once [authentication](/authentication/) has been configured, this
very basic dashboard can be uploaded to Grafana with this
command:

```
$ grr apply dashboard.yaml
Dashboard/prod-overview added
```

Let's change the `schemaVersion` field in `dashboard.yaml` to 18.
Then we can see what has changed, that is, what changes will be
applied to Grafana on the next `apply`:
```
$ grr diff dashboard.yaml
--- Remote
+++ Local
@@ -4,7 +4,7 @@
     folder: general
     name: prod-overview
 spec:
-    schemaVersion: 17
+    schemaVersion: 18
     tags:
         - templated
     timezone: browser
```

After this, we can apply this simple change:

```
$ grr apply dashboard.yaml
Dashboard/prod-overview updated
```

