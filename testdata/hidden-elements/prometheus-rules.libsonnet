local prom = import 'prom.libsonnet';
local promRuleGroupSet = prom.v1.ruleGroupSet;
local promRuleGroup = prom.v1.ruleGroup;
{
  prometheus_metamon::
    promRuleGroup.new('grizzly_alerts')
    + promRuleGroup.rule.newAlert(
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
    + promRuleGroup.rule.newRecording(
      'job:up:sum', {
        expr: 'sum by(job) (up)',
      },
    ),

  prometheusAlerts+:: {
    first_rules: promRuleGroupSet.new() + promRuleGroupSet.addGroup($.prometheus_metamon),
  },
}
