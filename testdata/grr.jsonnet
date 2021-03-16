local dashboard = import 'dashboard-simple.libsonnet';
local datasource = import 'datasource-prometheus.libsonnet';
local prometheus = import 'prometheus-rules.libsonnet';
local sm = import 'synthetic-monitoring-simple.libsonnet';

{
  dashboards: [dashboard],
  datasources: [datasource],
  prometheus: [prometheus],
  sm: [sm],
}
