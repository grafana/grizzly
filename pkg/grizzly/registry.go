package grizzly

import (
	"fmt"
	"net/http/httputil"
	"strings"

	"github.com/gobwas/glob"
)

// Provider describes a single Endpoint Provider
type Provider interface {
	Name() string
	Group() string
	Version() string
	APIVersion() string
	GetHandlers() []Handler
}

type ProxyProvider interface {
	// SetupProxy establishes the proxy connection
	SetupProxy() (*httputil.ReverseProxy, error)
}

// ProviderSet records providers
type Registry struct {
	Providers    []Provider
	Handlers     map[string]Handler
	HandlerOrder []Handler
}

// NewRegistry returns an empty registry
func NewRegistry(providers []Provider) Registry {
	registry := Registry{
		Handlers:     map[string]Handler{},
		HandlerOrder: []Handler{},
	}

	registry.Providers = providers
	for _, provider := range providers {
		for _, handler := range provider.GetHandlers() {
			registry.Handlers[handler.Kind()] = handler
			registry.HandlerOrder = append(registry.HandlerOrder, handler)
		}
	}
	return registry
}

// GetHandler returns a single provider based upon a JSON path
func (r *Registry) GetHandler(kind string) (Handler, error) {
	handler, exists := r.Handlers[kind]
	if !exists {
		return nil, fmt.Errorf("couldn't find a handler for %s: %w", kind, ErrHandlerNotFound)
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
		if (strings.Contains(target, "/") && strings.Split(target, "/")[0] == key) ||
			(strings.Contains(target, ".") && strings.Split(target, ".")[0] == key) {
			return true
		} else if strings.EqualFold(target, key) {
			return true
		}
	}
	return false
}

// ResourceMatchesTarget identifies whether a resource is in a target list
func (r *Registry) ResourceMatchesTarget(kind string, uid string, targets []string) bool {
	if len(targets) == 0 {
		return true
	}
	// I mistakenly assumed 'dot' was a special character for globs, so opted for '/' as separator.
	// This keeps back-compat
	slashKey := fmt.Sprintf("%s/%s", kind, uid)
	dotKey := fmt.Sprintf("%s.%s", kind, uid)
	for _, target := range targets {
		if strings.Contains(target, ".") || strings.Contains(target, "/") {
			g, err := glob.Compile(target)
			if err != nil {
				continue
			}

			if g.Match(slashKey) || g.Match(dotKey) {
				return true
			}
		} else if strings.EqualFold(target, kind) {
			return true
		}
	}
	return false
}

func (r *Registry) Sort(resources Resources) Resources {
	sorted := NewResources()
	resourceByKind := resources.GroupByKind()

	for _, handler := range r.HandlerOrder {
		handlerResources := resourceByKind[handler.Kind()]
		sorted.Merge(handler.Sort(handlerResources))
	}

	return sorted
}

func (r *Registry) Detect(data any) string {
	m, ok := data.(map[string]any)
	if !ok {
		return ""
	}
	for _, handler := range r.HandlerOrder {
		if handler.Detect(m) {
			return handler.Kind()
		}
	}
	return ""
}

func (r *Registry) GetProxyProvider() (*ProxyProvider, error) {
	var proxyProvider *ProxyProvider
	for _, provider := range r.Providers {
		pp, ok := provider.(ProxyProvider)
		if ok {
			if proxyProvider == nil {
				proxyProvider = &pp
			} else {
				return nil, fmt.Errorf("only one proxy provider currently supported")
			}
		}
	}
	return proxyProvider, nil
}
