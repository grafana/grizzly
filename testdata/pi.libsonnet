{
  annotations: {
    list: [
      {
        builtIn: 1,
        datasource: '-- Grafana --',
        enable: true,
        hide: true,
        iconColor: 'rgba(0, 211, 255, 1)',
        name: 'Annotations & Alerts',
        type: 'dashboard',
      },
    ],
  },
  editable: true,
  gnetId: null,
  graphTooltip: 0,
  iteration: 1613251806619,
  links: [],
  panels: [
    {
      datasource: null,
      fieldConfig: {
        defaults: {
          color: {
            mode: 'thresholds',
          },
          custom: {},
          decimals: 1,
          mappings: [],
          thresholds: {
            mode: 'percentage',
            steps: [
              {
                color: 'green',
                value: null,
              },
              {
                color: '#EAB839',
                value: 80,
              },
              {
                color: 'red',
                value: 90,
              },
            ],
          },
          unit: 'percentunit',
        },
        overrides: [],
      },
      gridPos: {
        h: 6,
        w: 6,
        x: 0,
        y: 0,
      },
      id: 15,
      options: {
        reduceOptions: {
          calcs: [
            'mean',
          ],
          fields: '',
          values: false,
        },
        showThresholdLabels: false,
        showThresholdMarkers: true,
        text: {},
      },
      pluginVersion: '7.4.0',
      targets: [
        {
          expr: 'rate(bt_homehub_download_bytes_total[5m])/bt_homehub_download_rate_mbps/1024',
          interval: '',
          legendFormat: 'Download',
          refId: 'A',
        },
      ],
      timeFrom: null,
      timeShift: null,
      title: 'Download Saturation',
      type: 'gauge',
    },
    {
      aliasColors: {},
      bars: false,
      dashLength: 10,
      dashes: false,
      datasource: null,
      fieldConfig: {
        defaults: {
          custom: {},
          thresholds: {
            mode: 'absolute',
            steps: [],
          },
          unit: 'percentunit',
        },
        overrides: [],
      },
      fill: 4,
      fillGradient: 0,
      gridPos: {
        h: 6,
        w: 12,
        x: 6,
        y: 0,
      },
      hiddenSeries: false,
      id: 14,
      legend: {
        alignAsTable: false,
        avg: false,
        current: false,
        max: false,
        min: false,
        rightSide: false,
        show: true,
        total: false,
        values: false,
      },
      lines: true,
      linewidth: 2,
      nullPointMode: 'null',
      options: {
        alertThreshold: true,
      },
      percentage: false,
      pluginVersion: '7.4.0',
      pointradius: 2,
      points: false,
      renderer: 'flot',
      seriesOverrides: [],
      spaceLength: 10,
      stack: false,
      steppedLine: false,
      targets: [
        {
          expr: 'rate(bt_homehub_download_bytes_total[5m])/bt_homehub_download_rate_mbps/1024',
          interval: '',
          legendFormat: 'Download',
          refId: 'A',
        },
        {
          expr: '-rate(bt_homehub_upload_bytes_total[5m])/bt_homehub_upload_rate_mbps/1024',
          interval: '',
          legendFormat: 'Upload',
          refId: 'B',
        },
      ],
      thresholds: [],
      timeFrom: null,
      timeRegions: [
        {
          '$$hashKey': 'object:66',
          colorMode: 'background6',
          fill: true,
          fillColor: 'rgba(234, 112, 112, 0.12)',
          line: false,
          lineColor: 'rgba(237, 46, 24, 0.60)',
          op: 'time',
        },
      ],
      timeShift: null,
      title: 'Link Saturation Percent',
      tooltip: {
        shared: true,
        sort: 0,
        value_type: 'individual',
      },
      type: 'graph',
      xaxis: {
        buckets: null,
        mode: 'time',
        name: null,
        show: true,
        values: [],
      },
      yaxes: [
        {
          format: 'percentunit',
          label: null,
          logBase: 1,
          max: null,
          min: null,
          show: true,
        },
        {
          format: 'short',
          label: null,
          logBase: 1,
          max: null,
          min: null,
          show: true,
        },
      ],
      yaxis: {
        align: false,
        alignLevel: null,
      },
    },
    {
      datasource: null,
      fieldConfig: {
        defaults: {
          color: {
            mode: 'thresholds',
          },
          custom: {},
          decimals: 1,
          mappings: [],
          thresholds: {
            mode: 'percentage',
            steps: [
              {
                color: 'green',
                value: null,
              },
              {
                color: '#EAB839',
                value: 80,
              },
              {
                color: 'red',
                value: 90,
              },
            ],
          },
          unit: 'percentunit',
        },
        overrides: [],
      },
      gridPos: {
        h: 6,
        w: 6,
        x: 18,
        y: 0,
      },
      id: 16,
      options: {
        reduceOptions: {
          calcs: [
            'mean',
          ],
          fields: '',
          values: false,
        },
        showThresholdLabels: false,
        showThresholdMarkers: true,
        text: {},
      },
      pluginVersion: '7.4.0',
      targets: [
        {
          expr: 'rate(bt_homehub_upload_bytes_total[5m])/bt_homehub_upload_rate_mbps/1024',
          interval: '',
          legendFormat: 'Upload',
          refId: 'B',
        },
      ],
      timeFrom: null,
      timeShift: null,
      title: 'Upload Saturation',
      type: 'gauge',
    },
    {
      datasource: null,
      fieldConfig: {
        defaults: {
          color: {
            mode: 'thresholds',
          },
          custom: {},
          decimals: 0,
          mappings: [],
          thresholds: {
            mode: 'absolute',
            steps: [
              {
                color: 'green',
                value: null,
              },
            ],
          },
          unit: 'short',
        },
        overrides: [],
      },
      gridPos: {
        h: 3,
        w: 6,
        x: 0,
        y: 6,
      },
      id: 21,
      options: {
        colorMode: 'value',
        graphMode: 'area',
        justifyMode: 'auto',
        orientation: 'auto',
        reduceOptions: {
          calcs: [
            'mean',
          ],
          fields: '',
          values: false,
        },
        text: {},
        textMode: 'auto',
      },
      pluginVersion: '7.4.0',
      targets: [
        {
          expr: 'sum(increase(wordpress_post_view_count[24h]))',
          interval: '',
          legendFormat: '',
          refId: 'A',
        },
      ],
      timeFrom: null,
      timeShift: null,
      title: 'Hits in last Day',
      type: 'stat',
    },
    {
      datasource: null,
      description: '',
      fieldConfig: {
        defaults: {
          color: {
            mode: 'continuous-GrYlRd',
          },
          custom: {},
          decimals: 0,
          mappings: [],
          thresholds: {
            mode: 'absolute',
            steps: [
              {
                color: 'green',
                value: null,
              },
              {
                color: 'red',
                value: 80,
              },
            ],
          },
          unit: 'none',
        },
        overrides: [],
      },
      gridPos: {
        h: 6,
        w: 18,
        x: 6,
        y: 6,
      },
      id: 18,
      options: {
        colorMode: 'background',
        graphMode: 'area',
        justifyMode: 'center',
        orientation: 'auto',
        reduceOptions: {
          calcs: [
            'mean',
          ],
          fields: '',
          values: false,
        },
        text: {
          titleSize: 15,
        },
        textMode: 'auto',
      },
      pluginVersion: '7.4.0',
      targets: [
        {
          expr: 'sum by(namespace) (increase(wordpress_post_view_count[1h]))',
          interval: '',
          legendFormat: '{{namespace}}',
          refId: 'A',
        },
      ],
      timeFrom: null,
      timeShift: null,
      title: 'Last Hour Hits per Site',
      type: 'stat',
    },
    {
      datasource: null,
      fieldConfig: {
        defaults: {
          color: {
            mode: 'thresholds',
          },
          custom: {},
          decimals: 0,
          mappings: [],
          thresholds: {
            mode: 'absolute',
            steps: [
              {
                color: 'green',
                value: null,
              },
            ],
          },
          unit: 'short',
        },
        overrides: [],
      },
      gridPos: {
        h: 3,
        w: 6,
        x: 0,
        y: 9,
      },
      id: 20,
      options: {
        colorMode: 'value',
        graphMode: 'area',
        justifyMode: 'auto',
        orientation: 'auto',
        reduceOptions: {
          calcs: [
            'mean',
          ],
          fields: '',
          values: false,
        },
        text: {},
        textMode: 'auto',
      },
      pluginVersion: '7.4.0',
      targets: [
        {
          expr: 'sum(increase(wordpress_post_view_count[1h]))',
          interval: '',
          legendFormat: '',
          refId: 'A',
        },
      ],
      timeFrom: null,
      timeShift: null,
      title: 'Hits in last Hour',
      type: 'stat',
    },
  ],
  refresh: false,
  schemaVersion: 27,
  style: 'dark',
  tags: [],
  templating: {
    list: [
      {
        allValue: null,
        current: {
          selected: false,
          text: 'SG4B1000E020',
          value: 'SG4B1000E020',
        },
        datasource: 'Prometheus',
        definition: 'label_values(bt_homehub_build_info, firmware)',
        description: null,
        'error': null,
        hide: 2,
        includeAll: false,
        label: 'firmware',
        multi: false,
        name: 'firmware',
        options: [
          {
            selected: true,
            text: 'SG4B1000E020',
            value: 'SG4B1000E020',
          },
        ],
        query: 'label_values(bt_homehub_build_info, firmware)',
        refresh: 0,
        regex: '',
        skipUrlSync: false,
        sort: 0,
        tagValuesQuery: '',
        tags: [],
        tagsQuery: '',
        type: 'query',
        useTags: false,
      },
    ],
  },
  time: {
    from: 'now-12h',
    to: 'now',
  },
  timepicker: {},
  timezone: '',
  title: 'Pi Summary',
  uid: 'pi',
  version: 6,
}
