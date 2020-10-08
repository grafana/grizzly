package grizzly

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sort"
	"strings"
	"text/tabwriter"

	rulefmt "github.com/cortexproject/cortex/pkg/ruler/legacy_rulefmt"
	"github.com/fatih/color"
	"github.com/google/go-jsonnet"
	"github.com/grafana/grizzly/pkg/term"
	"github.com/kylelemons/godebug/diff"
	"golang.org/x/crypto/ssh/terminal"
	"gopkg.in/fsnotify.v1"
	"gopkg.in/yaml.v2"
)

var interactive = terminal.IsTerminal(int(os.Stdout.Fd()))

var (
	red    = color.New(color.FgRed).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
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
	res, err := parse(jsonnetFile)
	if err != nil {
		return err
	}

	f := "%s\t%s\n"
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)

	fmt.Fprintf(w, f, "KIND", "NAME")
	for _, r := range res {
		fmt.Fprintf(w, f, r.Kind(), r.UID())
	}

	return w.Flush()
}

type Group rulefmt.RuleGroup

func (g Group) Kind() string {
	return "Group"
}

func (g Group) UID() string {
	return g.Name
}

// UnmarshalJSON uses the YAML parser for this specific type, because the
// embedded prometheus types require this.
func (g *Group) UnmarshalJSON(data []byte) error {
	return yaml.Unmarshal(data, g)
}

type Rules struct {
	Groups []Group `json:"groups"`
}

type Mixin struct {
	Dashboards Boards `json:"grafanaDashboards"`
	Rules      Rules  `json:"prometheusRules"`
	Alerts     Rules  `json:"prometheusAlerts"`
}

func eval(jsonnetFile string) ([]byte, error) {
	const template = "(import '%s') + { grafanaDashboards+:::{} }"
	script := fmt.Sprintf(template, jsonnetFile)
	fmt.Println("SCRIPT", script)
	vm := jsonnet.MakeVM()
	vm.Importer(newExtendedImporter([]string{"vendor", "lib", "."}))

	result, err := vm.EvaluateSnippet(jsonnetFile, script)
	if err != nil {
		return nil, err
	}

	return []byte(result), nil
}

func parse(jsonnetFile string) (Resources, error) {
	data, err := eval(jsonnetFile)
	if err != nil {
		return nil, err
	}

	var m Mixin
	if err := json.Unmarshal([]byte(data), &m); err != nil {
		return nil, err
	}

	// Destructure Mixin into Resources slice
	var r Resources
	for _, b := range m.Dashboards {
		r = append(r, b)
	}

	for _, g := range m.Rules.Groups {
		r = append(r, g)
	}

	for _, g := range m.Alerts.Groups {
		r = append(r, g)
	}

	return r, nil
}

// Show renders a Jsonnet dashboard as JSON, consuming a jsonnet filename
func Show(config Config, jsonnetFile string, targets []string) error {
	res, err := parse(jsonnetFile)
	if err != nil {
		return err
	}

	if interactive && len(res) >= 2 {
		var items []term.PageItem
		for _, r := range res {
			items = append(items, term.PageItem{
				Name:    fmt.Sprintf("%s/%s", r.Kind(), r.UID()),
				Content: mustYAML(r),
			})
		}
		return term.Page(items)
	}

	fmt.Print(res.String())
	return nil
}

func mustYAML(i interface{}) string {
	data, err := yaml.Marshal(i)
	if err != nil {
		panic(err)
	}
	return string(data)
}

// Diff renders a Jsonnet dashboard and compares it with what is found in Grafana
func Diff(config Config, jsonnetFile string, targets []string) error {
	boards, err := renderDashboards(jsonnetFile, targets, 0)
	if err != nil {
		return err
	}

	for _, board := range boards {
		uid := board.UID()
		existingBoard, err := getDashboard(config, board.UID())
		if err == ErrNotFound {
			log.Println(uid, yellow("not present in Grafana"))
			continue
		}
		if err != nil {
			return fmt.Errorf("Error retrieving dashboard %s: %v", uid, err)
		}

		boardJSON, _ := board.GetDashboardJSON()
		existingBoardJSON, _ := existingBoard.GetDashboardJSON()

		if boardJSON == existingBoardJSON {
			fmt.Println(uid, yellow("no differences"))
		} else {
			fmt.Println(uid, red("changes detected:"))
			difference := diff.Diff(existingBoardJSON, boardJSON)
			fmt.Println(difference)
		}
	}
	return nil
}

// Apply renders Jsonnet dashboards then pushes them to Grafana via the API
func Apply(config Config, jsonnetFile string, targets []string) error {
	folderID, err := folderId(config, jsonnetFile)
	if err != nil {
		return err
	}
	boards, err := renderDashboards(jsonnetFile, targets, *folderID)
	if err != nil {
		return err
	}
	for _, k := range boardKeys(boards) {
		board := boards[k]

		uid := board.UID()
		existingBoard, err := getDashboard(config, uid)

		switch err {
		case ErrNotFound: // create new
			fmt.Println(uid, green("added"))
			if err := postDashboard(config, board); err != nil {
				return err
			}
		case nil: // update
			boardJSON, _ := board.GetAPIJSON()
			existingBoardJSON, _ := existingBoard.GetAPIJSON()

			if boardJSON == existingBoardJSON {
				fmt.Println(uid, yellow("unchanged"))
				continue
			}

			if err = postDashboard(config, board); err != nil {
				return err
			}
			log.Println(uid, green("updated"))

		default: // failed
			return fmt.Errorf("Error retrieving dashboard %s: %v", uid, err)
		}
	}
	return nil
}

func boardKeys(b Boards) (keys []string) {
	for k := range b {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Preview renders Jsonnet dashboards then pushes them to Grafana via the Snapshot API
func Preview(config Config, jsonnetFile string, targets []string, opts *PreviewOpts) error {
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

// Watch watches a directory for changes then pushes Jsonnet dashboards to Grafana
// when changes are noticed
func Watch(config Config, watchDir, jsonnetFile string, targets []string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					config, err := ParseEnvironment()
					if err != nil {
						log.Println("error:", err)
					}
					log.Println("Changes detected. Applying", jsonnetFile)
					if err := Apply(*config, jsonnetFile, targets); err != nil {
						log.Println("error:", err)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(watchDir)
	if err != nil {
		return err
	}
	<-done
	return nil
}

// Export renders Jsonnet dashboards then saves them to a directory
func Export(config Config, jsonnetFile, dashboardDir string, targets []string) error {
	boards, err := renderDashboards(jsonnetFile, targets, 0)
	if err != nil {
		return err
	}

	for _, board := range boards {
		uid := board.UID()
		boardJSON, err := board.GetDashboardJSON()
		if err != nil {
			return err
		}
		boardPath := path.Join(dashboardDir, uid)
		if !strings.HasSuffix(uid, ".json") {
			boardPath += ".json"
		}
		existingBoardJSONBytes, err := ioutil.ReadFile(boardPath)
		isNotExist := os.IsNotExist(err)
		if err != nil && !isNotExist {
			return err
		}
		existingBoardJSON := string(existingBoardJSONBytes)

		err = ioutil.WriteFile(boardPath, []byte(boardJSON), 0644)
		if err != nil {
			return err
		}

		if isNotExist {
			fmt.Println(uid, green("added"))
		} else if boardJSON == existingBoardJSON {
			fmt.Println(uid, yellow("unchanged"))
		} else {
			fmt.Println(uid, green("updated"))
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
	name, err := evalToString(jsonnet)
	if err != nil {
		return nil, err
	}
	name = strings.TrimSpace(strings.ReplaceAll(name, "\"", ""))
	folder, err := getFolder(config, name)
	if err != nil {
		return nil, err
	}
	return &folder, nil
}

func renderDashboards(jsonnetFile string, targets []string, folderId int64) (Boards, error) {
	jsonnet := fmt.Sprintf(`(import "%s").grafanaDashboards`, jsonnetFile)
	data, err := evalToString(jsonnet)
	if err != nil {
		return nil, err
	}

	var _boards Boards
	if err := json.Unmarshal([]byte(data), &_boards); err != nil {
		return nil, err
	}

	boards := make(map[string]Board)

	for k, b := range _boards {
		b.FolderID = folderId
		boards[k] = b
	}

	if len(targets) == 0 {
		return boards, nil
	}

	// TODO(sh0rez): use process.Matcher of Tanka instead
Outer:
	for key, b := range boards {
		uid := b.UID()
		for _, t := range targets {
			if t == uid {
				continue Outer
			}
		}
		delete(boards, key)
	}

	return boards, nil
}
