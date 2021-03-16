{
  makeResource(kind, name, spec, metadata={}):: {
    apiVersion: 'grizzly.grafana.com/v1alpha1',
    kind: kind,
    metadata: {
      name: name,
    } + metadata,
    spec: spec,
  },
}
