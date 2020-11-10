# Prometheus Provider for Grizzly

This provider offers a handler for Prometheus Alerts and Recording Rules.

For pure Prometheus, `grr export` can be used to convert Jsonnet based rules
into YAML files that can be loaded into Prometheus. In this context,
`grr apply` will not work, as Prometheus itself does not have an API for
the delivery of rules.

When Grafana Cloud Metrics, Grafana Metrics Enterprise or Cortex are used,
the full suite of Grizzly actions is available, e.g. `grr diff`, `grr apply`
and `grr watch`.