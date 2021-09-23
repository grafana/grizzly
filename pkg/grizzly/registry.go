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

// ProviderSet records providers
type registry struct {
	Providers    []Provider
	Handlers     map[string]Handler
	HandlerOrder map[string]int
}

// Global Handler registry
var Registry registry

// NewProviderRegistry returns a new registry instance
func ConfigureProviderRegistry(providers []Provider) {
	Registry.Providers = providers
	Registry.Handlers = map[string]Handler{}
	Registry.HandlerOrder = map[string]int{}
	position := 0
	for _, provider := range providers {
		for _, handler := range provider.GetHandlers() {
			Registry.Handlers[handler.Kind()] = handler
			Registry.HandlerOrder[handler.Kind()] = position
			position++
		}
	}
}

// GetHandler returns a single provider based upon a JSON path
func (r *registry) GetHandler(path string) (Handler, error) {
	handler, exists := r.Handlers[path]
	if !exists {
		return nil, fmt.Errorf("couldn't find a handler for %s: %w", path, ErrHandlerNotFound)
	}
	return handler, nil
}

// HandlerMatchesTarget identifies whether a handler is in a target list
func (r *registry) HandlerMatchesTarget(handler Handler, targets []string) bool {
	if len(targets) == 0 {
		return true
	}
	key := handler.Kind()

	for _, target := range targets {
		if (strings.Contains(target, "/") && strings.Split(target, "/")[0] == key) ||
			(strings.Contains(target, ".") && strings.Split(target, ".")[0] == key) {
			return true
		}
	}
	return false
}

// ResourceMatchesTarget identifies whether a resource is in a target list
func (r *registry) ResourceMatchesTarget(handler Handler, UID string, targets []string) bool {
	if len(targets) == 0 {
		return true
	}
	// I mistakenly assumed 'dot' was a special character for globs, so opted for '/' as separator.
	// This keeps back-compat
	slashKey := fmt.Sprintf("%s/%s", handler.Kind(), UID)
	dotKey := fmt.Sprintf("%s.%s", handler.Kind(), UID)
	for _, target := range targets {
		g := glob.MustCompile(target)
		if g.Match(slashKey) || g.Match(dotKey) {
			return true
		}
	}
	return false
}
