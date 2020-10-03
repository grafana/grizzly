{
  grafanaDatasources+:: {
    'prometheus.yml': {
      access: 'proxy',
      editable: false,
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
