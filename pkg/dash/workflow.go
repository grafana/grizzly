package dash

import (
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

// Show renders a Jsonnet dashboard as JSON, consuming a jsonnet filename
func Show(config Config, jsonnetFile string) error {
  boards, err := renderDashboards(jsonnetFile)
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
func Diff(config Config, jsonnetFile string) error {
  boards, err := renderDashboards(jsonnetFile)
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
func Apply(config Config, jsonnetFile string) error {
  boards, err := renderDashboards(jsonnetFile)
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

func renderDashboards(jsonnetFile string) (Boards, error) {
  template:=`
  local f = import "{{FILE}}";
  {
    [k]: { dashboard: f.grafanaDashboards[k], folderId: 0, overwrite: true}
    for k in std.objectFields(f.grafanaDashboards)
  }
  `
  jsonnet := strings.ReplaceAll(template, "{{FILE}}", jsonnetFile)
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