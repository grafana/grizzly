package grizzly

import (
	"fmt"
	"strings"

	"github.com/gobwas/glob"
)

// Provider describes a single Endpoint Provider
type Provider interface {
	Group() string
	Version() string
	APIVersion() string
	GetHandlers() []Handler
}

// Registry records providers
type Registry struct {
	Providers []Provider
	Handlers  map[string]Handler
}

// Global Handler registry
var registry Registry

// NewProviderRegistry returns a new registry instance
func NewProviderRegistry() Registry {
	registry := Registry{}
	registry.Providers = []Provider{}
	registry.Handlers = map[string]Handler{}
	return registry
}

// RegisterProvider will register a new provider
func (r *Registry) RegisterProvider(provider Provider) error {
	r.Providers = append(r.Providers, provider)
	for _, handler := range provider.GetHandlers() {
		r.Handlers[handler.Kind()] = handler
	}
	return nil
}

// GetHandler returns a single provider based upon a JSON path
func (r *Registry) GetHandler(path string) (Handler, error) {
	handler, exists := r.Handlers[path]
	if !exists {
		return nil, fmt.Errorf("couldn't find a handler for %s: %w", path, ErrHandlerNotFound)
	}
	return handler, nil
}

// HandlerMatchesTarget identifies whether a handler is in a target list
func (r *Registry) HandlerMatchesTarget(handler Handler, targets []string) bool {
	if len(targets) == 0 {
		return true
	}
	key := handler.Kind()

	for _, target := range targets {
		if strings.Contains(target, "/") && strings.Split(target, "/")[0] == key {
			return true
		}
	}
	return false
}

// ResourceMatchesTarget identifies whether a resource is in a target list
func (r *Registry) ResourceMatchesTarget(handler Handler, UID string, targets []string) bool {
	if len(targets) == 0 {
		return true
	}
	key := fmt.Sprintf("%s/%s", handler.Kind(), UID)
	for _, target := range targets {
		g := glob.MustCompile(target)
		if g.Match(key) {
			return true
		}
	}
	return false
}

// Notifier returns a notifier for responding to users
func (r *Registry) Notifier() *Notifier {
	return &Notifier{}
}
