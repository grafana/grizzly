package syntheticmonitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/synthetic-monitoring-agent/pkg/pb/synthetic_monitoring"
)

/*
 * @TODO
 * 1. The API does not have a GET method, so we have to fake it here
 * 2. The API expects an ID and a tenantId in an update, but these are
 *    generated by the server so cannot be represented in Jsonnet.
 *    Therefore, we have to pre-retrieve the check to get those values
 *    so we can inject them before posting JSON.
 * 3. This means pre-retrieving the check *twice*, once to establish
 *    whether this resource has changed or not (within Grizzly ifself)
 *    and again within this provider to retrieve IDs. Not ideal.
 * 4. The API expects probes to be specified by ID. This is not
 *    user-friendly. This code therefore takes in strings, and converts
 *    them to IDs, having requested an ID<->string mapping from the API.
 */

type Probes struct {
	ByID   map[int64]synthetic_monitoring.Probe
	ByName map[string]synthetic_monitoring.Probe
}

// SyntheticMonitoringHandler is a Grizzly Handler for Grafana Synthetic Monitoring
type SyntheticMonitoringHandler struct {
	grizzly.BaseHandler
}

// NewSyntheticMonitoringHandler returns a Grizzly Handler for Grafana Synthetic Monitoring
func NewSyntheticMonitoringHandler(provider grizzly.Provider) *SyntheticMonitoringHandler {
	return &SyntheticMonitoringHandler{
		BaseHandler: grizzly.NewBaseHandler(provider, "SyntheticMonitoringCheck", false),
	}
}

const (
	syntheticMonitoringPattern = "synthetic-monitoring/check-%s.%s"
)

// ResourceFilePath returns the location on disk where a resource should be updated
func (h *SyntheticMonitoringHandler) ResourceFilePath(resource grizzly.Resource, filetype string) string {
	return fmt.Sprintf(syntheticMonitoringPattern, resource.Name(), filetype)
}

// Unprepare removes unnecessary elements from a remote resource ready for presentation/comparison
func (h *SyntheticMonitoringHandler) Unprepare(resource grizzly.Resource) *grizzly.Resource {
	resource.DeleteSpecKey("tenantId")
	resource.DeleteSpecKey("id")
	resource.DeleteSpecKey("modified")
	resource.DeleteSpecKey("created")
	return &resource
}

// Prepare gets a resource ready for dispatch to the remote endpoint
func (h *SyntheticMonitoringHandler) Prepare(existing *grizzly.Resource, resource grizzly.Resource) *grizzly.Resource {
	resource.SetSpecValue("tenantId", existing.GetSpecValue("tenantId"))
	resource.SetSpecValue("id", existing.GetSpecValue("id"))
	_, exists := resource.GetSpecString("job")
	if !exists {
		resource.SetSpecString("job", resource.GetMetadata("name"))
	}
	return &resource
}

// Validate returns the uid of resource
func (h *SyntheticMonitoringHandler) Validate(resource grizzly.Resource) error {
	job, exist := resource.GetSpecString("job")
	if exist && job != resource.Name() {
		return fmt.Errorf("job '%s' and name '%s', don't match", job, resource.Name())
	}
	settings := resource.GetSpecValue("settings").(map[string]interface{})
	if _, ok := settings[resource.GetMetadata("type")]; !ok {
		return fmt.Errorf("type '%s' is incorrect", resource.GetMetadata("type"))
	}
	return nil
}

// GetUID returns the UID for a resource
func (h *SyntheticMonitoringHandler) GetUID(resource grizzly.Resource) (string, error) {
	if !resource.HasMetadata("type") {
		return "", fmt.Errorf("%s %s lacks a type metadata element", h.Kind(), resource.Name())
	}
	return fmt.Sprintf("%s.%s", resource.GetMetadata("type"), resource.Name()), nil
}
func (h *SyntheticMonitoringHandler) GetSpecUID(resource grizzly.Resource) (string, error) {
	return "", fmt.Errorf("GetSpecUID not implemented for Synthetic Monitoring")
}

// GetByUID retrieves JSON for a resource from an endpoint, by UID
func (h *SyntheticMonitoringHandler) GetByUID(uid string) (*grizzly.Resource, error) {
	return h.getRemoteCheck(uid)
}

// GetRemote retrieves a datasource as a Resource
func (h *SyntheticMonitoringHandler) GetRemote(resource grizzly.Resource) (*grizzly.Resource, error) {
	uid := fmt.Sprintf("%s.%s", resource.GetMetadata("type"), resource.Name())
	return h.getRemoteCheck(uid)
}

// ListRemote retrieves as list of UIDs of all remote resources
func (h *SyntheticMonitoringHandler) ListRemote() ([]string, error) {
	return h.getRemoteCheckList()
}

// Add adds a new check to the SyntheticMonitoring endpoint
func (h *SyntheticMonitoringHandler) Add(resource grizzly.Resource) error {
	return h.addCheck(resource)
}

// Update pushes an updated check to the SyntheticMonitoring endpoing
func (h *SyntheticMonitoringHandler) Update(existing, resource grizzly.Resource) error {
	return h.updateCheck(resource)
}

// getProbeList retrieves the list of probe and grouped by id and name
func (h *SyntheticMonitoringHandler) getProbeList() (Probes, error) {
	smClient, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return Probes{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	probeList, err := smClient.ListProbes(ctx)
	if err != nil {
		return Probes{}, fmt.Errorf("failed to initialize probes list: %v", err)
	}

	probes := Probes{
		ByID:   map[int64]synthetic_monitoring.Probe{},
		ByName: map[string]synthetic_monitoring.Probe{},
	}

	for _, probe := range probeList {
		if probe.Online && probe.Public {
			probes.ByID[probe.Id] = probe
			probes.ByName[probe.Name] = probe
		}
	}
	return probes, nil
}

// getRemoteCheck retrieves a check object from SM
func (h *SyntheticMonitoringHandler) getRemoteCheckList() ([]string, error) {
	smClient, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	checks, err := smClient.ListChecks(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get checks list: %v", err)
	}
	var checkIDs []string

	for _, check := range checks {
		checkIDs = append(checkIDs, h.getUID(check))
	}
	return checkIDs, nil
}

// getRemoteCheck retrieves a check object from SM
func (h *SyntheticMonitoringHandler) getRemoteCheck(uid string) (*grizzly.Resource, error) {
	smClient, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	checkList, err := smClient.ListChecks(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get checks list: %v", err)
	}

	probes, err := h.getProbeList()
	if err != nil {
		return nil, err
	}

	for _, check := range checkList {
		if h.getUID(check) == uid {
			var probeNames []string
			for _, probeID := range check.Probes {
				probeNames = append(probeNames, probes.ByID[probeID].Name)
			}
			data, err := json.Marshal(check)
			if err != nil {
				return nil, err
			}
			var specmap map[string]interface{}
			err = json.Unmarshal(data, &specmap)
			if err != nil {
				return nil, err
			}
			specmap["probes"] = probeNames
			resource, err := grizzly.NewResource(h.APIVersion(), h.Kind(), check.Job, specmap)
			if err != nil {
				return nil, err
			}
			resource.SetMetadata("type", h.getType(check))
			return &resource, nil
		}
	}
	return nil, grizzly.ErrNotFound
}

func (h *SyntheticMonitoringHandler) convertProbeNameToID(resource *grizzly.Resource) error {
	probes, err := h.getProbeList()
	if err != nil {
		return err
	}
	var probeIDs []int64

	for _, probename := range (*resource).GetSpecValue("probes").([]interface{}) {
		probeName := probename.(string)
		id := probes.ByName[probeName].Id
		probeIDs = append(probeIDs, id)
	}
	(*resource).SetSpecValue("probes", probeIDs)
	return nil
}

func (h *SyntheticMonitoringHandler) addCheck(resource grizzly.Resource) error {
	smClient, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = h.convertProbeNameToID(&resource)
	if err != nil {
		return err
	}

	theCheck, err := h.SpecToCheck(&resource)
	if err != nil {
		return fmt.Errorf("input file is invalid: %v", err)
	}
	_, err = smClient.AddCheck(ctx, theCheck)
	if err != nil {
		return err
	}
	return nil
}

func (h *SyntheticMonitoringHandler) updateCheck(resource grizzly.Resource) error {
	smClient, err := h.Provider.(ClientProvider).Client()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = h.convertProbeNameToID(&resource)
	if err != nil {
		return err
	}

	theCheck, err := h.SpecToCheck(&resource)
	if err != nil {
		return fmt.Errorf("input file is invalid: %v", err)
	}
	_, err = smClient.UpdateCheck(ctx, theCheck)
	if err != nil {
		return err
	}

	return nil
}

func (h *SyntheticMonitoringHandler) SpecToCheck(r *grizzly.Resource) (synthetic_monitoring.Check, error) {
	var smCheck synthetic_monitoring.Check
	data, err := json.Marshal(r.Body["spec"])
	if err != nil {
		return synthetic_monitoring.Check{}, nil
	}

	err = json.Unmarshal(data, &smCheck)
	if err != nil {
		return synthetic_monitoring.Check{}, nil
	}

	return smCheck, nil
}

// Probes allows accessing Probe objects by ID and by name
func (h *SyntheticMonitoringHandler) getType(check synthetic_monitoring.Check) string {
	if check.Settings.Ping != nil {
		return "ping"
	}
	if check.Settings.Http != nil {
		return "http"
	}
	if check.Settings.Tcp != nil {
		return "tcp"
	}
	if check.Settings.Dns != nil {
		return "dns"
	}
	return ""
}

func (h *SyntheticMonitoringHandler) getUID(check synthetic_monitoring.Check) string {
	return fmt.Sprintf("%s.%s", h.getType(check), check.Job)
}
