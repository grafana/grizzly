{
  makeResource(kind, name, spec):: {
    apiVersion: 'grafana.com/grizzly/v1',
    kind: kind,
    metadata: {
      name: name,
    },
    spec: spec,
  },
}
