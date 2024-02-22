local dashboard(name) = {
  "apiVersion": "grizzly.grafana.com/v1alpha1",
  "kind": "Dashboard",
  "metadata": {
    "name": name,
  },
  "spec": {
    "annotations": {
      "list": [
        {
          "builtIn": 1,
          "datasource": {
            "type": "grafana",
            "uid": "-- Grafana --"
          },
          "enable": true,
          "hide": true,
          "iconColor": "rgba(0, 211, 255, 1)",
          "name": "Annotations \u0026 Alerts",
          "type": "dashboard"
        }
      ]
    },
    "editable": true,
    "fiscalYearStartMonth": 0,
    "graphTooltip": 0,
    "id": 4278,
    "links": [],
    "liveNow": false,
    "panels": [
      {
        "datasource": {
          "type": "datasource",
          "uid": "grafana"
        },
        "fieldConfig": {
          "defaults": {
            "color": {
              "mode": "palette-classic"
            },
            "custom": {
              "axisCenteredZero": false,
              "axisColorMode": "text",
              "axisLabel": "",
              "axisPlacement": "auto",
              "axisShow": false,
              "fillOpacity": 80,
              "gradientMode": "none",
              "hideFrom": {
                "legend": false,
                "tooltip": false,
                "viz": false
              },
              "lineWidth": 1,
              "scaleDistribution": {
                "type": "linear"
              },
              "thresholdsStyle": {
                "mode": "off"
              }
            },
            "mappings": [],
            "thresholds": {
              "mode": "absolute",
              "steps": [
                {
                  "color": "green",
                  "value": null
                },
                {
                  "color": "red",
                  "value": 80
                }
              ]
            }
          },
          "overrides": []
        },
        "gridPos": {
          "h": 8,
          "w": 12,
          "x": 0,
          "y": 0
        },
        "id": 1,
        "options": {
          "barRadius": 0,
          "barWidth": 0.97,
          "fullHighlight": false,
          "groupWidth": 0.7,
          "legend": {
            "calcs": [],
            "displayMode": "list",
            "placement": "bottom",
            "showLegend": true
          },
          "orientation": "auto",
          "showValue": "auto",
          "stacking": "none",
          "tooltip": {
            "mode": "single",
            "sort": "none"
          },
          "xTickLabelRotation": 0,
          "xTickLabelSpacing": 0
        },
        "targets": [
          {
            "channel": "plugin/testdata/random-2s-stream",
            "datasource": {
              "type": "datasource",
              "uid": "grafana"
            },
            "filter": {
              "fields": [
                "Time",
                "Min"
              ]
            },
            "queryType": "randomWalk",
            "refId": "A"
          }
        ],
        "title": "Panel Title",
        "type": "barchart"
      }
    ],
    "refresh": "",
    "schemaVersion": 38,
    "tags": [],
    "templating": {
      "list": []
    },
    "time": {
      "from": "now-6h",
      "to": "now"
    },
    "timepicker": {},
    "timezone": "",
    "title": name,
    "uid": name,
    "version": 1,
    "weekStart": ""
  }
};

{
  application_1: {
    dashboard: dashboard("application-1"),
  },
  application_2: {
    dashboard: dashboard("application-2"),
  }
}