# Grizzly: Manage your Observability Systems

A utility for managing various observability resources with Jsonnet. Currently supported
are:

 * Grafana dashboards/dashboard folders
 * Grafana datasources
 * Grafana Cloud Prometheus recording rules/alerts
 * Grafana Synthetic Monitoring checks

See [docs](https://grafana.github.io/grizzly) for usage details.

## Building Grizzly from Source
To build Grizzly locally, inside a checkout, simply use:
```
$ make dev
```

## Running tests

There are both unit tests and integration tests in this repo.

`config_test.go` and `unit_test.go` can both be executed autonomously.

However, in order to run integration tests you need to make use of 
`make test`. This command spins up a Grafana instance with certain config 
and provisioned resources, so the integration tests can run against it.

You can also manually execute `make run-test-image-locally` and run the tests
for debugging.
