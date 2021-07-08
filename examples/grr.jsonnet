local dashboard = import 'dashboard-simple.libsonnet';
local datasource = import 'datasource-prometheus.libsonnet';
local folder = import 'folder-simple.libsonnet';
local prometheus = import 'prometheus-rules.libsonnet';
local sm = import 'synthetic-monitoring-simple.libsonnet';

{
  folders: [folder],
  dashboards: [dashboard],
  datasources: [datasource],
  prometheus: [prometheus],
  sm: [sm],
}
