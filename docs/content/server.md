---
date: 2024-03-14
title: Grizzly Server
---
## An HTTP Server for editing and reviewing

Do you have Grafana resources, dashboards, datasources, alerts, etc, on disk,
and want a way to edit or review them before publishing?

Grizzly Server is here to help.

With a Grafana instance configured (as described in the [configuration section](../configuration)),
you can run Grizzly against one or more local files and it will start up an
HTTP server:

```
grr serve <mydir>
```

By default, this starts an HTTP server on (http://localhost:8080)[http://localhost:8080].
Visiting this URL will show a list of the resources found in <mydir>. For
now, this is limited to Grafana Dashboards only.

Clicking on any of these dashboards will take you to your Grafana instance,
but with load/save events for your dashboards being handled by the Grizzly
server. Your dashboard won't reach Grafana itself. But Grafana will provide
the UI plus resolve any datasource requests to make your dashboards look
pretty!
