package grafana

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"
)

func getGrafanaURL(urlPath string) (string, error) {
	if grafanaURL, exists := os.LookupEnv("GRAFANA_URL"); exists {
		u, err := url.Parse(grafanaURL)
		if err != nil {
			return "", err
		}
		parts := strings.Split(urlPath, "?")
		u.Path = path.Join(u.Path, parts[0])
		if len(parts) > 1 {
			u.RawQuery = parts[1]
		}
		if token, exists := os.LookupEnv("GRAFANA_TOKEN"); exists {
			user, exists := os.LookupEnv("GRAFANA_USER")
			if !exists {
				user = "api_key"
			}
			u.User = url.UserPassword(user, token)
		}
		return u.String(), nil
	}
	return "", fmt.Errorf("Require GRAFANA_URL (optionally GRAFANA_TOKEN & GRAFANA_USER")
}

func getWSGrafanaURL(urlPath string) (string, string, error) {
	grafanaURL, exists := os.LookupEnv("GRAFANA_URL")
	if !exists {
		return "", "", fmt.Errorf("Require GRAFANA_URL (optionally GRAFANA_TOKEN if auth required) for websocket actions")
	}
	u, err := url.Parse(grafanaURL)
	if err != nil {
		return "", "", err
	}
	if u.Scheme == "https" {
		u.Scheme = "wss"
	} else {
		u.Scheme = "ws"
	}
	u.Path = path.Join(u.Path, urlPath)
	grafanaURL = u.String()
	token, ok := os.LookupEnv("GRAFANA_TOKEN")
	if ok {
		u.User = nil
		return u.String(), token, nil
	}
	if u.User != nil {
		token, ok := u.User.Password()
		if !ok {
			return u.String(), "", nil
		}
		u.User = nil
		return u.String(), token, nil
	}
	return u.String(), "", nil
}
