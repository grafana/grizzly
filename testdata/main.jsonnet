local dashboard = import 'dashboard-simple.libsonnet';
local datasource = import 'datasource-prometheus.libsonnet';
local sm = import 'synthetic-monitoring-simple.libsonnet';

dashboard + datasource + sm {}
