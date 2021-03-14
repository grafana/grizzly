local util = import 'util.libsonnet';

local datasource = {
  access: 'proxy',
  isDefault: true,
  jsonData: {
    httpMethod: 'GET',
  },
  name: 'prometheus',
  type: 'prometheus',
  url: 'http://localhost/prometheus/',
};
util.makeResource('Datasource', datasource.name, datasource)
