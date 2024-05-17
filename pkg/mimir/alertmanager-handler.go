package mimir

import (
	"fmt"

	"github.com/grafana/grizzly/pkg/mimir/client"
	"github.com/grafana/grizzly/pkg/mimir/models"
	"gopkg.in/yaml.v3"

	"github.com/grafana/grizzly/pkg/grizzly"
)

// AlertmanagerHandler is a Grizzly Handler for Alertmanager Configuration
type AlertmanagerHandler struct {
	grizzly.BaseHandler
	clientTool client.Mimir
}

// NewAlertmanagerHandler returns a new Grizzly Handler for Alertmanager Configuration
func NewAlertmanagerHandler(provider *Provider, clientTool client.Mimir) *AlertmanagerHandler {
	return &AlertmanagerHandler{
		BaseHandler: grizzly.NewBaseHandler(provider, "PrometheusAlertmanagerConfig", false),
		clientTool:  clientTool,
	}
}

const (
	prometheusAlertmanagerConfigPattern = "prometheus/alertmanager-%s.%s"
)

// ResourceFilePath returns the location on disk where a resource should be updated
func (h *AlertmanagerHandler) ResourceFilePath(resource grizzly.Resource, filetype string) string {
	return fmt.Sprintf(prometheusAlertmanagerConfigPattern, resource.Name(), filetype)
}

// Validate returns the uid of resource
func (h *AlertmanagerHandler) Validate(resource grizzly.Resource) error {
	uid, exist := resource.GetSpecString("uid")
	if exist && uid != resource.Name() {
		return fmt.Errorf("uid '%s' and name '%s', don't match", uid, resource.Name())
	}
	return nil
}

// GetUID returns the UID for a resource
func (h *AlertmanagerHandler) GetUID(resource grizzly.Resource) (string, error) {
	return resource.Name(), nil
}

func (h *AlertmanagerHandler) GetSpecUID(resource grizzly.Resource) (string, error) {
	return "", fmt.Errorf("GetSpecUID not implemented for PrometheusAlertmanagerConfig")
}

// GetByUID retrieves JSON for a resource from an endpoint, by UID
func (h *AlertmanagerHandler) GetByUID(uid string) (*grizzly.Resource, error) {
	return h.getRemoteAlertmanagerConfig()
}

// GetRemote retrieves an alertmanager config as a Resource
func (h *AlertmanagerHandler) GetRemote(resource grizzly.Resource) (*grizzly.Resource, error) {
	return h.getRemoteAlertmanagerConfig()
}

// ListRemote retrieves as list of UIDs of all remote resources
func (h *AlertmanagerHandler) ListRemote() ([]string, error) {
	_, err := h.getRemoteAlertmanagerConfig()
	if err != nil {
		return nil, err
	}
	resources := []string{"alertmanger-config"}
	return resources, nil
}

// Add pushes an alertmanager config to Mimir via the API
func (h *AlertmanagerHandler) Add(resource grizzly.Resource) error {
	return h.writeAlertmanagerConfig(resource)
}

// Update pushes an alertmanager config to Mimir via the API
func (h *AlertmanagerHandler) Update(existing, resource grizzly.Resource) error {
	return h.writeAlertmanagerConfig(resource)
}

// getRemoteAlertmanagerConfig retrieves an alertmanager config object from Mimir
func (h *AlertmanagerHandler) getRemoteAlertmanagerConfig() (*grizzly.Resource, error) {
	cfg, err := h.clientTool.GetAlertmanagerConfig()
	if err != nil {
		return nil, err
	}

	var spec map[string]interface{}
	cfgYaml, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(cfgYaml, &spec)
	if err != nil {
		return nil, err
	}

	resource, err := grizzly.NewResource(h.APIVersion(), h.Kind(), "alertmanager-config", spec)
	if err != nil {
		return nil, err
	}

	return &resource, nil
}

func (h *AlertmanagerHandler) writeAlertmanagerConfig(resource grizzly.Resource) error {
	newConfig := models.PrometheusAlertmanagerConfig{
		TemplateFiles:      map[string]string{},
		AlertmanagerConfig: "", //resource.Spec()["alertmanager_config"].(string),
	}
	alertmanagerConfigIn, found := resource.Spec()["alertmanager_config"]
	if found {
		alertmanagerConfig, err := yaml.Marshal(alertmanagerConfigIn)
		if err != nil {
			return err
		}
		newConfig.AlertmanagerConfig = string(alertmanagerConfig)
	}

	templateFilesIn, found := resource.Spec()["template_files"]
	if found {
		templateFiles := make(map[string]string)
		for key, in := range templateFilesIn.(map[string]interface{}) {
			templateFiles[key] = in.(string)
		}
		newConfig.TemplateFiles = templateFiles
	}

	return h.clientTool.CreateAlertmangerConfig(newConfig)
}
