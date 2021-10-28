package grafana

import (
	"testing"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/synthetic-monitoring-agent/pkg/pb/synthetic_monitoring"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyntheticMonitoring(t *testing.T) {

	grizzly.ConfigureProviderRegistry(
		[]grizzly.Provider{
			&Provider{},
		})

	t.Run("Check getUID is functioning correctly", func(t *testing.T) {
		resource := grizzly.Resource{
			"metadata": map[string]interface{}{
				"name": "test",
				"type": "http",
			},
		}
		handler := SyntheticMonitoringHandler{}
		uid, err := handler.GetUID(resource)
		require.NoError(t, err)
		require.Equal(t, uid, "http.test")
	})
}

func TestSyntheticMonitoringCheckUID(t *testing.T) {
	testCases := []struct {
		name        string
		check       synthetic_monitoring.Check
		expectedUID string
	}{
		{
			name: "http check",
			check: synthetic_monitoring.Check{
				Job: "https://website.com",
				Settings: synthetic_monitoring.CheckSettings{
					Http: &synthetic_monitoring.HttpSettings{},
				},
			},
			expectedUID: "http.https://website.com",
		},
		{
			name: "ping check",
			check: synthetic_monitoring.Check{
				Job: "10.1.2.3",
				Settings: synthetic_monitoring.CheckSettings{
					Ping: &synthetic_monitoring.PingSettings{},
				},
			},
			expectedUID: "ping.10.1.2.3",
		},
		{
			name: "dns check",
			check: synthetic_monitoring.Check{
				Job: "website.com",
				Settings: synthetic_monitoring.CheckSettings{
					Dns: &synthetic_monitoring.DnsSettings{},
				},
			},
			expectedUID: "dns.website.com",
		},
		{
			name: "tcp check",
			check: synthetic_monitoring.Check{
				Job: "website.com",
				Settings: synthetic_monitoring.CheckSettings{
					Tcp: &synthetic_monitoring.TcpSettings{},
				},
			},
			expectedUID: "tcp.website.com",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedUID, getUID(tc.check))
		})
	}
}
