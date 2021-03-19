package grizzly

import (
	"bufio"
	_ "embed"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/google/go-jsonnet"
	"github.com/grafana/grizzly/pkg/term"
	"github.com/grafana/tanka/pkg/jsonnet/native"
	"github.com/grafana/tanka/pkg/kubernetes/manifest"
	"github.com/grafana/tanka/pkg/process"
	"github.com/pmezard/go-difflib/difflib"
	"golang.org/x/crypto/ssh/terminal"
	"gopkg.in/fsnotify.v1"
	"gopkg.in/yaml.v3"
)

var interactive = terminal.IsTerminal(int(os.Stdout.Fd()))

// Get retrieves a resource from a remote endpoint using its UID
func Get(config Config, UID string) error {
	count := strings.Count(UID, ".")
	var handlerName, resourceID string
	if count == 1 {
		parts := strings.SplitN(UID, ".", 2)
		handlerName = parts[0]
		resourceID = parts[1]
	} else if count == 2 {
		parts := strings.SplitN(UID, ".", 3)
		handlerName = parts[0] + "." + parts[1]
		resourceID = parts[2]

	} else {
		return fmt.Errorf("UID must be <provider>.<uid>: %s", UID)
	}

	handler, err := config.Registry.GetHandler(handlerName)
	if err != nil {
		return err
	}

	resource, err := handler.GetByUID(resourceID)
	if err != nil {
		return err
	}

	resource = handler.Unprepare(*resource)
	rep, err := resource.YAML()
	if err != nil {
		return err
	}

	fmt.Println(rep)
	return nil
}

// List outputs the keys resources found in resulting json.
func List(config Config, resources Resources) error {
	f := "%s\t%s\t%s\n"
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)

	fmt.Fprintf(w, f, "API VERSION", "KIND", "UID")
	for handler, resourceList := range resources {
		for _, r := range resourceList {
			fmt.Fprintf(w, f, handler.APIVersion(), handler.Kind(), r.Name())
		}
	}
	return w.Flush()
}

//go:embed grizzly.jsonnet
var script string

// Parse parses a file into a list of resources
func Parse(config Config, file string, targets []string) (Resources, error) {
	if strings.HasSuffix(file, ".yaml") || strings.HasSuffix(file, ".") {
		return ParseYAML(config, file, targets)
	}
	return ParseJsonnet(config, file, targets)
}

// ParseYAML evaluates a YAML file and parses it into resources
func ParseYAML(config Config, yamlFile string, targets []string) (Resources, error) {
	f, err := os.Open(yamlFile)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(f)
	decoder := yaml.NewDecoder(reader)
	manifests := map[string]manifest.Manifest{}
	var m manifest.Manifest
	for i := 0; decoder.Decode(&m) == nil; i++ {
		manifests[strconv.Itoa(i)] = m
	}
	return ManifestsAsResources(config, yamlFile, manifests, targets)
}

// ParseJsonnet evaluates a jsonnet file and parses it into resources
func ParseJsonnet(config Config, jsonnetFile string, targets []string) (Resources, error) {
	script := fmt.Sprintf(script, jsonnetFile)
	vm := jsonnet.MakeVM()
	vm.Importer(newExtendedImporter([]string{"vendor", "lib", "."}))
	for _, nf := range native.Funcs() {
		vm.NativeFunction(nf)
	}

	result, err := vm.EvaluateSnippet(jsonnetFile, script)
	if err != nil {
		return nil, err
	}
	var data interface{}
	if err := json.Unmarshal([]byte(result), &data); err != nil {
		return nil, err
	}

	extracted, err := process.Extract(data)
	if err != nil {
		return nil, err
	}

	// Unwrap *List types
	if err := process.Unwrap(extracted); err != nil {
		return nil, err
	}
	return ManifestsAsResources(config, jsonnetFile, extracted, targets)
}

func ParseDirectory(config Config, source string, m manifest.Manifest, targets []string) (Resources, error) {
	resource := Resource(m)
	var files []string
	if resource.HasSpecKey("glob") {
		glob := resource.GetSpecString("glob")
		globPath := filepath.Join(source, glob)
		globbedFiles, err := filepath.Glob(globPath)
		if err != nil {
			return nil, err
		}
		for _, file := range globbedFiles {
			files = append(files, file)
		}
	} else {
		path := resource.GetSpecString("path")
		fullpath := filepath.Join(source, path)
		fis, err := ioutil.ReadDir(fullpath)
		if err != nil {
			return nil, err
		}
		for _, fi := range fis {
			files = append(files, filepath.Join(fullpath, fi.Name()))
		}
		log.Println("NONGLOB", files)
	}
	resources := Resources{}
	for _, file := range files {
		newResources, err := Parse(config, file, targets)
		if err != nil {
			return nil, err
		}
		resources.AddResources(newResources)
	}
	return resources, nil
}

func ManifestsAsResources(config Config, file string, manifests map[string]manifest.Manifest, targets []string) (Resources, error) {
	resources := Resources{}
	for _, m := range manifests {
		if m.Kind() == "Directory" {
			source := filepath.Dir(file)
			newResources, err := ParseDirectory(config, source, m, targets)
			if err != nil {
				return nil, err
			}
			resources.AddResources(newResources)
		} else {
			handler, err := config.Registry.GetHandler(m.Kind())
			if err != nil {
				log.Println("Error getting handler", err)
				continue
			}
			parsedResources, err := handler.Parse(m)
			if err != nil {
				return nil, err
			}
			for _, resource := range parsedResources {
				handler, err = config.Registry.GetHandler(resource.Kind())
				if !resource.MatchesTarget(targets) {
					continue
				}
				resources.AddResource(handler, resource)
			}
		}
	}
	return resources, nil
}

// Show displays resources
func Show(config Config, resources Resources) error {

	var items []term.PageItem
	for handler, resourceList := range resources {
		for _, resource := range resourceList {
			resource = *(handler.Unprepare(resource))

			rep, err := resource.YAML()
			if err != nil {
				return err
			}
			if interactive {
				items = append(items, term.PageItem{
					Name:    fmt.Sprintf("%s/%s", resource.Kind(), resource.Name()),
					Content: rep,
				})
			} else {
				fmt.Println(rep)
				fmt.Println("---")
			}
		}
	}
	if interactive {
		return term.Page(items)
	}
	return nil
}

// Diff compares resources to those at the endpoints
func Diff(config Config, resources Resources) error {

	for handler, resourceList := range resources {
		for _, resource := range resourceList {
			local, err := resource.YAML()
			if err != nil {
				return nil
			}
			resource = *handler.Unprepare(resource)
			uid := resource.Name()
			remote, err := handler.GetRemote(resource)
			if err == ErrNotFound {
				config.Notifier.NotFound(resource)
				continue
			}
			if err != nil {
				return fmt.Errorf("Error retrieving resource from %s %s: %v", resource.Kind(), uid, err)
			}
			remote = handler.Unprepare(*remote)
			remoteRepresentation, err := (*remote).YAML()
			if err != nil {
				return err
			}

			if local == remoteRepresentation {
				config.Notifier.NoChanges(resource)
			} else {
				diff := difflib.UnifiedDiff{
					A:        difflib.SplitLines(remoteRepresentation),
					B:        difflib.SplitLines(local),
					FromFile: "Remote",
					ToFile:   "Local",
					Context:  3,
				}
				difference, _ := difflib.GetUnifiedDiffString(diff)
				config.Notifier.HasChanges(resource, difference)
			}
		}
	}
	return nil
}

// Apply pushes resources to endpoints
func Apply(config Config, resources Resources) error {
	for handler, resourceList := range resources {
		for _, resource := range resourceList {
			existingResource, err := handler.GetRemote(resource)
			if err == ErrNotFound {

				err := handler.Add(resource)
				if err != nil {
					return err
				}
				config.Notifier.Added(resource)
				continue
			} else if err != nil {
				return err
			}
			resourceRepresentation, err := resource.YAML()
			if err != nil {
				return err
			}
			resource = *handler.Prepare(*existingResource, resource)
			existingResource = handler.Unprepare(*existingResource)
			existingResourceRepresentation, err := existingResource.YAML()
			if err != nil {
				return nil
			}
			if resourceRepresentation == existingResourceRepresentation {
				config.Notifier.NoChanges(resource)
			} else {
				err = handler.Update(*existingResource, resource)
				if err != nil {
					return err
				}
				config.Notifier.Updated(resource)
			}
		}
	}
	return nil
}

// Preview pushes resources to endpoints as previews, if supported
func Preview(config Config, resources Resources, opts *PreviewOpts) error {
	for handler, resourceList := range resources {
		for _, resource := range resourceList {
			previewHandler, ok := handler.(PreviewHandler)
			if !ok {
				config.Notifier.NotSupported(handler.Kind(), resource.Name(), "preview")
				return nil
			}
			err := previewHandler.Preview(resource, config.Notifier, opts)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Parser encapsulates the action of parsing a resource (jsonnet or otherwise)
type Parser interface {
	Name() string
	Parse(config Config) (Resources, error)
}

// Watch watches a directory for changes then pushes Jsonnet resource to endpoints
// when changes are noticed
func Watch(config Config, watchDir string, parser Parser) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		log.Println("Watching for changes")
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("Changes detected. Applying", parser.Name())
					resources, err := parser.Parse(config)
					if err != nil {
						log.Println("Error: ", err)
					}
					err = Apply(config, resources)
					if err != nil {
						log.Println("Error: ", err)
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

// Listen waits for remote changes to a resource and saves them to disk
func Listen(config Config, UID, filename string) error {
	count := strings.Count(UID, ".")
	var handlerName, resourceID string
	if count == 1 {
		parts := strings.SplitN(UID, ".", 2)
		handlerName = parts[0]
		resourceID = parts[1]
	} else if count == 2 {
		parts := strings.SplitN(UID, ".", 3)
		handlerName = parts[0] + "." + parts[1]
		resourceID = parts[2]

	} else {
		return fmt.Errorf("UID must be <provider>.<uid>: %s", UID)
	}

	handler, err := config.Registry.GetHandler(handlerName)
	if err != nil {
		return err
	}
	listenHandler, ok := handler.(ListenHandler)
	if !ok {
		config.Notifier.NotSupported(handler.Kind(), resourceID, "listen")
		return nil
	}
	return listenHandler.Listen(config.Notifier, resourceID, filename)
}

// Export renders Jsonnet resources then saves them to a directory
func Export(config Config, exportDir string, resources Resources) error {
	if _, err := os.Stat(exportDir); os.IsNotExist(err) {
		err = os.Mkdir(exportDir, 0755)
		if err != nil {
			return err
		}
	}

	for handler, resourceList := range resources {
		for _, resource := range resourceList {
			updatedResource, err := resource.YAML()
			if err != nil {
				return err
			}
			extension := handler.GetExtension()
			dir := fmt.Sprintf("%s/%s", exportDir, resource.Kind())
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				err = os.Mkdir(dir, 0755)
				if err != nil {
					return err
				}
			}
			path := fmt.Sprintf("%s/%s.%s", dir, resource.Name(), extension)

			existingResourceBytes, err := ioutil.ReadFile(path)
			isNotExist := os.IsNotExist(err)
			if err != nil && !isNotExist {
				return err
			}
			existingResource := string(existingResourceBytes)
			if existingResource == updatedResource {
				config.Notifier.NoChanges(resource)
			} else {
				err = ioutil.WriteFile(path, []byte(updatedResource), 0644)
				if err != nil {
					return err
				}
				if isNotExist {
					config.Notifier.Added(resource)
				} else {
					config.Notifier.Updated(resource)
				}
			}
		}
	}
	return nil
}
