{
  grafanaDashboardFolder:: 'sample',
  grafanaDashboards+:: {
    'my-dash.json': {
      uid: 'prod-overview',
      title: 'Production Overview',
      tags: ['templated'],
      timezone: 'browser',
      schemaVersion: 17,
    },
  },
}
