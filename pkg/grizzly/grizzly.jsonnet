local main = import '%s';
local convert(main, apiVersion) = {
  local makeResource(kind, name, spec=null, data=null, metadata={}) = {
    apiVersion: apiVersion,
    kind: kind,
    metadata: {
      name: std.strReplace(std.strReplace(std.strReplace(name, '.json', ''), '.yaml', ''), '.yml', ''),
    } + metadata,
    [if spec != null then 'spec']: spec,
    [if data != null then 'data']: std.manifestJsonEx(data, '  '),
  },

  grafana: {
    dashboards:
      local fromMap(dashboards, folder) = [
        makeResource('Dashboard', k, data=dashboards[k], metadata={ folder: folder })
        for k in std.objectFields(dashboards)
      ];
      local folder = if 'grafanaDashboardFolder' in main then main.grafanaDashboardFolder else 'General';
      if 'grafanaDashboards' in main
      then fromMap(main.grafanaDashboards, folder)
      else {},

    datasources:
      local fromMap(datasources) = [
        makeResource('Datasource', k, spec=datasources[k])
        for k in std.objectFields(datasources)
      ];
      if 'grafanaDatasources' in main
      then fromMap(main.grafanaDatasources)
      else {},
  },

  prometheus:
    local fromMap(name) =
      if name in main
      then [makeResource('PrometheusRuleGroup', k, spec={ groups: main[name] }) for k in std.objectFields(main[name])]
      else [];
    fromMap('prometheusRules')
    + fromMap('prometheusAlerts'),

  syntheticMonitoringChecks:
    local fromMap(checks) = [
      makeResource('SyntheticMonitoringCheck', k, data=checks[k])
      for k in std.objectFields(checks)
    ];
    if 'syntheticMonitoring' in main
    then fromMap(main.syntheticMonitoring)
    else {},
};
convert(main, 'grizzly.grafana.com/v1alpha1') + main
