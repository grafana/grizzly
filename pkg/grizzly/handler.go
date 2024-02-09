package grizzly

import (
	"github.com/grafana/tanka/pkg/kubernetes/manifest"
)

// Handler describes a handler for a single API resource handled by a single provider
type Handler interface {
	APIVersion() string
	Kind() string

	// ResourceFilePath returns the location on disk where a resource should be updated
	ResourceFilePath(resource Resource, filetype string) string

	// Parse parses a manifest object into a struct for this resource type
	Parse(m manifest.Manifest) (Resources, error)

	// Unprepare removes unnecessary elements from a remote resource ready for presentation/comparison
	Unprepare(resource Resource) *Resource

	// Prepare gets a resource ready for dispatch to the remote endpoint
	Prepare(existing, resource Resource) *Resource

	// Retrieves a UID for a resource
	GetUID(resource Resource) (string, error)

	// Get retrieves JSON for a resource from an endpoint, by UID
	GetByUID(UID string) (*Resource, error)

	// GetRemote retrieves a remote equivalent of a remote resource
	GetRemote(resource Resource) (*Resource, error)

	// ListRemote retrieves as list of UIDs of all remote resources
	ListRemote() ([]string, error)

	// Add pushes a new resource to the endpoint
	Add(resource Resource) error

	// Update pushes an existing resource to the endpoint
	Update(existing, resource Resource) error

	// Validate gets or build the uid of corresponding resource
	Validate(resource Resource) error

	// Sort sorts resources as defined by the handler
	Sort(resources Resources) Resources

	// UsesFolders identifies whether this resource lives within a folder
	UsesFolders() bool
}

// PreviewHandler describes a handler that has the ability to render
// a preview of a resource
type PreviewHandler interface {
	// Preview renders Jsonnet then pushes them to the endpoint if previews are possible
	Preview(resource Resource, opts *PreviewOpts) error
}

// ListenHandler describes a handler that has the ability to watch a single
// resource for changes, and write changes to that resource to a local file
type ListenHandler interface {
	// Listen watches a resource and update local file on changes
	Listen(UID, filename string) error
}
