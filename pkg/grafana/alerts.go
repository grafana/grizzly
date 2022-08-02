package grafana

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/grafana/grizzly/pkg/grizzly"
	log "github.com/sirupsen/logrus"
)

func alertDatasources() []string {
	return []string{"loki", "prometheus"}
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

// getRemoteAlertGrouping retrieves a datasource object from Grafana
func getRemoteAlertGroup(datasourceAndGroup string) (*grizzly.Resource, error) {
	client := new(http.Client)

	s := strings.Split(datasourceAndGroup, "||")

	datasource := s[0]
	datasources, err := getRemoteDatasources()

	if err != nil {
		return nil, err
	}

	datasourceDef, err := getDatasource(datasources, datasource)
	datasourceID := fmt.Sprintf("%s", datasourceDef["uid"])

	if datasourceID == "" {
		return nil, fmt.Errorf("Invalid Datasource %s", datasource)
	}
	if err != nil {
		return nil, err
	}

	grafanaURL, err := getGrafanaURL(fmt.Sprintf("api/ruler/%s/api/v1/rules/%s", datasourceID, s[1]))

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", grafanaURL, nil)
	if err != nil {
		return nil, err
	}

	if grafanaToken, ok := getGrafanaToken(); ok {
		req.Header.Set("Authorization", "Bearer "+grafanaToken)
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusNotFound:
		return nil, grizzly.ErrNotFound
	case resp.StatusCode >= 400:
		return nil, errors.New(resp.Status)
	}

	data, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	var d AlertWrapper
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, grizzly.APIErr{Err: err, Body: data}
	}
	h := AlertsHandler{}

	var mapStringInterface map[string]interface{}
	inrec, _ := json.Marshal(d)
	json.Unmarshal(inrec, &mapStringInterface)

	resource := grizzly.NewResource(h.APIVersion(), h.Kind(), s[1], mapStringInterface)

	nG := strings.Split(s[1], "/")

	resource.SetMetadata("namespace", nG[0])
	resource.SetMetadata("group", nG[1])
	resource.SetMetadata("datasource", datasourceDef["name"].(string))
	resource.SetMetadata("datasource_uid", datasourceDef["uid"].(string))
	resource.SetMetadata("datasource_type", datasourceDef["type"].(string))

	return &resource, nil
}

// Alert encapsulates a alert
type AlertGroup map[string]interface{}
type AlertNamespaces map[string][]AlertGroup

// UID retrieves the UID from a alert
func (d *AlertGroup) Name() string {
	name, ok := (*d)["name"]
	if !ok {
		return ""
	}
	return name.(string)
}

// toJSON returns JSON for a alert
func (d *AlertGroup) toJSON() (string, error) {
	j, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return "", err
	}
	return string(j), nil
}

// toJSON returns JSON for a alert
func (d *GrafanaAlertWrapper) toJSON() (string, error) {
	j, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return "", err
	}
	return string(j), nil
}

type Rule struct {
	Alert        string                 `json:"alert"`
	Expr         string                 `json:"expr"`
	For          string                 `json:"for"`
	Labels       map[string]interface{} `json:"labels"`
	Annotations  map[string]interface{} `json:"annotations"`
	GrafanaAlert map[string]interface{} `json:"grafana_alert"`
}

// AlertWrapper adds wrapper to a alert JSON. Caters both for Grafana's POST
// API as well as GET which require different JSON.
type AlertWrapper struct {
	Name string `json:"name"`
	// Rules    []Rule `json:"rules"`
	Rules []interface{} `json:"rules"`
}

// AlertWrapper adds wrapper to a alert JSON. Caters both for Grafana's POST
// API as well as GET which require different JSON.
type GrafanaAlertWrapper struct {
	Name string `json:"name"`
	// Rules    []Rule `json:"rules"`
	Rules    []interface{} `json:"rules"`
	Interval int           `json:"interval"`
}

// UID retrieves the UID from a alert wrapper
func (d *AlertWrapper) UID(s string) string {
	return d.Name
}

// toJSON returns JSON expected by Grafana API
func (d *AlertWrapper) toJSON() (string, error) {
	// d.Overwrite = true
	j, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return "", err
	}
	return string(j), nil
}

// PrometheusAlertGroup encapsulates a list of alerts
type PrometheusAlertGroup struct {
	Namespace string                   `yaml:"-"`
	Name      string                   `yaml:"name"`
	Alerts    []map[string]interface{} `yaml:"alerts"`
}

// PrometheusAlertGrouping encapsulates a set of named alert groups
type PrometheusAlertGrouping struct {
	Namespace string                 `json:"namespace"`
	Groups    []PrometheusAlertGroup `json:"groups"`
}

func postAlertGroup(resource grizzly.Resource) error {
	client := new(http.Client)
	datasource := resource.GetMetadata("datasource")
	namespace := resource.GetMetadata("namespace")

	datasources, err := getRemoteDatasources()

	datasourceID, _ := datasourceIDFromName(datasources, datasource)

	subType := ""
	if datasourceID != "grafana" {
		subType = "?subtype=cortex"
	}

	grafanaURL, err := getGrafanaURL(fmt.Sprintf("api/ruler/%s/api/v1/rules/%s%s", datasourceID, namespace, subType))

	if err != nil {
		return err
	}

	spec := resource["spec"].(map[string]interface{})

	wrappedJSON := ""
	if datasourceID == "grafana" {
		interval, ok := spec["interval"].(int)
		if !ok {
			interval = 0
		}

		wrappedBoard := GrafanaAlertWrapper{
			Name:     spec["name"].(string),
			Rules:    spec["rules"].([]interface{}),
			Interval: interval,
		}

		wrappedJSON, _ = wrappedBoard.toJSON()

	} else {
		wrappedBoard := AlertWrapper{
			Name:  spec["name"].(string),
			Rules: spec["rules"].([]interface{}),
		}

		wrappedJSON, _ = wrappedBoard.toJSON()

	}

	req, err := http.NewRequest("POST", grafanaURL, bytes.NewBufferString(wrappedJSON))

	if err != nil {
		return err
	}

	if grafanaToken, ok := getGrafanaToken(); ok {
		req.Header.Set("Authorization", "Bearer "+grafanaToken)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusAccepted:
		return nil
	case http.StatusPreconditionFailed:
		d := json.NewDecoder(resp.Body)
		var r struct {
			Message string `json:"message"`
		}
		if err := d.Decode(&r); err != nil {
			return fmt.Errorf("Failed to decode actual error (412 Precondition failed): %s", err)
		}

		return fmt.Errorf("Error while applying '%s' to Grafana: %s", resource.Name(), r.Message)
	default:
		return NewErrNon200Response("alert", resource.Name(), resp)
	}
}

/*
 * Not all datasources can be alerted on.  This may need work as it makes some assumptions on
 * what is going on in the backend I'm not sure about.  It assumes that Lokia nd Promethues are t
 * the only datasources you can alert on(which at least seems not true with the introduction of
 * mimir, though I'm not sure if the datasource will still say Loki/prometheus even though the
 * storage backend is different.  It also assumes "grafana" is an alert type outside of the
 * specific datasources(the Web UI pulls these as three different alert types even though
 * 'Grafana' alerts can query loki or prometheus)  And it assumes that there are certain
 * data sources that can't be used as alert backends, even if they are loki/prometheus.
 * From looking at the json blob I get back from /api/datasources, it seems like anything
 * with the jsonData.manageAlerts = false is one of those things which isn't an alert datasource
 * (If there is a better way to figure this out, that'd be great!)
 */
func alertableDatasourceIDs(datasources []map[string]interface{}) ([]string, error) {
	datasource_uids := []string{"grafana"}

	for _, datasource := range datasources {
		if contains(alertDatasources(), datasource["type"].(string)) {
			if _, ok := datasource["jsonData"]; ok {
				if _, ok := datasource["jsonData"].(map[string]interface{})["manageAlerts"]; ok {
					if datasource["jsonData"].(map[string]interface{})["manageAlerts"].(bool) == false {
						continue
					}
				}
			}

			datasource_uids = append(datasource_uids, datasource["uid"].(string))
		}
	}

	return datasource_uids, nil
}

func getDatasourceID(datasources []map[string]interface{}, datasourceSlug string) (string, error) {
	if datasourceSlug == "grafana" {
		return "grafana", nil
	}

	for _, datasource := range datasources {
		if datasourceSlug == datasource["type"].(string) {
			if _, ok := datasource["jsonData"]; ok {
				if _, ok := datasource["jsonData"].(map[string]interface{})["manageAlerts"]; ok {
					if datasource["jsonData"].(map[string]interface{})["manageAlerts"].(bool) == false {
						continue
					}
				}
			}
			return fmt.Sprintf("%s", datasource["uid"].(string)), nil
		}
	}

	return "", nil
}

func datasourceIDFromName(datasources []map[string]interface{}, datasourceSlug string) (string, error) {
	if datasourceSlug == "grafana" {
		return "grafana", nil
	}

	for _, datasource := range datasources {
		if datasourceSlug == datasource["name"].(string) {
			if _, ok := datasource["jsonData"]; ok {
				if _, ok := datasource["jsonData"].(map[string]interface{})["manageAlerts"]; ok {
					if datasource["jsonData"].(map[string]interface{})["manageAlerts"].(bool) == false {
						continue
					}
				}
			}
			return fmt.Sprintf("%s", datasource["uid"].(string)), nil
		}
	}

	return "", nil
}

func getDatasource(datasources []map[string]interface{}, datasourceSlug string) (map[string]interface{}, error) {
	if datasourceSlug == "grafana" {
		return map[string]interface{}{"uid": "grafana", "name": "grafana", "type": "grafana"}, nil
	}

	for _, datasource := range datasources {
		if datasourceSlug == datasource["name"].(string) {
			if _, ok := datasource["jsonData"]; ok {
				if _, ok := datasource["jsonData"].(map[string]interface{})["manageAlerts"]; ok {
					if datasource["jsonData"].(map[string]interface{})["manageAlerts"].(bool) == false {
						continue
					}
				}
			}
			datasource["uid"] = fmt.Sprintf("%s", datasource["uid"])
			return datasource, nil
		}
	}
	return map[string]interface{}{}, nil
}

func getDatasourceFromID(datasources []map[string]interface{}, datasourceID string) (map[string]interface{}, error) {
	if datasourceID == "grafana" {
		return map[string]interface{}{"uid": "grafana", "name": "grafana", "type": "grafana"}, nil
	}

	for _, datasource := range datasources {
		foundDatasourceID := ""

		switch v := datasource["uid"].(type) {
		case string:
			foundDatasourceID = v
		case float64:
			foundDatasourceID = fmt.Sprintf("%d", int(v))
		}
		if datasourceID == foundDatasourceID {
			if _, ok := datasource["jsonData"]; ok {
				if _, ok := datasource["jsonData"].(map[string]interface{})["manageAlerts"]; ok {
					if datasource["jsonData"].(map[string]interface{})["manageAlerts"].(bool) == false {
						continue
					}
				}
			}
			datasource["uid"] = datasourceID
			return datasource, nil
		}
	}
	return map[string]interface{}{}, nil
}

func getDatasourceType(datasources []map[string]interface{}, datasource_uid string) (string, error) {
	if datasource_uid == "grafana" {
		return "grafana", nil
	}

	for _, datasource := range datasources {
		if datasource_uid == fmt.Sprintf("%d", int(datasource["uid"].(float64))) {
			return datasource["type"].(string), nil
		}
	}

	return "", nil
}

func getRemoteAlertGroupList() ([]string, error) {
	batchSize := 500

	client := new(http.Client)
	namespaceGroups := []string{}

	datasources, _ := getRemoteDatasources()
	datasource_uids, _ := alertableDatasourceIDs(datasources)

	for _, datasource := range datasource_uids {
		grafanaURL, err := getGrafanaURL(fmt.Sprintf("/api/ruler/%s/api/v1/rules?subtype=cortex&limit=%d", datasource, batchSize))

		if err != nil {
			return nil, err
		}

		req, err := http.NewRequest("GET", grafanaURL, nil)
		if err != nil {
			return nil, err
		}

		if grafanaToken, ok := getGrafanaToken(); ok {
			req.Header.Set("Authorization", "Bearer "+grafanaToken)
		}

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		switch {
		case resp.StatusCode == http.StatusNotFound:
			log.Warnf("Backend %s not found.  Alerts for this backend will not be pulled.", datasource)
			continue
		case resp.StatusCode >= 400:
			return nil, errors.New(resp.Status)
		}
		body, err := io.ReadAll(resp.Body)

		if err != nil {
			return nil, err
		}

		var alerts AlertNamespaces

		if err := json.Unmarshal([]byte(string(body)), &alerts); err != nil {
			return nil, grizzly.APIErr{Err: err, Body: body}
		}

		for namespace, groupList := range alerts {
			for _, group := range groupList {
				datasourceDef, _ := getDatasourceFromID(datasources, datasource)
				namespaceGroup := fmt.Sprintf("%s||%s/%s", datasourceDef["name"], namespace, group.Name())
				namespaceGroups = append(namespaceGroups, namespaceGroup)
			}
		}
	}
	return namespaceGroups, nil
}
