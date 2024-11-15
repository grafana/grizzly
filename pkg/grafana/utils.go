package grafana

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"

	gclient "github.com/grafana/grafana-openapi-client-go/client"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/grafana/grizzly/internal/httputils"
	"github.com/grafana/grizzly/pkg/grizzly"
)

var (
	stringtrue     = "true"
	folderURLRegex = regexp.MustCompile("/dashboards/f/([^/]+)")
)

func extractFolderUID(client *gclient.GrafanaHTTPAPI, d models.DashboardFullWithMeta) string {
	folderUID := d.Meta.FolderUID
	if folderUID == "" {
		urlPaths := folderURLRegex.FindStringSubmatch(d.Meta.FolderURL)
		if len(urlPaths) == 0 {
			if d.Meta.FolderID == generalFolderID { // nolint:staticcheck
				return generalFolderUID
			}
			folder, err := getFolderByID(client, d.Meta.FolderID) // nolint:staticcheck
			if err != nil {
				return ""
			}
			return folder.UID
		}
		folderUID = urlPaths[1]
	}
	return folderUID
}

func structToMap(s interface{}) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func authenticateAndProxyHandler(s grizzly.Server, provider grizzly.Provider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")

		config := provider.(ClientProvider).Config()
		if config.URL == "" {
			httputils.Error(w, "Error: No Grafana URL configured", fmt.Errorf("no Grafana URL configured"), http.StatusBadRequest)
			return
		}

		req, err := http.NewRequest(http.MethodGet, config.URL+r.URL.Path, nil)
		if err != nil {
			httputils.Error(w, http.StatusText(http.StatusInternalServerError), err, http.StatusInternalServerError)
			return
		}

		if config.User != "" {
			req.SetBasicAuth(config.User, config.Token)
		} else if config.Token != "" {
			req.Header.Set("Authorization", "Bearer "+config.Token)
		}

		req.Header.Set("User-Agent", s.UserAgent)

		client, err := httputils.NewHTTPClient()
		if err != nil {
			httputils.Error(w, http.StatusText(http.StatusInternalServerError), err, http.StatusInternalServerError)
			return
		}

		resp, err := client.Do(req)

		if err == nil {
			body, _ := io.ReadAll(resp.Body)
			httputils.Write(w, body)
			return
		}

		msg := ""
		if config.Token == "" {
			msg += "<p><b>Warning:</b> No service account token specified.</p>"
		}

		if resp.StatusCode == http.StatusFound {
			w.WriteHeader(http.StatusUnauthorized)
			httputils.Write(w, []byte(fmt.Sprintf("%s<p>Authentication error</p>", msg)))
		} else {
			body, _ := io.ReadAll(resp.Body)
			w.WriteHeader(resp.StatusCode)
			httputils.Write(w, []byte(fmt.Sprintf("%s%s", msg, string(body))))
		}
	}
}
