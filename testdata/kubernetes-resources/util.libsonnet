{
  makeResource(apiVersion, kind, name, spec):: {
    apiVersion: apiVersion,
    kind: kind,
    metadata: {
      name: name,
    },
    spec: spec,
  },
}
