package grizzly

import (
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/grafana/grizzly/pkg/grizzly/notifier"
	"github.com/grafana/grizzly/pkg/term"
	"github.com/pmezard/go-difflib/difflib"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
	"gopkg.in/fsnotify.v1"
)

var interactive = terminal.IsTerminal(int(os.Stdout.Fd()))

// Get retrieves a resource from a remote endpoint using its UID
func Get(UID string) error {
	log.Info("Getting ", UID)

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

	handler, err := Registry.GetHandler(handlerName)
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
func List(resources Resources) error {
	log.Infof("Listing %d resources", resources.Len())

	f := "%s\t%s\t%s\n"
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)

	fmt.Fprintf(w, f, "API VERSION", "KIND", "UID")
	for _, resource := range resources {
		handler, err := Registry.GetHandler(resource.Kind())
		if err != nil {
			return err
		}
		fmt.Fprintf(w, f, handler.APIVersion(), handler.Kind(), resource.Name())
	}
	return w.Flush()
}

// ListRetmote outputs the keys of remote resources
func ListRemote(opts Opts) error {
	log.Info("Listing remotes")

	f := "%s\t%s\t%s\n"
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)

	fmt.Fprintf(w, f, "API VERSION", "KIND", "UID")
	for name, handler := range Registry.Handlers {
		if !Registry.HandlerMatchesTarget(handler, opts.Targets) {
			continue
		}
		log.Debugf("Listing remote values for handler %s", name)
		IDs, err := handler.ListRemote()
		if err != nil {
			return err
		}
		for _, id := range IDs {
			if Registry.ResourceMatchesTarget(handler, id, opts.Targets) {
				fmt.Fprintf(w, f, handler.APIVersion(), handler.Kind(), id)
			}
		}
	}
	return w.Flush()
}

// Pulls remote resources
func Pull(resourcePath string, opts Opts) error {
	log.Infof("Pulling resources from %s", resourcePath)

	if !(opts.Directory) {
		return fmt.Errorf("pull only works with -d option")
	}

	for name, handler := range Registry.Handlers {
		if !Registry.HandlerMatchesTarget(handler, opts.Targets) {
			notifier.Info(notifier.SimpleString(handler.Kind()), "skipped")
			continue
		}
		log.Debugf("Listing remote values for handler %s", name)
		UIDs, err := handler.ListRemote()
		if err != nil {
			return err
		}
		if len(UIDs) == 0 {
			notifier.Info(nil, "No resources found")
		}
		notifier.Warn(nil, fmt.Sprintf("Pulling %d resources", len(UIDs)))
		for _, UID := range UIDs {
			if !Registry.ResourceMatchesTarget(handler, UID, opts.Targets) {
				continue
			}
			resource, err := handler.GetByUID(UID)
			if errors.Is(err, ErrNotFound) {
				notifier.NotFound(notifier.SimpleString(UID))
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
			notifier.Info(resource, "pulled")
		}
	}
	return nil
}

// Show displays resources
func Show(resources Resources) error {
	log.Infof("Showing %d resources", resources.Len())

	var items []term.PageItem
	for _, resource := range resources {
		handler, err := Registry.GetHandler(resource.Kind())
		if err != nil {
			return err
		}
		resource = *(handler.Unprepare(resource))

		rep, err := resource.YAML()
		if err != nil {
			return err
		}
		if interactive {
			items = append(items, term.PageItem{
				Name:    fmt.Sprintf("%s.%s", resource.Kind(), resource.Name()),
				Content: rep,
			})
		} else {
			fmt.Printf("%s.%s:\n", resource.Kind(), resource.Name())
			fmt.Println(rep)
		}
	}
	if interactive {
		return term.Page(items)
	}
	return nil
}

// Diff compares resources to those at the endpoints
func Diff(resources Resources) error {
	log.Infof("Diff-ing %d resources", resources.Len())

	for _, resource := range resources {
		handler, err := Registry.GetHandler(resource.Kind())
		if err != nil {
			return err
		}

		local, err := resource.YAML()
		if err != nil {
			return err
		}

		resource = *handler.Unprepare(resource)
		uid := resource.Name()

		log.Debugf("Getting the remote value for `%s`", resource.Key())
		remote, err := handler.GetRemote(resource)
		if errors.Is(err, ErrNotFound) {
			notifier.NotFound(resource)
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
			notifier.NoChanges(resource)
		} else {
			diff := difflib.UnifiedDiff{
				A:        difflib.SplitLines(remoteRepresentation),
				B:        difflib.SplitLines(local),
				FromFile: "Remote",
				ToFile:   "Local",
				Context:  3,
			}
			difference, _ := difflib.GetUnifiedDiffString(diff)
			notifier.HasChanges(resource, difference)
		}
	}

	return nil
}

// Apply pushes resources to endpoints
func Apply(resources Resources) error {
	log.Infof("Applying %d resources", resources.Len())

	for _, resource := range resources {
		handler, err := Registry.GetHandler(resource.Kind())
		if err != nil {
			return err
		}

		log.Debugf("Getting the remote value for `%s`", resource.Key())
		existingResource, err := handler.GetRemote(resource)
		if errors.Is(err, ErrNotFound) {
			log.Debugf("`%s` was not found, adding it...", resource.Key())

			err := handler.Validate(resource)
			if err != nil {
				return fmt.Errorf("resource %s is not valid: %v", resource.Key(), err)
			}

			if err := handler.Add(resource); err != nil {
				return err
			}

			notifier.Added(resource)
			continue
		}
		if err != nil {
			return err
		}
		log.Debugf("`%s` was found, updating it...", resource.Key())

		resourceRepresentation, err := resource.YAML()
		if err != nil {
			return err
		}

		resource = *handler.Prepare(*existingResource, resource)
		existingResource = handler.Unprepare(*existingResource)
		existingResourceRepresentation, err := existingResource.YAML()
		if err != nil {
			return err
		}

		if resourceRepresentation == existingResourceRepresentation {
			notifier.NoChanges(resource)
			continue
		}

		if err = handler.Update(*existingResource, resource); err != nil {
			return err
		}

		notifier.Updated(resource)
	}

	return nil
}

// Preview pushes resources to endpoints as previews, if supported
func Preview(resources Resources, opts *PreviewOpts) error {
	for _, resource := range resources {
		handler, err := Registry.GetHandler(resource.Kind())
		if err != nil {
			return err
		}
		previewHandler, ok := handler.(PreviewHandler)
		if !ok {
			notifier.NotSupported(resource, "preview")
			continue
		}
		err = previewHandler.Preview(resource, opts)
		if err != nil {
			return err
		}
	}
	return nil
}

// WatchParser encapsulates the action of parsing a resource (jsonnet or otherwise)
type WatchParser interface {
	Name() string
	Parse() (Resources, error)
}

// Watch watches a directory for changes then pushes Jsonnet resource to endpoints
// when changes are noticed.
func Watch(watchDir string, parser WatchParser) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		log.Info("Watching for changes")
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Info("Changes detected. Applying ", parser.Name())
					resources, err := parser.Parse()
					if err != nil {
						log.Error("Error: ", err)
					}
					err = Apply(resources)
					if err != nil {
						log.Error("Error: ", err)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Error("error: ", err)
			}
		}
	}()

	if err := filepath.WalkDir(watchDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return watcher.Add(path)
		}

		return nil
	}); err != nil {
		return err
	}

	<-done

	return nil
}

// Listen waits for remote changes to a resource and saves them to disk
func Listen(UID, filename string) error {
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

	handler, err := Registry.GetHandler(handlerName)
	if err != nil {
		return err
	}
	listenHandler, ok := handler.(ListenHandler)
	if !ok {
		uid := fmt.Sprintf("%s.%s", handler.Kind(), resourceID)
		notifier.NotSupported(notifier.SimpleString(uid), "listen")
		return nil
	}
	return listenHandler.Listen(resourceID, filename)
}

// Export renders Jsonnet resources then saves them to a directory
func Export(exportDir string, resources Resources) error {
	if _, err := os.Stat(exportDir); os.IsNotExist(err) {
		err = os.Mkdir(exportDir, 0755)
		if err != nil {
			return err
		}
	}

	for _, resource := range resources {
		handler, err := Registry.GetHandler(resource.Kind())
		if err != nil {
			return err
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

		existingResourceBytes, err := os.ReadFile(path)
		isNotExist := os.IsNotExist(err)
		if err != nil && !isNotExist {
			return err
		}
		existingResource := string(existingResourceBytes)
		if existingResource == updatedResource {
			notifier.NoChanges(resource)
		} else {
			err = os.WriteFile(path, []byte(updatedResource), 0644)
			if err != nil {
				return err
			}
			if isNotExist {
				notifier.Added(resource)
			} else {
				notifier.Updated(resource)
			}
		}
	}
	return nil
}
