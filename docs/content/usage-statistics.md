---
date: "2024-10-24T00:00:00+00:00"
title: "Usage statistics"
---

By default, Grizzly sends anonymous, but uniquely identifiable usage information to Grafana Labs. These statistics are sent to stats.grafana.org.

Statistics help Grafana better understand how Grizzly is used. This helps us prioritize features and documentation.

The usage information includes the following details:

* A hash of your config file
* Timestamp of when the report was created
* The version of Grizzly.
* The operating system Grizzly is running on.
* The system architecture Grizzly is running on.
* The operation performed (`pull`/`apply`)
* The number of resources affected

This list may change over time.
For auditing, the code performing this reporting can be found at the end of the [events.go](https://github.com/grafana/grizzly/blob/main/pkg/grizzly/events.go) file.

## Opt-out of data collection

You can use the `--disable-reporting` command line flag to disable the reporting and opt-out of the data collection.
