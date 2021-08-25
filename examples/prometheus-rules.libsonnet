local ruleGroup = (import 'prom.libsonnet').v1.ruleGroup;

ruleGroup.new('grizzly_alerts')
+ ruleGroup.rule.newAlert(
  'PromScrapeFailed', {
    expr: 'up != 1',
    'for': '1m',
    labels: {
      severity: 'critical',
    },
    annotations: {
      message: 'Prometheus failed to scrape a target {{ $labels.job }}  / {{ $labels.instance }}',
    },
  }
)
+ ruleGroup.rule.newRecording(
  'job:up:sum', {
    expr: 'sum by(job) (up)',
  },
)
