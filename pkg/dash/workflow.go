package dash

import (
	"encoding/json"
	"fmt"

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

// Show renders a Jsonnet dashboard as JSON, consuming a jsonnet filename
func Show(config Config, jsonnetFile string, targets *[]string) error {
	boards, err := renderDashboards(jsonnetFile, targets)
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
	boards, err := renderDashboards(jsonnetFile, targets)
	if err != nil {
		return err
	}

	for name, board := range boards {
		fmt.Printf("\n== %s ==\n", name)
		normalize(board)

		existingBoard, err := getDashboard(config, board.UID)
		if err != nil {
			return fmt.Errorf("Error retrieving dashboard %s: %v", name, err)
		}
		normalize(*existingBoard)

		boardJSON, _ := board.GetDashboardJSON()
		existingBoardJSON, _ := existingBoard.GetDashboardJSON()

		if boardJSON == existingBoardJSON {
			fmt.Println("No differences")
		} else {
			difference := diff.Diff(existingBoardJSON, boardJSON)
			fmt.Println(difference)
		}
	}
	return nil
}

// Apply renders a Jsonnet dashboard then pushes it to Grafana via the API
func Apply(config Config, jsonnetFile string, targets *[]string) error {
	boards, err := renderDashboards(jsonnetFile, targets)
	if err != nil {
		return err
	}

	for name, board := range boards {
		fmt.Printf("\n== %s ==\n", name)

		err = postDashboard(config, board)
		if err != nil {
			return err
		}
	}
	return nil
}

func renderDashboards(jsonnetFile string, targets *[]string) (Boards, error) {
	t := []byte("[]")
	if len(*targets) > 0 {
		t, _ = json.Marshal(targets)
	}
	jsonnet := fmt.Sprintf(`
local f = import "%s";
local t = %s;
{
  [k]: { dashboard: f.grafanaDashboards[k], folderId: 0, overwrite: true}
  for k in std.filter(
		function(n) if std.length(t) > 0 then std.member(t, n) else true,
		std.objectFields(f.grafanaDashboards)
	)
}`, jsonnetFile, t)
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
