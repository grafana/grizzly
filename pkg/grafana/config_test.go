package grafana

import (
	"os"
	"testing"
)

func TestParseEnvironment(t *testing.T) {
	tests := map[string]struct {
		url    string
		path   string
		user   string
		token  string
		expect string
		err    bool
	}{
		"GRAFANA_URL only": {
			"https://my.grafana.net",
			"/this",
			"",
			"",
			"https://my.grafana.net/this",
			false,
		},
		"w/ token": {
			"https://my.grafana.net",
			"/that",
			"",
			"token",
			"https://api_key:token@my.grafana.net/that",
			false,
		},
		"Basic auth": {
			"https://my.grafana.net",
			"/secure",
			"user",
			"pass",
			"https://user:pass@my.grafana.net/secure",
			false,
		},
		"GRAFANA_URL blank": {
			"",
			"",
			"",
			"",
			"",
			true,
		},
	}
	for testName, test := range tests {
		if test.url != "" {
			os.Setenv("GRAFANA_URL", test.url)
		} else {
			os.Unsetenv("GRAFANA_URL")
		}
		if test.user != "" {
			os.Setenv("GRAFANA_USER", test.user)
		} else {
			os.Unsetenv("GRAFANA_USER")
		}
		if test.token != "" {
			os.Setenv("GRAFANA_TOKEN", test.token)
		} else {
			os.Unsetenv("GRAFANA_TOKEN")
		}
		t.Logf("Running test case, %q...", testName)
		url, err := getGrafanaURL(test.path)
		if err != nil && !test.err {
			t.Errorf("Unexpected error getting Jsonnet files: %s", err)
		}
		if url != test.expect {
			t.Errorf("Expected GrafanaURL %s, got: %s", test.expect, url)
		}
	}
}
