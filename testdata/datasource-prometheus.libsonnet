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
      basicAuth: false,
      basicAuthPassword: '',
      basicAuthUser: '',
      database: '',
      orgId: 1,
      password: '',
      readOnly: false,
      secureJsonFields: {},
      typeLogoUrl: '',
      user: '',
      withCredentials: false,
    },
  },
}
