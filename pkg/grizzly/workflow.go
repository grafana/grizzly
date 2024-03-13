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
	"github.com/hashicorp/go-multierror"
	"github.com/pmezard/go-difflib/difflib"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
	"gopkg.in/fsnotify.v1"
)

var interactive = terminal.IsTerminal(int(os.Stdout.Fd()))

// Get retrieves a resource from a remote endpoint using its UID
func Get(registry Registry, UID string, onlySpec bool, outputFormat string) error {
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

	handler, err := registry.GetHandler(handlerName)
	if err != nil {
		return err
	}

	resource, err := handler.GetByUID(resourceID)
	if err != nil {
		return err
	}

	resource = handler.Unprepare(*resource)

	content, _, _, err := Format(registry, "", resource, outputFormat, onlySpec)
	if err != nil {
		return err
	}

	fmt.Println(string(content))
	return nil
}

// List outputs the keys resources found in resulting json.
func List(registry Registry, resources Resources) error {
	log.Infof("Listing %d resources", resources.Len())

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
func ListRemote(registry Registry, targets []string) error {
	log.Info("Listing remotes")

	f := "%s\t%s\t%s\n"
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)

	fmt.Fprintf(w, f, "API VERSION", "KIND", "UID")
	for name, handler := range registry.Handlers {
		if !registry.HandlerMatchesTarget(handler, targets) {
			continue
		}
		log.Debugf("Listing remote values for handler %s", name)
		IDs, err := handler.ListRemote()
		if err != nil {
			return err
		}
		for _, id := range IDs {
			if registry.ResourceMatchesTarget(handler.Kind(), id, targets) {
				fmt.Fprintf(w, f, handler.APIVersion(), handler.Kind(), id)
			}
		}
	}
	return w.Flush()
}

// Pull pulls remote resources and stores them in the local file system.
// The given resourcePath must be a directory, where all resources will be stored.
// If opts.JSONSpec is true, which is only applicable for dashboards, saves the spec as a JSON file.
func Pull(registry Registry, resourcePath string, onlySpec bool, outputFormat string, targets []string, continueOnError bool) error {
	resourcePathIsFile, err := isFile(resourcePath)
	if err != nil {
		return err
	}

	if resourcePathIsFile {
		return fmt.Errorf("pull <resource-path> must be a directory")
	}

	var finalErr error

	log.Infof("Pulling resources to %s", resourcePath)
	for name, handler := range registry.Handlers {
		if !registry.HandlerMatchesTarget(handler, targets) {
			notifier.Info(notifier.SimpleString(handler.Kind()), "skipped")
			continue
		}

		log.Debugf("Listing remote values for handler %s", name)
		UIDs, err := handler.ListRemote()
		if err != nil {
			finalErr = multierror.Append(finalErr, err)

			if continueOnError {
				notifier.Error(notifier.SimpleString(name), fmt.Sprintf("failed listing remote values: %s", err))
				continue
			}

			return finalErr
		}
		if len(UIDs) == 0 {
			notifier.Info(nil, "No resources found")
			continue
		}

		notifier.Warn(nil, fmt.Sprintf("Pulling %d resources", len(UIDs)))
		for _, UID := range UIDs {
			if !registry.ResourceMatchesTarget(handler.Kind(), UID, targets) {
				continue
			}

			resource, err := handler.GetByUID(UID)
			if errors.Is(err, ErrNotFound) {
				finalErr = multierror.Append(finalErr, err)
				notifier.NotFound(notifier.SimpleString(UID))
				if continueOnError {
					continue
				}

				return nil
			}
			if err != nil {
				finalErr = multierror.Append(finalErr, err)

				if continueOnError {
					notifier.Error(notifier.SimpleString(UID), fmt.Sprintf("failed pulling resource: %s", err))
					continue
				}

				return finalErr
			}

			resource = handler.Unprepare(*resource)

			content, filename, _, err := Format(registry, resourcePath, resource, outputFormat, onlySpec)
			if err != nil {
				finalErr = multierror.Append(finalErr, err)

				if continueOnError {
					notifier.Error(resource, fmt.Sprintf("failed formatting resource: %s", err))
					continue
				}

				return finalErr
			}

			err = WriteFile(filename, content)
			if err != nil {
				finalErr = multierror.Append(finalErr, err)

				if continueOnError {
					notifier.Error(resource, fmt.Sprintf("failed writing resource to file: %s", err))
					continue
				}

				return finalErr
			}

			notifier.Info(resource, "pulled")
		}
	}

	return finalErr
}

// Show displays resources
func Show(registry Registry, resources Resources, outputFormat string) error {
	log.Infof("Showing %d resources", resources.Len())

	var items []term.PageItem
	for _, resource := range resources {
		handler, err := registry.GetHandler(resource.Kind())
		if err != nil {
			return err
		}
		resource = *(handler.Unprepare(resource))

		content, _, _, err := Format(registry, "", &resource, outputFormat, false) // we always show full resource, even if only-spec was specified
		if err != nil {
			return err
		}

		if interactive {
			items = append(items, term.PageItem{
				Name:    resource.Ref().String(),
				Content: string(content),
			})
		} else {
			fmt.Printf("%s:\n", resource.Ref().String())
			fmt.Println(string(content))
		}
	}
	if interactive {
		return term.Page(items)
	}
	return nil
}

// Diff compares resources to those at the endpoints
func Diff(registry Registry, resources Resources, onlySpec bool, outputFormat string) error {
	log.Infof("Diff-ing %d resources", resources.Len())

	for _, resource := range resources {
		handler, err := registry.GetHandler(resource.Kind())
		if err != nil {
			return err
		}

		local, _, _, err := Format(registry, "", &resource, outputFormat, onlySpec)
		if err != nil {
			return err
		}

		resource = *handler.Unprepare(resource)
		uid := resource.Name()

		log.Debugf("Getting the remote value for `%s`", resource.Ref())
		remote, err := handler.GetRemote(resource)
		if errors.Is(err, ErrNotFound) {
			notifier.NotFound(resource)
			continue
		}

		if err != nil {
			return fmt.Errorf("Error retrieving resource from %s %s: %v", resource.Kind(), uid, err)
		}

		remote = handler.Unprepare(*remote)

		remoteRepresentation, _, _, err := Format(registry, "", remote, outputFormat, onlySpec)
		if err != nil {
			return err
		}

		if string(local) == string(remoteRepresentation) {
			notifier.NoChanges(resource)
		} else {
			diff := difflib.UnifiedDiff{
				A:        difflib.SplitLines(string(remoteRepresentation)),
				B:        difflib.SplitLines(string(local)),
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

type eventsRecorder interface {
	Record(event Event)
}

// Apply pushes resources to endpoints
func Apply(registry Registry, resources Resources, continueOnError bool, eventsRecorder eventsRecorder) error {
	var finalErr error

	for _, resource := range resources {
		err := applyResource(registry, resource, eventsRecorder)
		if err != nil {
			finalErr = multierror.Append(finalErr, err)

			eventsRecorder.Record(Event{
				Type:        ResourceFailure,
				ResourceRef: resource.Ref().String(),
				Details:     err.Error(),
			})

			if !continueOnError {
				return finalErr
			}
		}
	}

	return finalErr
}

func applyResource(registry Registry, resource Resource, trailRecorder eventsRecorder) error {
	resourceRef := resource.Ref().String()

	handler, err := registry.GetHandler(resource.Kind())
	if err != nil {
		return err
	}

	log.Debugf("Getting the remote value for `%s`", resource.Ref())
	existingResource, err := handler.GetRemote(resource)
	if errors.Is(err, ErrNotFound) {
		log.Debugf("`%s` was not found, adding it...", resource.Ref())

		if err := handler.Add(resource); err != nil {
			return err
		}

		trailRecorder.Record(Event{
			Type:        ResourceAdded,
			ResourceRef: resourceRef,
		})
		return nil
	}
	if err != nil {
		return err
	}

	log.Debugf("`%s` was found, updating it...", resource.Ref())

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
		trailRecorder.Record(Event{
			Type:        ResourceNotChanged,
			ResourceRef: resourceRef,
		})
		return nil
	}

	if err = handler.Update(*existingResource, resource); err != nil {
		return err
	}

	trailRecorder.Record(Event{
		Type:        ResourceUpdated,
		ResourceRef: resourceRef,
	})

	return nil
}

// WatchParser encapsulates the action of parsing a resource (jsonnet or otherwise)
type WatchParser interface {
	Name() string
	Parse() (Resources, error)
}

// Watch watches a directory for changes then pushes Jsonnet resource to endpoints
// when changes are noticed.
func Watch(registry Registry, watchDir string, parser WatchParser, trailRecorder eventsRecorder) error {
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
					err = Apply(registry, resources, false, trailRecorder) // TODO?
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

// Serve starts an HTTP server that can be used to navigate Grizzly resources,
// as well as allowing visualisation of resources handed to Grizzly.
// If pure files, they can be saved too.
func Serve(registry Registry, parser WatchParser, resourcePath string, port int, openBrowser, onlySpec bool, outputFormat string) error {
	server, err := NewGrizzlyServer(registry, parser, resourcePath, port, openBrowser, onlySpec, outputFormat)
	if err != nil {
		return err
	}
	return server.Start()
}

// Export renders Jsonnet resources then saves them to a directory
func Export(registry Registry, exportDir string, resources Resources, onlySpec bool, outputFormat string) error {
	if _, err := os.Stat(exportDir); os.IsNotExist(err) {
		err = os.Mkdir(exportDir, 0755)
		if err != nil {
			return err
		}
	}

	for _, resource := range resources {
		updatedResourceBytes, _, extension, err := Format(registry, "", &resource, outputFormat, onlySpec)
		if err != nil {
			return err
		}

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
		updatedResource := string(updatedResourceBytes)
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

func isFile(resourcePath string) (bool, error) {
	stat, err := os.Stat(resourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return !stat.IsDir(), nil
}
