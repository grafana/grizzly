package grizzly

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"

	"gopkg.in/yaml.v2"
)

// Folder encapsulates a folder object from the Grafana API
type Folder struct {
	Id    int64
	Uid   string
	Title string
}

// Board enscapsulates a dashboard for upload to Grafana API
type Board struct {
	Dashboard map[string]interface{} `json:"dashboard"`
	Overwrite bool                   `json:"overwrite"`
}

func (b Board) String() string {
	data, err := yaml.Marshal(b)
	if err != nil {
		panic(err)
	}

	return string(data)
}

func (b *Board) UnmarshalJSON(data []byte) error {
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	delete(m, "version")
	delete(m, "id")

	b.Dashboard = m
	return nil
}

func (b Board) UID() string {
	return b.Dashboard["uid"].(string)
}

func (b Board) Kind() string {
	return "Dashboard"
}

// Boards encasulates a set of dashboards ready for upload
type Boards map[string]Board

func (bPtr *Boards) UnmarshalJSON(data []byte) error {
	if *bPtr == nil {
		*bPtr = make(Boards)
	}

	var m map[string]Board
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	// check uids missing
	var missing ErrUidsMissing
	for key, board := range m {
		if _, ok := board.Dashboard["uid"]; !ok {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return missing
	}

	for k, v := range m {
		(*bPtr)[k] = v
	}

	// check duplicate uids
	//           uid -> name
	uids := make(map[string]string)
	for name, board := range m {
		has, exist := uids[board.UID()]
		if exist {
			return fmt.Errorf("UID '%s' claimed by '%s' is also used by '%s'. UIDs must be unique.", board.UID(), name, has)
		}
		uids[board.UID()] = name
	}

	return nil
}

type ErrUidsMissing []string

func (e ErrUidsMissing) Error() string {
	return fmt.Sprintf("One or more dashboards have no UID set. UIDs are required for Grizzly to operate properly:\n - %s", strings.Join(e, "\n - "))
}

// GetAPIJSON returns JSON expected by Grafana API
func (b Board) GetAPIJSON() (string, error) {
	b.Overwrite = true
	j, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return "", err
	}
	return string(j), nil
}

// GetDashboardJSON returns JSON representation of a dashboard
func (b Board) GetDashboardJSON() (string, error) {
	j, err := json.MarshalIndent(b.Dashboard, "", "  ")
	if err != nil {
		return "", err
	}
	return string(j), nil
}

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

func getDashboard(config Config, uid string) (*Board, error) {
	if config.GrafanaURL == "" {
		return nil, errors.New("Must set GRAFANA_URL environment variable")
	}

	u, err := url.Parse(config.GrafanaURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, "api/dashboards/uid", uid)

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotFound:
		return nil, ErrNotFound
	default:
		if resp.StatusCode >= 400 {
			return nil, errors.New(resp.Status)
		}
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var board Board
	if err := json.Unmarshal(data, &board); err != nil {
		return nil, APIErr{err, data}
	}

	return &board, nil
}

type APIErr struct {
	err  error
	body []byte
}

func (e APIErr) Error() string {
	return fmt.Sprintf("Failed to parse response: %s.\n\nResponse:\n%s", e.err, string(e.body))
}

func postDashboard(config Config, board Board) error {
	if config.GrafanaURL == "" {
		return errors.New("Must set GRAFANA_URL environment variable")
	}

	u, err := url.Parse(config.GrafanaURL)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, "api/dashboards/db")
	boardJSON, err := board.GetAPIJSON()
	if err != nil {
		return err
	}

	resp, err := http.Post(u.String(), "application/json", bytes.NewBufferString(boardJSON))
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

type PreviewOpts struct {
	ExpiresSeconds int
	// Other properties not yet implemented
	// https://grafana.com/docs/grafana/latest/http_api/snapshot/#create-new-snapshot
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
