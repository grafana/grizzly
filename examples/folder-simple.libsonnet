local util = import 'util.libsonnet';

local folder = {
  title: 'Sample',
};

util.makeResource('DashboardFolder', 'sample', folder, {})
