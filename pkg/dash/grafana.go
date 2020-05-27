package dash

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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
}

// Boards encasulates a set of dashboards ready for upload
type Boards map[string]Board

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
