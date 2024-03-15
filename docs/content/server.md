---
date: 2024-03-14
title: Grizzly Server
---
## An HTTP Server for editing and reviewing
When we manage Grafana dashboards on disk, the Grizzly server makes easy to
edit and review these resources within an actual Grafana instance, but without
needing to publish the dashboard to Grafana. Grafana provides the UI and datasources
to make dashboards look right, Grizzly provides the dashboard, directly from
local disk. Clicking save will update your local dashboard file, without interacting
with Grafana itself.

With a Grafana instance configured (as described in the [configuration section](../configuration)),
you can run Grizzly against one or more local files and it will start up an
HTTP server:

```
grr serve <mydir>
```

By default, this starts an HTTP server on (http://localhost:8080)[http://localhost:8080].
Visiting this URL will show a list of the resources found in <mydir>. For
now, this is limited to Grafana Dashboards only.
