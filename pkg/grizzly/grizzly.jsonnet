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

    folders:
      local is_alpha(x) =
    std.member("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_", x);
      local uid(folder) = std.join("", std.filter(is_alpha, std.stringChars(folder)));
     if ('grafanaDashboardFolder' in main) && main.grafanaDashboardFolder != 'General'
      then makeResource(
        'DashboardFolder',
        uid(main.grafanaDashboardFolder),
        spec={
          title: main.grafanaDashboardFolder,
        }),
    dashboards:
      local uid(k, dashboard) =
        if std.objectHasAll(dashboard, "uid")
        then dashboard.uid
        else k;
      local folder =
        if 'grafanaDashboardFolder' in main
        then main.grafanaDashboardFolder
        else 'General';
      local fromMap(dashboards, folder) = [
        makeResource(
          'Dashboard',
          uid(k, dashboards[k]),
          spec=dashboards[k] + {
            uid::'',
          },
          metadata={ folder: folder }
        )
        for k in std.objectFields(dashboards)
      ];
      if 'grafanaDashboards' in main
      then fromMap(main.grafanaDashboards, folder)
      else {},

    datasources:
      local fromMap(datasources) = [
        makeResource(
          'Datasource',
          k,
          spec=datasources[k] + {
            name:: ''
          },
        )
        for k in std.objectFields(datasources)
      ];
      if 'grafanaDatasources' in main
      then fromMap(main.grafanaDatasources)
      else {},
  },

  prometheus:
    local groupNamespace(key) = std.objectFields(main[key])[0];
    local groupName(key) = main[key][groupNamespace(key)].groups[0].name;
    local fromMap(key) =
      if key in main
      then [
              makeResource(
                'PrometheusRuleGroup',
                groupName(key),
                spec={
                  rules: main[key][groupNamespace(key)].groups[0].rules,
                },
                metadata={ namespace: groupNamespace(key)})
              for k in std.objectFields(main[key])
           ]
      else [];
    fromMap('prometheusRules')
    + fromMap('prometheusAlerts'),

  syntheticMonitoringChecks:
    local fromMap(checks) = [
      makeResource(
        'SyntheticMonitoringCheck',
        checks[k].job,
        spec=checks[k] + {
          job::'',
        },
        metadata={type: std.objectFields(checks[k].settings)[0]}
      )
      for k in std.objectFields(checks)
    ];
    if 'syntheticMonitoring' in main
    then fromMap(main.syntheticMonitoring)
    else {},
};
convert(main, 'grizzly.grafana.com/v1alpha1') + main
