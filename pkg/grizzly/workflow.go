package grizzly

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/google/go-jsonnet"
	"github.com/grafana/grizzly/pkg/term"
	"github.com/kylelemons/godebug/diff"
	"github.com/malcolmholmes/grizzly/pkg/grizzly"
	"golang.org/x/crypto/ssh/terminal"
	"gopkg.in/fsnotify.v1"
)

var interactive = terminal.IsTerminal(int(os.Stdout.Fd()))

// Get retrieves JSON for a dashboard from Grafana, using the dashboard's UID
func Get(config Config, UID string) error {
	if !strings.Contains(UID, ".") {
		return fmt.Errorf("UID must be <provider>.<uid>: %s", UID)
	}
	parts := strings.SplitN(UID, ".", 2)
	path := parts[0]
	id := parts[1]

	provider, err := config.Registry.GetProvider(path)
	if err != nil {
		return err
	}

	resource, err := provider.GetByUID(id)
	if err != nil {
		return err
	}

	resource.Detail = provider.Unprepare(resource.Detail)
	rep, err := resource.GetRepresentation()
	if err != nil {
		return err
	}

	fmt.Println(rep)
	return nil
}

// List outputs the keys of the grafanaDashboards object.
func List(config Config, jsonnetFile string) error {
	resources, err := parse(config, jsonnetFile)
	if err != nil {
		return err
	}

	f := "%s\t%s\n"
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)

	fmt.Fprintf(w, f, "KIND", "NAME")
	for _, r := range resources {
		fmt.Fprintf(w, f, r.Kind(), r.UID)
	}

	return w.Flush()
}

func getPrivateElementsScript(jsonnetFile string, providers []Provider) string {
	const script = `
    local src = import '%s';
    src + {
    %s
    }
	`
	providerStrings := []string{}
	for _, provider := range providers {
		jsonPath := provider.GetJSONPath()
		providerStrings = append(providerStrings, fmt.Sprintf("  %s+::: {},", jsonPath))
	}
	return fmt.Sprintf(script, jsonnetFile, strings.Join(providerStrings, "\n"))
}

func parse(config Config, jsonnetFile string) (Resources, error) {

	script := getPrivateElementsScript(jsonnetFile, config.Registry.ProviderList)
	vm := jsonnet.MakeVM()
	vm.Importer(newExtendedImporter([]string{"vendor", "lib", "."}))

	result, err := vm.EvaluateSnippet(jsonnetFile, script)
	if err != nil {
		return nil, err
	}

	msi := map[string]interface{}{}
	if err := json.Unmarshal([]byte(result), &msi); err != nil {
		return nil, err
	}

	r := Resources{}
	for k, v := range msi {
		provider, err := config.Registry.GetProvider(k)
		if err != nil {
			fmt.Println("Skipping unregistered path", k)
			continue
		}
		resources, err := provider.Parse(v)
		if err != nil {
			return nil, err
		}
		for kk, vv := range resources {
			r[kk] = vv
		}
	}
	return r, nil
}

// Show renders a Jsonnet dashboard as JSON, consuming a jsonnet filename
func Show(config Config, jsonnetFile string, targets []string) error {
	resources, err := parse(config, jsonnetFile)
	if err != nil {
		return err
	}

	var items []term.PageItem
	for _, resource := range resources {
		resource.Detail = resource.Provider.Unprepare(resource.Detail)

		rep, err := resource.GetRepresentation()
		if err != nil {
			return err
		}
		if interactive {
			items = append(items, term.PageItem{
				Name:    fmt.Sprintf("%s/%s", resource.Kind(), resource.UID),
				Content: rep,
			})
		} else {
			fmt.Printf("%s/%s:\n", resource.Kind(), resource.UID)
			fmt.Println(rep)
		}
	}
	if interactive {
		return term.Page(items)
	}
	return nil
}

// Diff renders Jsonnet resources and compares them to those at the endpoints
func Diff(config Config, jsonnetFile string, targets []string) error {
	resources, err := parse(config, jsonnetFile)
	if err != nil {
		return err
	}

	for _, resource := range resources {
		provider := resource.Provider
		local, err := resource.GetRepresentation()
		if err != nil {
			return nil
		}
		resource.Detail = provider.Unprepare(resource.Detail)
		uid := resource.UID
		remote, err := provider.GetRemote(resource.UID)
		if err == ErrNotFound {
			log.Printf("%s/%s %s\n", resource.Path, uid, Yellow("not present in "+resource.Kind()))
			continue
		}
		if err != nil {
			return fmt.Errorf("Error retrieving resource from %s %s: %v", resource.Kind(), uid, err)
		}
		remote.Detail = provider.Unprepare(remote.Detail)
		remoteRepresentation, err := (*remote).GetRepresentation()
		if err != nil {
			return err
		}

		if local == remoteRepresentation {
			fmt.Printf("%s/%s %s\n", resource.Path, uid, Yellow("no differences"))
		} else {
			fmt.Printf("%s/%s %s\n", resource.Path, uid, Red("changes detected:"))
			difference := diff.Diff(remoteRepresentation, local)
			fmt.Println(difference)
		}
	}
	return nil
}

// Apply renders Jsonnet dashboards then pushes them to Grafana via the API
func Apply(config Config, jsonnetFile string, targets []string) error {
	resources, err := parse(config, jsonnetFile)
	if err != nil {
		return err
	}

	for _, resource := range resources {
		if resource.MatchesTarget(targets) {
			provider := resource.Provider
			existingResource, err := provider.GetRemote(resource.UID)
			if err == ErrNotFound {

				err := provider.Add(resource.Detail)
				if err != nil {
					return err
				}
				fmt.Println(resource.UID, Green("added"))
				continue
			} else if err != nil {
				return err
			}
			resource.Detail = provider.Prepare(existingResource.Detail, resource.Detail)
			resourceRepresentation, err := resource.GetRepresentation()
			existingResourceRepresentation, err := existingResource.GetRepresentation()
			if resourceRepresentation == existingResourceRepresentation {
				fmt.Println(resource.UID, Yellow("unchanged"))
			} else {
				err = provider.Update(existingResource.Detail, resource.Detail)
				if err != nil {
					return err
				}
				log.Println(resource.UID, Green("updated"))
			}
		}
	}
	return nil
}

// Preview renders Jsonnet dashboards then pushes them to Grafana via the Snapshot API
func Preview(config Config, jsonnetFile string, targets []string, opts *PreviewOpts) error {
	resources, err := parse(config, jsonnetFile)
	if err != nil {
		return err
	}
	for _, resource := range resources {
		if resource.MatchesTarget(targets) {
			err := resource.Provider.Preview(resource.Detail, opts)
			if err == ErrNotImplemented {
				log.Println(resource.Provider.GetName()+" provider", Red("does not support preview"))
			}
			if err != nil {
				return err
			}
		}
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
					log.Println("Changes detected. Applying", jsonnetFile)
					resources, err := parse(config, jsonnetFile)
					if err != nil {
						log.Println("error:", err)
						continue
					}
					for _, resource := range resources {
						if resource.MatchesTarget(targets) {
							provider := resource.Provider
							existingResource, err := provider.GetRemote(resource.UID)
							if err == grizzly.ErrNotFound {
								err := resource.Provider.Add(resource.Detail)
								if err != nil {
									log.Println("Error:", err)
								}
							} else {
								err := resource.Provider.Update(existingResource.Detail, resource.Detail)
								if err != nil {
									log.Println("Error:", err)
								}
							}
						}
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
	resources, err := parse(config, jsonnetFile)
	if err != nil {
		return err
	}
	if _, err := os.Stat(dashboardDir); os.IsNotExist(err) {
		err = os.Mkdir(dashboardDir, 0755)
		if err != nil {
			return err
		}
	}
	for _, resource := range resources {
		if resource.MatchesTarget(targets) {
			updatedResource, err := resource.GetRepresentation()
			if err != nil {
				return err
			}
			uid := resource.UID
			extension := resource.Provider.GetExtension()
			dir := fmt.Sprintf("%s/%s", dashboardDir, resource.Kind())
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				err = os.Mkdir(dir, 0755)
				if err != nil {
					return err
				}
			}
			path := fmt.Sprintf("%s/%s.%s", dir, resource.UID, extension)

			existingResourceBytes, err := ioutil.ReadFile(path)
			isNotExist := os.IsNotExist(err)
			if err != nil && !isNotExist {
				return err
			}
			existingResource := string(existingResourceBytes)
			if existingResource == updatedResource {
				fmt.Println(uid, Yellow("unchanged"))
			} else {
				err = ioutil.WriteFile(path, []byte(updatedResource), 0644)
				if err != nil {
					return err
				}
				if isNotExist {
					fmt.Println(uid, Green("added"))
				} else {
					fmt.Println(uid, Green("updated"))
				}
			}
		}
	}
	return nil
}
