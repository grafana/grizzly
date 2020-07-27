package grizzly

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
    "context"
	"golang.org/x/oauth2"
	"net/http"
	"net/url"
	"path"
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
	FolderID  int                    `json:"folderId"`
	Overwrite bool                   `json:"overwrite"`
	UID       string
	Name      string
}

// Boards encasulates a set of dashboards ready for upload
type Boards map[string]Board

// http encapsulates the HTTP Client that uses our GrafanaToken
func grafanaHttpClient(config Config) (*http.Client) {
    ctx := context.Background()
    client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{
                AccessToken: config.GrafanaToken,
                TokenType:   "Bearer",
    }))
    return client
}

// GetAPIJSON returns JSON expected by Grafana API
func (b Board) GetAPIJSON() (string, error) {
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

func parseDashboards(raw string) (Boards, error) {
	var boards Boards
	if err := json.Unmarshal([]byte(raw), &boards); err != nil {
		return nil, err
	}
	newBoards := make(Boards)
	for key, board := range boards {
		board.UID = fmt.Sprintf("%v", board.Dashboard["uid"])
		board.Name = key
		newBoards[key] = board
	}
	return newBoards, nil
}

func parseDashboard(raw string) (*Board, error) {
	var board Board
	if err := json.Unmarshal([]byte(raw), &board); err != nil {
		return nil, err
	}
	board.UID = fmt.Sprintf("%v", board.Dashboard["uid"])
	return &board, nil
}

func searchFolder(config Config, name string) (*Folder, error) {
    http := grafanaHttpClient(config)


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


    http := grafanaHttpClient(config)

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
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 404 {
		return nil, ErrNotFound
	}
	if resp.StatusCode >= 400 {
		return nil, errors.New(resp.Status)
	}
	board, err := parseDashboard(string(body))
	if err != nil {
		return nil, err
	}
	return board, nil
}

func postDashboard(config Config, board Board) error {

	if config.GrafanaURL == "" {
		return errors.New("Must set GRAFANA_URL environment variable")
	}


    http := grafanaHttpClient(config)

	u, err := url.Parse(config.GrafanaURL)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, "api/dashboards/db")
	boardJSON, err := board.GetAPIJSON()
	if err != nil {
		return err
	}
	resp, err := http.Post(u.String(), "application/json", bytes.NewBuffer([]byte(boardJSON)))
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Non-200 response from Grafana: %s", resp.Status)
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


    http := grafanaHttpClient(config)

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
