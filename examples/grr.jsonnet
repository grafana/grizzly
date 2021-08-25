local dashboard = import 'dashboard-simple.libsonnet';
local datasource = import 'datasource-prometheus.libsonnet';
local folder = import 'folder-simple.libsonnet';
local grr = import 'grizzly/grizzly.libsonnet';
local prometheus = import 'prometheus-rules.libsonnet';
local sm = import 'synthetic-monitoring-simple.libsonnet';

{
  folders: [
    grr.folder.new('sample', 'Sample'),
  ],
  dashboards: [
    grr.dashboard.new('prod-overview', dashboard)
    + grr.resource.addMetadata('folder', 'sample'),
  ],
  datasources: [
    grr.datasource.new('prometheus', datasource),
  ],
  prometheus_rule_groups: [
    grr.rule_group.new('first_rules', 'grizzly_alerts', prometheus),
  ],
  sm: [
    grr.synthetic_monitoring_check.new('http', 'grafana-com', sm),
  ],
}
