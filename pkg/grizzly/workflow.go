package grizzly

import (
	_ "embed"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/grafana/grizzly/pkg/term"
	"github.com/pmezard/go-difflib/difflib"
	"golang.org/x/crypto/ssh/terminal"
	"gopkg.in/fsnotify.v1"
)

var interactive = terminal.IsTerminal(int(os.Stdout.Fd()))

// Get retrieves a resource from a remote endpoint using its UID
func Get(registry Registry, UID string) error {
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

	handler, err := registry.GetHandler(handlerName)
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
func List(registry Registry, resources Resources) error {
	f := "%s\t%s\t%s\n"
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)

	fmt.Fprintf(w, f, "API VERSION", "KIND", "UID")
	for _, resource := range resources {
		handler, err := registry.GetHandler(resource.Kind())
		if err != nil {
			return err
		}
		fmt.Fprintf(w, f, handler.APIVersion(), handler.Kind(), resource.Name())
	}
	return w.Flush()
}

// ListRetmote outputs the keys of remote resources
func ListRemote(registry Registry, opts Opts) error {
	f := "%s\t%s\t%s\n"
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)

	fmt.Fprintf(w, f, "API VERSION", "KIND", "UID")
	for _, handler := range registry.Handlers {
		if !registry.HandlerMatchesTarget(handler, opts.Targets) {
			continue
		}
		IDs, err := handler.ListRemote()
		if err != nil {
			return err
		}
		for _, id := range IDs {
			if registry.ResourceMatchesTarget(handler, id, opts.Targets) {
				fmt.Fprintf(w, f, handler.APIVersion(), handler.Kind(), id)
			}
		}
	}
	return w.Flush()
}

// Pulls remote resources
func Pull(registry Registry, resourcePath string, opts Opts) error {

	if !(opts.Directory) {
		return fmt.Errorf("pull only works with -d option")
	}

	for _, handler := range registry.Handlers {
		if !registry.HandlerMatchesTarget(handler, opts.Targets) {
			registry.Notifier().Info(SimpleString(handler.Kind()), "skipped")
			continue
		}
		UIDs, err := handler.ListRemote()
		if err != nil {
			return err
		}
		if len(UIDs) == 0 {
			registry.Notifier().Info(nil, "No resources found")
		}
		registry.Notifier().Warn(nil, fmt.Sprintf("Pulling %d resources", len(UIDs)))
		for _, UID := range UIDs {
			if !registry.ResourceMatchesTarget(handler, UID, opts.Targets) {
				continue
			}
			resource, err := handler.GetByUID(UID)
			if errors.As(err, &ErrNotFound) {
				registry.Notifier().NotFound(SimpleString(UID))
				return nil
			}
			if err != nil {
				return err
			}

			path := filepath.Join(resourcePath, handler.ResourceFilePath(*resource, "yaml"))
			err = MarshalYAML(*resource, path)
			if err != nil {
				return err
			}
			registry.Notifier().Info(resource, "pulled")
		}
	}
	return nil
}

// Show displays resources
func Show(registry Registry, resources Resources) error {

	var items []term.PageItem
	for _, resource := range resources {
		handler, err := registry.GetHandler(resource.Kind())
		if err != nil {
			return nil
		}
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
			fmt.Printf("%s/%s:\n", resource.Kind(), resource.Name())
			fmt.Println(rep)
		}
	}
	if interactive {
		return term.Page(items)
	}
	return nil
}

// Diff compares resources to those at the endpoints
func Diff(registry Registry, resources Resources) error {

	for _, resource := range resources {
		handler, err := registry.GetHandler(resource.Kind())
		if err != nil {
			return nil
		}
		local, err := resource.YAML()
		if err != nil {
			return nil
		}
		resource = *handler.Unprepare(resource)
		uid := resource.Name()
		remote, err := handler.GetRemote(resource)
		if err == ErrNotFound {
			registry.Notifier().NotFound(resource)
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
			registry.Notifier().NoChanges(resource)
		} else {
			diff := difflib.UnifiedDiff{
				A:        difflib.SplitLines(remoteRepresentation),
				B:        difflib.SplitLines(local),
				FromFile: "Remote",
				ToFile:   "Local",
				Context:  3,
			}
			difference, _ := difflib.GetUnifiedDiffString(diff)
			registry.Notifier().HasChanges(resource, difference)
		}
	}
	return nil
}

// Apply pushes resources to endpoints
func Apply(registry Registry, resources Resources) error {
	for _, resource := range resources {
		handler, err := registry.GetHandler(resource.Kind())
		if err != nil {
			return nil
		}
		existingResource, err := handler.GetRemote(resource)
		if err == ErrNotFound {

			err := handler.Add(resource)
			if err != nil {
				return err
			}
			registry.Notifier().Added(resource)
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
			registry.Notifier().NoChanges(resource)
		} else {
			err = handler.Update(*existingResource, resource)
			if err != nil {
				return err
			}
			registry.Notifier().Updated(resource)
		}
	}
	return nil
}

// Preview pushes resources to endpoints as previews, if supported
func Preview(registry Registry, resources Resources, opts *PreviewOpts) error {
	for _, resource := range resources {
		handler, err := registry.GetHandler(resource.Kind())
		if err != nil {
			return nil
		}
		previewHandler, ok := handler.(PreviewHandler)
		if !ok {
			registry.Notifier().NotSupported(resource, "preview")
			return nil
		}
		err = previewHandler.Preview(resource, *registry.Notifier(), opts)
		if err != nil {
			return err
		}
	}
	return nil
}

// WatchParser encapsulates the action of parsing a resource (jsonnet or otherwise)
type WatchParser interface {
	Name() string
	Parse(registry Registry) (Resources, error)
}

// Watch watches a directory for changes then pushes Jsonnet resource to endpoints
// when changes are noticed
func Watch(registry Registry, watchDir string, parser WatchParser) error {
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
					resources, err := parser.Parse(registry)
					if err != nil {
						log.Println("Error: ", err)
					}
					err = Apply(registry, resources)
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
func Listen(registry Registry, UID, filename string) error {
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

	handler, err := registry.GetHandler(handlerName)
	if err != nil {
		return err
	}
	listenHandler, ok := handler.(ListenHandler)
	if !ok {
		uid := fmt.Sprintf("%s.%s", handler.Kind(), resourceID)
		registry.Notifier().NotSupported(SimpleString(uid), "listen")
		return nil
	}
	return listenHandler.Listen(*registry.Notifier(), resourceID, filename)
}

// Export renders Jsonnet resources then saves them to a directory
func Export(registry Registry, exportDir string, resources Resources) error {
	if _, err := os.Stat(exportDir); os.IsNotExist(err) {
		err = os.Mkdir(exportDir, 0755)
		if err != nil {
			return err
		}
	}

	for _, resource := range resources {
		handler, err := registry.GetHandler(resource.Kind())
		if err != nil {
			return nil
		}
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
			registry.Notifier().NoChanges(resource)
		} else {
			err = ioutil.WriteFile(path, []byte(updatedResource), 0644)
			if err != nil {
				return err
			}
			if isNotExist {
				registry.Notifier().Added(resource)
			} else {
				registry.Notifier().Updated(resource)
			}
		}
	}
	return nil
}
