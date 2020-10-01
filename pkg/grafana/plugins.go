package grafana

/*
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
)

// PluginProvider is a Grizzly Provider for Grafana Plugins
type PluginProvider struct{}

// NewPluginProvider returns configuration defining a new Grafana Provider
func NewPluginProvider() *PluginProvider {
	return &PluginProvider{}
}

// GetName returns the name for this provider
func (p *PluginProvider) GetName() string {
	return "grafana"
}

// GetJSONPath returns a paths within Jsonnet output that this provider will consume
func (p *PluginProvider) GetJSONPath() string {
	return "grafanaPlugins"
}

// GetExtension returns the file name extension for a Plugin
func (p *PluginProvider) GetExtension() string {
	return "json"
}

func (p *PluginProvider) newPluginResource(uid string, board interface{}) grizzly.Resource {
	resource := grizzly.Resource{
		UID:      uid,
		Provider: p,
		Resource: board,
		Path:     p.GetJSONPath(),
	}
	return resource
}

// Parse parses an interface{} object into a struct for this resource type
func (p *PluginProvider) Parse(i interface{}) (grizzly.Resources, error) {
	resources := grizzly.Resources{}
	msi := i.(map[string]interface{})
	for k, v := range msi {
		board := Board{}
		err := mapstructure.Decode(v, &board)
		if err != nil {
			return nil, err
		}
		resource := p.newPluginResource(k, board)
		key := resource.Key()
		resources[key] = resource
	}
	return resources, nil
}

// GetByUID retrieves JSON for a resource from an endpoint, by UID
func (p *PluginProvider) GetByUID(UID string) (*grizzly.Resource, error) {
	board, err := getRemotePlugin(UID)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving Plugin %s: %v", UID, err)
	}
	resource := p.newPluginResource(UID, board)
	return &resource, nil
}

// GetRepresentation renders a resource as JSON or YAML as appropriate
func (p *PluginProvider) GetRepresentation(uid string, res interface{}) (string, error) {
	board := res.(*Board)
	return board.toJSON()
}

// GetRemoteRepresentation retrieves a Plugin as JSON
func (p *PluginProvider) GetRemoteRepresentation(uid string) (string, error) {
	board, err := getRemotePlugin(uid)
	if err != nil {
		return "", err
	}
	return board.toJSON()
}

// Apply pushes a Plugin to Grafana via the API
func (p *PluginProvider) Apply(resource grizzly.Resource) error {
	board := resource.Resource.(Board)

	// @TODO SUPPORT FOLDERS!!

	uid := board.UID()
	existingBoard, err := getRemotePlugin(uid)

	switch err {
	case grizzly.ErrNotFound: // create new
		fmt.Println(uid, grizzly.Green("added"))
		if err := postPlugin(board); err != nil {
			return err
		}
	case nil: // update
		boardJSON, _ := board.toJSON()
		existingBoardJSON, _ := existingBoard.toJSON()

		if boardJSON == existingBoardJSON {
			fmt.Println(uid, grizzly.Yellow("unchanged"))
			return nil
		}

		if err = postPlugin(board); err != nil {
			return err
		}
		log.Println(uid, grizzly.Green("updated"))

	default: // failed
		return fmt.Errorf("Error retrieving Plugin %s: %v", uid, err)
	}
	return nil
}

// Preview renders Jsonnet then pushes them to the endpoint if previews are possible
func (p *PluginProvider) Preview(resource grizzly.Resource) error {
	return nil
}

///////////////////////////////////////////////////////////////////////////

// getRemotePlugin retrieves a Plugin object from Grafana
func getRemotePlugin(uid string) (*Board, error) {
	grafanaURL, err := getGrafanaURL("api/Plugins/uid/" + uid)
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

	type nestedBoard struct {
		Plugin Board `json:"Plugin"`
	}
	var b nestedBoard
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, APIErr{err, data}
	}

	board := Board{Plugin: b.Plugin.Plugin}

	return &board, nil
}

func postPlugin(board Board) error {
	grafanaURL, err := getGrafanaURL("api/Plugins/db")
	if err != nil {
		return err
	}
	boardJSON, err := board.GetAPIJSON()
	if err != nil {
		return err
	}

	resp, err := http.Post(grafanaURL, "application/json", bytes.NewBufferString(boardJSON))
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
		fmt.Println(boardJSON)
		return fmt.Errorf("Error while applying '%s' to Grafana: %s", board.UID(), r.Message)
	default:
		return fmt.Errorf("Non-200 response from Grafana while applying '%s': %s", resp.Status, board.UID())
	}

	return nil
}
*/
