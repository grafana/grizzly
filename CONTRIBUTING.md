# Contributing Guidelines

This document is a guide to help you through the process of contributing to `grizzly`.

## Developing the extension

### Create a fork

[Fork][fork], then clone the repository:

```shell
git clone git@github.com:{your_github_username}/grizzly.git
cd grizzly
git remote add upstream https://github.com/grafana/grizzly.git
git fetch upstream
```

### Install dependencies

```shell
go mod vendor
```

### Building from sources

To build Grizzly locally, inside a checkout, run:

```shell
make dev
```

### Run locally

To run Grizzly locally, use:

```shell
go run cmd/grr/*.go
```

## Run the tests

There are both unit tests and integration tests in this repo.

`config_test.go` and `unit_test.go` can both be executed autonomously.

However, in order to run integration tests you need to make use of 
`make test`. This command spins up a Grafana instance with certain config 
and provisioned resources, so the integration tests can run against it.

You can also manually execute `make run-test-image-locally` and run the tests
for debugging.

## Releasing grizzly

Releasing is done as follows:

1. Trigger the release pipeline by creating and pushing a tag: `git tag {version} && git push origin {version}`
2. The pipeline publishes docker images of grizzly and creates a GitHub release. The latter, however, is marked as draft and as such needs to be reviewed and updated with change notes before being published.


[fork]: https://github.com/grafana/grizzly/fork
