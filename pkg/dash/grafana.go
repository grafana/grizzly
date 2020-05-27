package dash

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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
	url := fmt.Sprintf("%s/api/search?query=%s", config.GrafanaURL, name)
	resp, err := http.Get(url)
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

	url := fmt.Sprintf("%s/api/dashboards/uid/%s", config.GrafanaURL, uid)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	board, err := parseDashboard(string(body))
	if err != nil {
		return nil, err
	}
	return board, nil
}

func postDashboard(config Config, board Board) error {
	url := fmt.Sprintf("%s/api/dashboards/db", config.GrafanaURL)

	boardJSON, err := board.GetAPIJSON()
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer([]byte(boardJSON)))
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Non-200 response from Grafana: %s", resp.Status)
	}
	return nil
}
