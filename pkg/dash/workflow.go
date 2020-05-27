package dash

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kylelemons/godebug/diff"
)

// Get retrieves JSON for a dashboard from Grafana, using the dashboard's UID
func Get(config Config, dashboardUID string) error {
	board, err := getDashboard(config, dashboardUID)
	if err != nil {
		return fmt.Errorf("Error retrieving dashboard %s: %v", dashboardUID, err)
	}
	dashboardJSON, _ := board.GetDashboardJSON()
	fmt.Println(dashboardJSON)
	return nil
}

// List outputs the keys of the grafanaDashboards object.
func List(jsonnetFile string) error {
	keys, err := dashboardKeys(jsonnetFile)
	if err != nil {
		return err
	}
	fmt.Println(strings.Join(keys, "\n"))
	return nil
}

// Show renders a Jsonnet dashboard as JSON, consuming a jsonnet filename
func Show(config Config, jsonnetFile string, targets *[]string) error {
	boards, err := renderDashboards(jsonnetFile, targets, 0)
	if err != nil {
		return err
	}

	for name, board := range boards {
		fmt.Printf("\n== %s ==\n", name)
		j, err := board.GetDashboardJSON()
		if err != nil {
			return err
		}
		fmt.Println(j)
	}
	return nil
}

func normalize(board Board) {
	board.Dashboard["version"] = nil
	board.Dashboard["id"] = nil
}

// Diff renders a Jsonnet dashboard and compares it with what is found in Grafana
func Diff(config Config, jsonnetFile string, targets *[]string) error {
	boards, err := renderDashboards(jsonnetFile, targets, 0)
	if err != nil {
		return err
	}

	for name, board := range boards {
		normalize(board)

		existingBoard, err := getDashboard(config, board.UID)
		if err != nil {
			return fmt.Errorf("Error retrieving dashboard %s: %v", name, err)
		}
		normalize(*existingBoard)

		boardJSON, _ := board.GetDashboardJSON()
		existingBoardJSON, _ := existingBoard.GetDashboardJSON()

		if boardJSON == existingBoardJSON {
			fmt.Println(name, "no differences")
		} else {
			fmt.Println(name, "changes detected:")
			difference := diff.Diff(existingBoardJSON, boardJSON)
			fmt.Println(difference)
		}
	}
	return nil
}

// Apply renders a Jsonnet dashboard then pushes it to Grafana via the API
func Apply(config Config, jsonnetFile string, targets *[]string) error {
	folderID, err := folderId(config, jsonnetFile)
	if err != nil {
		var fID int64 = 0
		folderID = &fID
		fmt.Println("Folder not found and/or configured. Applying to \"General\" folder.")
	}
	boards, err := renderDashboards(jsonnetFile, targets, *folderID)
	if err != nil {
		return err
	}
	for name, board := range boards {
		normalize(board)
		existingBoard, err := getDashboard(config, board.UID)
		if err != nil {
			return fmt.Errorf("Error retrieving dashboard %s: %v", name, err)
		}
		normalize(*existingBoard)

		boardJSON, _ := board.GetDashboardJSON()
		existingBoardJSON, _ := existingBoard.GetDashboardJSON()

		if boardJSON == existingBoardJSON {
			fmt.Println(name, "unchanged")
		} else {
			fmt.Println(name, "updated")

			err = postDashboard(config, board)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func dashboardKeys(jsonnetFile string) ([]string, error) {
	jsonnet := fmt.Sprintf(`
local f = import "%s";
std.objectFields(f.grafanaDashboards)`, jsonnetFile)
	output, err := evalToString(jsonnet)
	if err != nil {
		return nil, err
	}
	var keys []string
	err = json.Unmarshal([]byte(output), &keys)
	if err != nil {
		return nil, err
	}
	return keys, nil
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

func renderDashboards(jsonnetFile string, targets *[]string, folderId int64) (Boards, error) {
	t := []byte("[]")
	if len(*targets) > 0 {
		t, _ = json.Marshal(targets)
	}
	jsonnet := fmt.Sprintf(`
local f = import "%s";
local t = %s;
{
  [k]: { dashboard: f.grafanaDashboards[k], folderId: %d, overwrite: true}
  for k in std.filter(
    function(n) if std.length(t) > 0 then std.member(t, n) else true,
    std.objectFields(f.grafanaDashboards)
  )
}`, jsonnetFile, t, folderId)
	output, err := evalToString(jsonnet)
	if err != nil {
		return nil, err
	}
	boards, err := parseDashboards(output)
	if err != nil {
		return nil, err
	}
	return boards, nil
}
