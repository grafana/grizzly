---
date: 2024-03-14
title: Grizzly Server
---
## An HTTP Server for editing and reviewing
When we manage Grafana dashboards on disk, the Grizzly server makes easy to
edit and review these resources within an actual Grafana instance, but without
needing to publish the dashboard to Grafana. Grafana provides the UI and datasources
to make dashboards look right, Grizzly provides the dashboard, directly from
local disk.

If the file on disk is writeable (i.e. a pure YAML or JSON file - Jsonnet cannot be written),
then clicking the `Save` icon in Grafana will update your local dashboard file. It will not
change anything in Grafana itself.

With a Grafana instance configured (as described in the [configuration section](../configuration)),
you can run Grizzly against one or more local files and it will start up an
HTTP server:

```
grr serve <mydir>
```

By default, this starts an HTTP server on [http://localhost:8080](http://localhost:8080).
Visiting this URL will show a list of the resources found in <mydir>. For
now, this is limited to Grafana Dashboards only.
