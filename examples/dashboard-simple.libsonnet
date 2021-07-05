local util = import 'util.libsonnet';

local dashboard = {
  uid: 'prod-overview',
  title: 'Production Overview',
  tags: ['templated'],
  timezone: 'browser',
  schemaVersion: 17,
};

util.makeResource('Dashboard', dashboard.uid, dashboard, { folder: 'sample' })
