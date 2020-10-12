{
  grafanaDashboardFolder: 'grizzly',
  grafanaDashboards+:: {
    'my-dash.json': {
      uid: 'prod-overview',
      title: 'Production Overview',
      folderName: 'sample',
      tags: ['templated'],
      timezone: 'browser',
      schemaVersion: 17,
    },
  },
}
