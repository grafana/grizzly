---
date: "2021-06-28T00:00:00+00:00"
title: "Workflows"
---

Grizzly is a flexible tool that supports a number of workflows.

# Jsonnet
The principal workflow for Grizzly involves Jsonnet.

With this workflow, resources are described in Jsonnet. When Grizzly is invoked,
it evaluates the Jsonnet and renders the resources into JSON/YAML. This it can
then push to the relevant systems.

This is a unidirectional process. It doesn't make sense to pull Jsonnet from a
remote system.

# Pull/Push
Sometimes we want to migrate resources between systems - e.g. migrating dashboards
from one Grafana instance to another.

With the combination of `grr pull -d` and `grr apply -d` it is possible to download
all, or a selection of remote resources, and then push them up to an alternative
instance.

# Listen
Particularly when working with visual resources, it is often easier to use the
remote system to edit resources, e.g. using Grafana to edit a dashboard.

However, in this scenario, the versioning and traceability of our dashboards is
limited.

With `grr listen`, we can have Grizzly 'subscribe' to change notifications on
Grafana, and be notified when dashboards are saved. When saves happen, the JSON
for the dashboard will be written to local disk. If this happens to be within a
git checkout, for example, it then becomes easy to make changes via Grafana itself,
then push the JSON representation of those changes to Git.

