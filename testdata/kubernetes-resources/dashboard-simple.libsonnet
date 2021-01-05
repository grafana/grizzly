local util = import 'util.libsonnet';

util.makeResource(
  'Dashboard', 'my-dash.json', {
    uid: 'prod-overview',
    title: 'Production Overview',
    tags: ['templated'],
    timezone: 'browser',
    schemaVersion: 17,
  },
)
