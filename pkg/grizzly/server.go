package grizzly

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gorilla/websocket"
	"github.com/grafana/grizzly/pkg/grizzly/livereload"
	"github.com/hashicorp/go-multierror"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	proxy       *httputil.ReverseProxy
	port        int
	openBrowser bool

	parser     Parser
	parserOpts ParserOptions
	parserErr  error

	Registry     Registry
	Resources    Resources
	UserAgent    string
	ResourcePath string
	OnlySpec     bool
	OutputFormat string
	Watch        bool
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func NewGrizzlyServer(registry Registry, parser Parser, parserOpts ParserOptions, resourcePath string, port int, openBrowser, watch, onlySpec bool, outputFormat string) (*Server, error) {
	prov, err := registry.GetProxyProvider()
	if err != nil {
		return nil, err
	}

	proxy, err := (*prov).SetupProxy()
	if err != nil {
		return nil, err
	}

	return &Server{
		Registry:     registry,
		Resources:    NewResources(),
		parser:       parser,
		parserOpts:   parserOpts,
		UserAgent:    "grizzly",
		ResourcePath: resourcePath,
		port:         port,
		openBrowser:  openBrowser,
		OnlySpec:     onlySpec,
		OutputFormat: outputFormat,
		proxy:        proxy,
		Watch:        watch,
	}, nil
}

var mustProxyGET = []string{
	"/public/*",
	"/api/datasources/proxy/*",
	"/api/datasources/*",
	"/api/plugins/*",
	"/avatar/*",
}
var mustProxyPOST = []string{
	"/api/datasources/proxy/*",
	"/api/ds/query",
}
var blockJSONget = map[string]string{
	"/api/ma/events":    "[]",
	"/api/live/publish": "[]",
	"/api/live/list":    "[]",
	"/api/user/orgs":    "[]",
	"/api/annotations":  "[]",
	"/api/search":       "[]",
	"/api/usage/*":      "[]",

	"/api/access-control/user/actions": `{"dashboards:write": true}`,
	"/api/prometheus/grafana/api/v1/rules": `{
      "status": "success",
      "data": { "groups": [] }
    }`,
	"/api/folders": "[]",
	"/api/recording-rules/writer": `{
      "id": "cojWep7Vz",
      "data_source_uid": "grafanacloud-prom",
      "remote_write_path": "/api/prom/push"
    }`,
}

var blockJSONpost = map[string]string{
	"/api/frontend-metrics": "[]",
	"/api/search-v2":        "[]",
	"/api/live/publish":     "{}",
	"/api/ma/events":        "null",
}

func (p *Server) Start() error {
	assetsFS, err := fs.Sub(embedFS, "embed/assets")
	if err != nil {
		return fmt.Errorf("could not create a sub-tree from the embedded assets FS: %w", err)
	}

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Handle("/grizzly/assets/*", http.StripPrefix("/grizzly/assets/", http.FileServer(http.FS(assetsFS))))

	for _, handler := range p.Registry.Handlers {
		proxyHandler, ok := handler.(ProxyHandler)
		if ok {
			for _, endpoint := range proxyHandler.GetProxyEndpoints(*p) {
				switch endpoint.Method {
				case "GET":
					r.Get(endpoint.URL, endpoint.Handler)
				case "POST":
					r.Post(endpoint.URL, endpoint.Handler)
				default:
					return fmt.Errorf("unknown endpoint method %s for handler %s", endpoint.Method, handler.Kind())
				}
			}
		}
	}
	for _, pattern := range mustProxyGET {
		r.Get(pattern, p.ProxyRequestHandler)
	}
	for _, pattern := range mustProxyPOST {
		r.Post(pattern, p.ProxyRequestHandler)
	}
	for pattern, response := range blockJSONget {
		r.Get(pattern, p.blockHandler(response))
	}
	for pattern, response := range blockJSONpost {
		r.Post(pattern, p.blockHandler(response))
	}
	r.Get("/", p.RootHandler)
	r.Get("/api/live/ws", livereload.LiveReloadHandlerFunc(upgrader))

	if _, err := p.ParseResources(p.ResourcePath); err != nil {
		fmt.Print(err)
	}

	if p.openBrowser {
		browser, err := NewBrowserInterface(p.Registry, p.ResourcePath, p.port)
		if err != nil {
			return err
		}
		err = browser.Open(p.Resources)
		if err != nil {
			return err
		}
	}
	if p.Watch {
		livereload.Initialize()
		watcher, err := NewWatcher(p.updateWatchedResource)
		if err != nil {
			return err
		}
		err = watcher.Watch(p.ResourcePath)
		if err != nil {
			return err
		}
	}

	fmt.Printf("Listening on %s\n", p.URL("/"))
	return http.ListenAndServe(fmt.Sprintf(":%d", p.port), r)
}

func (p *Server) ParseResources(resourcesPath string) (Resources, error) {
	resources, err := p.parser.Parse(resourcesPath, p.parserOpts)
	p.parserErr = err
	p.Resources.Merge(resources)
	return resources, err
}

func (p *Server) URL(path string) string {
	if len(path) == 0 || path[0] != '/' {
		path = "/" + path
	}

	return fmt.Sprintf("http://localhost:%d%s", p.port, path)
}

func (p *Server) updateWatchedResource(name string) error {
	resources, err := p.ParseResources(name)
	if errors.As(err, &UnrecognisedFormatError{}) {
		log.Printf("Skipping %s", name)
		return nil
	}
	if err != nil {
		log.Error("Error: ", err)
		return err
	}
	for _, resource := range resources.AsList() {
		handler, err := p.Registry.GetHandler(resource.Kind())
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}
		_, ok := handler.(ProxyHandler)
		if ok {
			log.Info("Changes detected. Applying ", name)
			err = livereload.Reload(resource.Kind(), resource.Name(), resource.Spec())
			if err != nil {
				return err
			}
		}
	}
	return nil
}
func (p *Server) blockHandler(response string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(response)); err != nil {
			log.Errorf("error writing response: %v", err)
		}
	}
}

// ProxyRequestHandler handles the http request using proxy
func (p *Server) ProxyRequestHandler(w http.ResponseWriter, r *http.Request) {
	p.proxy.ServeHTTP(w, r)
}

func (p *Server) RootHandler(w http.ResponseWriter, _ *http.Request) {
	var parseErrors []error

	if p.parserErr != nil {
		if merr, ok := p.parserErr.(*multierror.Error); ok {
			parseErrors = merr.Errors
		} else {
			parseErrors = []error{p.parserErr}
		}
	}

	templateVars := map[string]any{
		"Resources":   p.Resources.AsList(),
		"ParseErrors": parseErrors,
		"ServerPort":  p.port,
	}
	if err := templates.ExecuteTemplate(w, "proxy/index.html.tmpl", templateVars); err != nil {
		SendError(w, "Error while executing template", err, 500)
		return
	}
}

func (p *Server) UpdateResource(name string, resource Resource) error {
	out, _, _, err := Format(p.Registry, p.ResourcePath, &resource, p.OutputFormat, p.OnlySpec)
	if err != nil {
		return fmt.Errorf("error formatting content: %s", err)
	}

	existing, found := p.Resources.Find(NewResourceRef("Dashboard", name))
	if !found {
		return fmt.Errorf("dashboard with UID %s not found", name)
	}
	if !existing.Source.Rewritable {
		return fmt.Errorf("the source for this dashboard is not rewritable")
	}
	return os.WriteFile(existing.Source.Path, out, 0644)
}
