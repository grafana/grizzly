# Grizzly and Jsonnet hidden elements

Historically, Grizzly consumed observability resources via hidden elements
within the Jsonnet code. This was driven by the design of
[Monitoring Mixins](https://github.com/monitoring-mixins/docs).

However, extracting hidden elements from Jsonnet is not how Jsonnet is
intended to be used. Instead, we recommend the coding style show in
the [directory above](../README.md).

This directory contains examples that demonstrates the hidden elements
method

To look at the examples in this directory, execute:

```
grr show testdata/hidden-elements/main.jsonnet
```
