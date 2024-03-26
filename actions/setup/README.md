# Grizzly - `setup` GitHub Action

This action downloads and adds Grizzly to the PATH.

## Example usage

```yaml
uses: grafana/grizzly/actions/setup@main
with:
  version: 'v0.4.0'
  grafana.url: 'https://my-grafana-instance:3000'
  grafana.user: admin
  grafana.token: ${{ secrets.GRIZZLY_GRAFANA_TOKEN}}
```

## Inputs

### `version`

**Required** Version of Grizzly to install. Default `"latest"`.

### `grafana.url`

**Optional** Sets `grafana.url` in the configuration.

### `grafana.token`

**Optional** Sets `grafana.token` in the configuration.

### `grafana.user`

**Optional** Sets `grafana.user` in the configuration.

### `mimir.address`

**Optional** Sets `mimir.address` in the configuration.

### `mimir.tenant-id`

**Optional** Sets `mimir.tenant-id` in the configuration.

### `mimir.api-key`

**Optional** Sets `mimir.api-key` in the configuration.

### `synthetic-monitoring.token`

**Optional** Sets `synthetic-monitoring.token` in the configuration.

### `synthetic-monitoring.stack-id`

**Optional** Sets `synthetic-monitoring.stack-id` in the configuration.

### `synthetic-monitoring.metrics-id`

**Optional** Sets `synthetic-monitoring.metrics-id` in the configuration.

### `synthetic-monitoring.logs-id`

**Optional** Sets `synthetic-monitoring.logs-id` in the configuration.

### `targets`

**Optional** Sets `targets` in the configuration.

### `output-format`

**Optional** Sets `output-format` in the configuration.

### `only-spec`

**Optional** Sets `only-spec` in the configuration.
