local sm = import 'synthetic-monitoring/sm.libsonnet';
{
  syntheticMonitoring+:: {
    grafanaHttpCheck: sm.new('grafana', 'https://grafana.com/')
                      + sm.withHttp()
                      + sm.withProbes('all'),  // enable all probes
    grafanaPingCheck: sm.new('grafana', 'grafana.com')
                      + sm.withPing()
                      + sm.withProbes('continents'),  // one check per continent
    grafanaDnsCheck: sm.new('grafana', 'grafana.com')
                     + sm.withDns()
                     + sm.withProbes('europe'),  // just check from Europe
    grafanaTcpCheck: sm.new('grafana', 'grafana.com:443')
                     + sm.withTcp()
                     + sm.withProbes('small'),  // just use a smaller, predefined set of checks
  },
}
