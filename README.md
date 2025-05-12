# Grizzly: Manage your Observability Systems

> [!IMPORTANT]
> Grizzly has been marked for deprecation and will be superseded by [grafanactl](https://github.com/grafana/grafanactl). See this [discussion](https://github.com/grafana/grafanactl/discussions/97) for more information.

Grafana Grizzly is a command line tool that allows you to manage your observability resources as code.

Now you can define your dashboards, alerting and recording rules,… all from within your codebase.

You can build Grizzly into your continuous deployment pipelines, meaning whenever you deploy your application, your observability is updated too.

Currently supported are:

* [Grafana](https://grafana.github.io/grizzly/grafana/)
  * [Alerting contact points](https://grafana.github.io/grizzly/grafana/#contact-points)
  * [Alerting notification templates](https://grafana.github.io/grizzly/grafana/#notification-templates)
  * [Alerting notification policy](https://grafana.github.io/grizzly/grafana/#notification-policy)
  * [Alert rule groups](https://grafana.github.io/grizzly/grafana/#alertrulegroup)
  * [Dashboards](https://grafana.github.io/grizzly/grafana/#dashboards)
  * [Dashboard folders](https://grafana.github.io/grizzly/grafana/#folders)
  * [Datasources](https://grafana.github.io/grizzly/grafana/#datasources)
  * [Library elements](https://grafana.github.io/grizzly/grafana/#library-elements)
* [Prometheus](https://grafana.github.io/grizzly/prometheus/)
  * [Alerts](https://grafana.github.io/grizzly/prometheus/#prometheus-alerts)
  * [Recording rules](https://grafana.github.io/grizzly/prometheus/#prometheus-recording-rules)
* [Grafana Synthetic Monitoring](https://grafana.github.io/grizzly/synthetic-monitoring/)
  * [Monitoring Checks](https://grafana.github.io/grizzly/synthetic-monitoring/#grafana-cloud-synthetic-monitoring-checks)

See [docs](https://grafana.github.io/grizzly) for more details.

## Contributing

See our [contributing guide](CONTRIBUTING.md).
