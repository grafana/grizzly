package grafana

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
)

// DashboardProvider is a Grizzly Provider for Grafana dashboards
type DashboardProvider struct{}

// NewDashboardProvider returns configuration defining a new Grafana Provider
func NewDashboardProvider() *DashboardProvider {
	return &DashboardProvider{}
}

// GetName returns the name for this provider
func (p *DashboardProvider) GetName() string {
	return "grafana"
}

// GetJSONPath returns a paths within Jsonnet output that this provider will consume
func (p *DashboardProvider) GetJSONPath() string {
	return "grafanaDashboards"
}

// GetExtension returns the file name extension for a dashboard
func (p *DashboardProvider) GetExtension() string {
	return "json"
}

func (p *DashboardProvider) newDashboardResource(uid, filename string, board Dashboard) grizzly.Resource {
	resource := grizzly.Resource{
		UID:      uid,
		Filename: filename,
		Provider: p,
		Detail:   board,
		Path:     p.GetJSONPath(),
	}
	return resource
}

// Parse parses an interface{} object into a struct for this resource type
func (p *DashboardProvider) Parse(i interface{}) (grizzly.Resources, error) {
	resources := grizzly.Resources{}
	msi := i.(map[string]interface{})
	for k, v := range msi {
		board := Dashboard{}
		err := mapstructure.Decode(v, &board)
		if err != nil {
			return nil, err
		}
		resource := p.newDashboardResource(board.UID(), k, board)
		key := resource.Key()
		resources[key] = resource
	}
	return resources, nil
}

// GetByUID retrieves JSON for a resource from an endpoint, by UID
func (p *DashboardProvider) GetByUID(UID string) (*grizzly.Resource, error) {
	board, err := getRemoteDashboard(UID)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving dashboard %s: %v", UID, err)
	}
	resource := p.newDashboardResource(UID, "", *board)
	return &resource, nil
}

// GetRepresentation renders a resource as JSON or YAML as appropriate
func (p *DashboardProvider) GetRepresentation(uid string, detail map[string]interface{}) (string, error) {
	j, err := json.MarshalIndent(detail, "", "  ")
	if err != nil {
		return "", err
	}
	return string(j), nil
}

// GetRemoteRepresentation retrieves a dashboard as JSON
func (p *DashboardProvider) GetRemoteRepresentation(uid string) (string, error) {
	board, err := getRemoteDashboard(uid)

	if err != nil {
		return "", err
	}
	return board.toJSON()
}

// Apply pushes a dashboard to Grafana via the API
func (p *DashboardProvider) Apply(detail map[string]interface{}) error {
	board := Dashboard(detail)

	// @TODO SUPPORT FOLDERS!!

	uid := board.UID()
	existingBoard, err := getRemoteDashboard(uid)

	switch err {
	case grizzly.ErrNotFound: // create new
		fmt.Println(uid, grizzly.Green("added"))
		if err := postDashboard(board); err != nil {
			return err
		}
	case nil: // update
		boardJSON, _ := board.toJSON()
		existingBoardJSON, _ := existingBoard.toJSON()

		if boardJSON == existingBoardJSON {
			fmt.Println(uid, grizzly.Yellow("unchanged"))
			return nil
		}

		if err = postDashboard(board); err != nil {
			return err
		}
		log.Println(uid, grizzly.Green("updated"))

	default: // failed
		return fmt.Errorf("Error retrieving dashboard %s: %v", uid, err)
	}
	return nil
}

// Preview renders Jsonnet then pushes them to the endpoint if previews are possible
func (p *DashboardProvider) Preview(detail map[string]interface{}) error {
	return nil
}

/*
// Preview renders Jsonnet then pushes them to the endpoint if previews are possible
func (p *DashboardProvider) Preview(resource grizzly.Resource, opts *PreviewOpts) error {
	//folderID is not used in snapshots
	folderID := int64(0)
	boards, err := renderDashboards(jsonnetFile, targets, folderID)
	if err != nil {
		return err
	}
	for _, board := range boards {
		uid := board.UID()
		s, err := postSnapshot(config, board, opts)
		if err != nil {
			return err
		}
		fmt.Println("View", uid, green(s.URL))
		fmt.Println("Delete", uid, yellow(s.DeleteURL))
	}
	if opts.ExpiresSeconds > 0 {
		fmt.Print(yellow(fmt.Sprintf("Previews will expire and be deleted automatically in %d seconds\n", opts.ExpiresSeconds)))
	}
	return nil
}
*/
///////////////////////////////////////////////////////////////////////////

/*
func searchFolder(config Config, name string) (*Folder, error) {
	if config.GrafanaURL == "" {
		return nil, errors.New("Must set GRAFANA_URL environment variable")
	}

	u, err := url.Parse(config.GrafanaURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, "api/search")
	u.Query().Add("query", name)
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var folders []Folder
	if err := json.Unmarshal([]byte(string(body)), &folders); err != nil {
		return nil, err
	}
	var folder Folder
	for _, f := range folders {
		if f.Title == name {
			folder = f
			break
		}
	}
	return &folder, nil
}
*/

// getRemoteDashboard retrieves a dashboard object from Grafana
func getRemoteDashboard(uid string) (*Dashboard, error) {
	grafanaURL, err := getGrafanaURL("api/dashboards/uid/" + uid)
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(grafanaURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotFound:
		return nil, grizzly.ErrNotFound
	default:
		if resp.StatusCode >= 400 {
			return nil, errors.New(resp.Status)
		}
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var d DashboardWrapper
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, APIErr{err, data}
	}
	delete(d.Dashboard, "id")
	delete(d.Dashboard, "version")
	return &d.Dashboard, nil
}

func postDashboard(board Dashboard) error {
	grafanaURL, err := getGrafanaURL("api/dashboards/db")
	if err != nil {
		return err
	}

	// @TODO support folders:
	folderID := 0
	wrappedBoard := wrapDashboard(folderID, board)
	wrappedJSON, err := wrappedBoard.toJSON()

	resp, err := http.Post(grafanaURL, "application/json", bytes.NewBufferString(wrappedJSON))
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		break
	case http.StatusPreconditionFailed:
		d := json.NewDecoder(resp.Body)
		var r struct {
			Message string `json:"message"`
		}
		if err := d.Decode(&r); err != nil {
			return fmt.Errorf("Failed to decode actual error (412 Precondition failed): %s", err)
		}
		fmt.Println(wrappedJSON)
		return fmt.Errorf("Error while applying '%s' to Grafana: %s", board.UID(), r.Message)
	default:
		return fmt.Errorf("Non-200 response from Grafana while applying '%s': %s", resp.Status, board.UID())
	}

	return nil
}

/*
type SnapshotResp struct {
	DeleteKey string `json:"deleteKey"`
	DeleteURL string `json:"deleteUrl"`
	Key       string `json:"key"`
	URL       string `json:"url"`
}

type SnapshotReq struct {
	Dashboard map[string]interface{} `json:"dashboard"`
	Expires   int                    `json:"expires,omitempty"`
}

func postSnapshot(config Config, board Board, opts *PreviewOpts) (*SnapshotResp, error) {
	if config.GrafanaURL == "" {
		return nil, errors.New("Must set GRAFANA_URL environment variable")
	}

	u, err := url.Parse(config.GrafanaURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, "api/snapshots")

	sr := &SnapshotReq{
		Dashboard: board.Dashboard,
	}

	if opts.ExpiresSeconds > 0 {
		sr.Expires = opts.ExpiresSeconds
	}

	bs, err := json.Marshal(&sr)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(u.String(), "application/json", bytes.NewBuffer(bs))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Non-200 response from Grafana: %s", resp.Status)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Unable to read response body: %w", err)
	}
	s := &SnapshotResp{}
	err = json.Unmarshal(b, s)
	if err != nil {
		return nil, fmt.Errorf("Unable to unmarshal response body into SnapshotResp: %w", err)
	}
	return s, nil
}

func folderId(config Config, jsonnetFile string) (*int64, error) {
	jsonnet := fmt.Sprintf(`
local f = import "%s";
f.grafanaDashboardFolder`, jsonnetFile)
	output, err := evalToString(jsonnet)
	if err != nil {
		return nil, err
	}
	var name string
	err = json.Unmarshal([]byte(output), &name)
	if err != nil {
		return nil, err
	}
	folder, err := searchFolder(config, strings.TrimSpace(name))
	if err != nil {
		return nil, err
	}
	return &folder.Id, nil
}
*/

// Folder encapsulates a folder object from the Grafana API
type Folder struct {
	ID    int64
	UID   string
	Title string
}

// Dashboard encapsulates a dashboard
type Dashboard map[string]interface{}

// UID retrieves the UID from a dashboard
func (d *Dashboard) UID() string {
	uid, ok := (*d)["uid"]
	if !ok {
		return ""
	}
	return uid.(string)
}

// toJSON returns JSON for a dashboard
func (d *Dashboard) toJSON() (string, error) {
	j, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return "", err
	}
	return string(j), nil
}

// DashboardWrapper adds wrapper required by Grafana API
type DashboardWrapper struct {
	Dashboard Dashboard `json:"dashboard"`
	FolderID  int       `json:"folderId"`
	Overwrite bool      `json:"overwrite"`
}

func wrapDashboard(folderID int, dashboard Dashboard) DashboardWrapper {
	wrapper := DashboardWrapper{
		Dashboard: dashboard,
		FolderID:  folderID,
		Overwrite: true,
	}
	return wrapper
}

func (d DashboardWrapper) String() string {
	data, err := yaml.Marshal(d)
	if err != nil {
		panic(err)
	}

	return string(data)
}

// UID retrieves the UID from a dashboard wrapper
func (d *DashboardWrapper) UID() string {
	return d.Dashboard.UID()
}

// toJSON returns JSON expected by Grafana API
func (d *DashboardWrapper) toJSON() (string, error) {
	d.Overwrite = true
	j, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return "", err
	}
	return string(j), nil
}
