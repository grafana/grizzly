package grafana

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/grafana/grafana-openapi-client-go/client/provisioning"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/grafana/grizzly/pkg/grizzly"
)

const KindAlertNotificationTemplate = "AlertNotificationTemplate"

const notificationTemplatePattern = "alert-notification-templates/notificationTemplate-%s.%s"

// AlertNotificationTemplateHandler is a Grizzly Handler for Grafana contactPoints
type AlertNotificationTemplateHandler struct {
	grizzly.BaseHandler
}

// NewAlertNotificationTemplateHandler returns a new Grizzly Handler for Grafana contactPoints
func NewAlertNotificationTemplateHandler(provider grizzly.Provider) *AlertNotificationTemplateHandler {
	return &AlertNotificationTemplateHandler{
		BaseHandler: grizzly.NewBaseHandler(provider, KindAlertNotificationTemplate, false),
	}
}

// ProxyConfigurator provides a configurator object describing how to proxy folders.
func (h *AlertNotificationTemplateHandler) ProxyConfigurator() grizzly.ProxyConfigurator {
	return &alertNotificationTemplateProxyConfigurator{
		provider: h.Provider,
	}
}

// ResourceFilePath returns the location on disk where a resource should be updated
func (h *AlertNotificationTemplateHandler) ResourceFilePath(resource grizzly.Resource, filetype string) string {
	filename := strings.ReplaceAll(resource.Name(), string(os.PathSeparator), "-")
	return fmt.Sprintf(notificationTemplatePattern, filename, filetype)
}

// Prepare gets a resource ready for dispatch to the remote endpoint
func (h *AlertNotificationTemplateHandler) Prepare(existing *grizzly.Resource, resource grizzly.Resource) *grizzly.Resource {
	if !resource.HasSpecString("name") {
		resource.SetSpecString("name", resource.Name())
	}

	return &resource
}

// Unprepare removes unnecessary elements from a remote resource ready for presentation/comparison
func (h *AlertNotificationTemplateHandler) Unprepare(resource grizzly.Resource) *grizzly.Resource {
	resource.DeleteSpecKey("version")
	return &resource
}

func (h *AlertNotificationTemplateHandler) Validate(resource grizzly.Resource) error {
	name, exist := resource.GetSpecString("name")
	if resource.Name() != name && exist {
		return fmt.Errorf("spec.name '%s' and metadata.name '%s', don't match", name, resource.Name())
	}
	return nil
}

func (h *AlertNotificationTemplateHandler) GetSpecUID(resource grizzly.Resource) (string, error) {
	name, ok := resource.GetSpecString("name")
	if !ok {
		return "", fmt.Errorf("name not specified")
	}
	return name, nil
}

func (h *AlertNotificationTemplateHandler) GetByUID(uid string) (*grizzly.Resource, error) {
	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return nil, err
	}

	response, err := client.Provisioning.GetTemplate(uid)
	if err != nil {
		var gErr *provisioning.GetTemplateNotFound
		if errors.As(err, &gErr) {
			return nil, grizzly.ErrNotFound
		}
		return nil, err
	}

	spec, err := structToMap(response.GetPayload())
	if err != nil {
		return nil, err
	}

	resource, err := grizzly.NewResource(h.APIVersion(), h.Kind(), uid, spec)
	if err != nil {
		return nil, err
	}

	return &resource, nil
}

// GetRemote retrieves a contactPoint as a Resource
func (h *AlertNotificationTemplateHandler) GetRemote(resource grizzly.Resource) (*grizzly.Resource, error) {
	return h.GetByUID(resource.Name())
}

// ListRemote retrieves as list of UIDs of all remote resources
func (h *AlertNotificationTemplateHandler) ListRemote() ([]string, error) {
	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return nil, err
	}

	response, err := client.Provisioning.GetTemplates()
	if err != nil {
		return nil, err
	}
	templates := response.GetPayload()
	uids := make([]string, 0, len(templates))
	for _, template := range templates {
		uids = append(uids, template.Name)
	}
	return uids, nil
}

// Add pushes a contactPoint to Grafana via the API
func (h *AlertNotificationTemplateHandler) Add(resource grizzly.Resource) error {
	client, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return err
	}

	templateBody, _ := resource.GetSpecString("template")

	params := provisioning.NewPutTemplateParams().
		WithName(resource.Name()).
		WithBody(&models.NotificationTemplateContent{
			Template: templateBody,
		}).
		WithXDisableProvenance(&stringtrue)
	_, err = client.Provisioning.PutTemplate(params)
	return err
}

// Update pushes a contactPoint to Grafana via the API
func (h *AlertNotificationTemplateHandler) Update(existing, resource grizzly.Resource) error {
	// Add calls the "PUT" endpoint, allowing us to create or update a template.
	return h.Add(resource)
}
