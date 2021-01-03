local util = import 'util.libsonnet';

util.makeResource(
  'grafana.com/v1', 'Dashboard', 'my-dash.json', {
    uid: 'prod-overview',
    title: 'Production Overview',
    tags: ['templated'],
    timezone: 'browser',
    schemaVersion: 17,
  },
)
