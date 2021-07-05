{
  grafanaDatasources+:: {
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
}
