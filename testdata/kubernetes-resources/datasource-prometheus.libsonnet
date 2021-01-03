local util = import 'util.libsonnet';

util.makeResource(
  'grafana.com/v1', 'datasource', 'prometheus.yml', {
    'prometheus.yml': {
      access: 'proxy',
      isDefault: true,
      jsonData: {
        httpMethod: 'GET',
      },
      name: 'prometheus',
      type: 'prometheus',
      url: 'http://localhost/prometheus/',
    },
  },
)
