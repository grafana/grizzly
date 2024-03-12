# Grizzly - `setup` GitHub Action

This action downloads and adds Grizzly to the PATH.

## Inputs

### `version`

**Required** Version of Grizzly to install. Default `"latest"`.

## Example usage

```yaml
uses: grafana/grizzly/actions/setup@main
with:
  version: 'v0.4.0'
```
