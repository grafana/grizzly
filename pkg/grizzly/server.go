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

	Registry       Registry
	CurrentContext string
	Resources      Resources
	UserAgent      string
	ResourcePath   string
	OnlySpec       bool
	OutputFormat   string
	watch          bool
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func NewGrizzlyServer(registry Registry, resourcePath string, port int) (*Server, error) {
	prov, err := registry.GetProxyProvider()
	if err != nil {
		return nil, err
	}

	if prov == nil {
		return nil, fmt.Errorf("no proxy provider found")
	}

	proxy, err := (*prov).SetupProxy()
	if err != nil {
		return nil, err
	}

	return &Server{
		Registry:     registry,
		Resources:    NewResources(),
		UserAgent:    "grizzly",
		ResourcePath: resourcePath,
		port:         port,
		proxy:        proxy,
	}, nil
}

func (s *Server) SetParser(parser Parser, parserOpts ParserOptions) {
	s.parser = parser
	s.parserOpts = parserOpts
}

func (s *Server) SetContext(currentContext string) {
	s.CurrentContext = currentContext
}

func (s *Server) OpenBrowser() {
	s.openBrowser = true
}

func (s *Server) Watch() {
	s.watch = true
}

func (s *Server) SetFormatting(onlySpec bool, outputFormat string) {
	s.OnlySpec = onlySpec
	s.OutputFormat = outputFormat
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

func (s *Server) Start() error {
	assetsFS, err := fs.Sub(embedFS, "embed/assets")
	if err != nil {
		return fmt.Errorf("could not create a sub-tree from the embedded assets FS: %w", err)
	}

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Handle("/grizzly/assets/*", http.StripPrefix("/grizzly/assets/", http.FileServer(http.FS(assetsFS))))

	for _, handler := range s.Registry.Handlers {
		proxyHandler, ok := handler.(ProxyHandler)
		if ok {
			for _, endpoint := range proxyHandler.GetProxyEndpoints(*s) {
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
		r.Get(pattern, s.ProxyRequestHandler)
	}
	for _, pattern := range mustProxyPOST {
		r.Post(pattern, s.ProxyRequestHandler)
	}
	for pattern, response := range blockJSONget {
		r.Get(pattern, s.blockHandler(response))
	}
	for pattern, response := range blockJSONpost {
		r.Post(pattern, s.blockHandler(response))
	}
	r.Get("/", s.RootHandler)
	r.Get("/grizzly/{kind}/{name}", s.IframeHandler)
	r.Get("/api/live/ws", livereload.LiveReloadHandlerFunc(upgrader))

	if _, err := s.ParseResources(s.ResourcePath); err != nil {
		fmt.Print(err)
	}

	if s.openBrowser {
		browser, err := NewBrowserInterface(s.Registry, s.ResourcePath, s.port)
		if err != nil {
			return err
		}
		err = browser.Open(s.Resources)
		if err != nil {
			return err
		}
	}
	if s.watch {
		livereload.Initialize()
		watcher, err := NewWatcher(s.updateWatchedResource)
		if err != nil {
			return err
		}
		err = watcher.Watch(s.ResourcePath)
		if err != nil {
			return err
		}
	}

	fmt.Printf("Listening on %s\n", s.URL("/"))
	return http.ListenAndServe(fmt.Sprintf(":%d", s.port), r)
}

func (s *Server) ParseResources(resourcesPath string) (Resources, error) {
	resources, err := s.parser.Parse(resourcesPath, s.parserOpts)
	s.parserErr = err
	s.Resources.Merge(resources)
	return resources, err
}

func (s *Server) URL(path string) string {
	if len(path) == 0 || path[0] != '/' {
		path = "/" + path
	}

	return fmt.Sprintf("http://localhost:%d%s", s.port, path)
}

func (s *Server) updateWatchedResource(name string) error {
	resources, err := s.ParseResources(name)
	if errors.As(err, &UnrecognisedFormatError{}) {
		log.Printf("Skipping %s", name)
		return nil
	}
	if err != nil {
		log.Error("Error: ", err)
		return err
	}
	for _, resource := range resources.AsList() {
		handler, err := s.Registry.GetHandler(resource.Kind())
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
func (s *Server) blockHandler(response string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(response)); err != nil {
			log.Errorf("error writing response: %v", err)
		}
	}
}

// ProxyRequestHandler handles the http request using proxy
func (s *Server) ProxyRequestHandler(w http.ResponseWriter, r *http.Request) {
	s.proxy.ServeHTTP(w, r)
}

// RootHandler lists all local proxyable resources
func (s *Server) IframeHandler(w http.ResponseWriter, r *http.Request) {
	kind := chi.URLParam(r, "kind")
	name := chi.URLParam(r, "name")
	handler, err := s.Registry.GetHandler(kind)
	if err != nil {
		SendError(w, fmt.Sprintf("Error getting handler for %s/%s", kind, name), err, 500)
		return
	}
	proxyHandler, ok := handler.(ProxyHandler)
	if !ok {
		SendError(w, fmt.Sprintf("%s is not supported by the Grizzly server", kind), fmt.Errorf("%s is not supported by the Grizzly server", kind), 500)
		return
	}
	url := proxyHandler.ProxyURL(name)
	templateVars := map[string]string{
		"IframeURL":      url,
		"CurrentContext": s.CurrentContext,
	}

	if err := templates.ExecuteTemplate(w, "proxy/iframe.html.tmpl", templateVars); err != nil {
		SendError(w, "Error while executing template", err, 500)
		return
	}
}

func (s *Server) RootHandler(w http.ResponseWriter, _ *http.Request) {
	var parseErrors []error

	if s.parserErr != nil {
		if merr, ok := s.parserErr.(*multierror.Error); ok {
			parseErrors = merr.Errors
		} else {
			parseErrors = []error{s.parserErr}
		}
	}

	templateVars := map[string]any{
		"Resources":      s.Resources.AsList(),
		"ParseErrors":    parseErrors,
		"ServerPort":     s.port,
		"CurrentContext": s.CurrentContext,
	}
	if err := templates.ExecuteTemplate(w, "proxy/index.html.tmpl", templateVars); err != nil {
		SendError(w, "Error while executing template", err, 500)
		return
	}
}

func (s *Server) UpdateResource(name string, resource Resource) error {
	out, _, _, err := Format(s.Registry, s.ResourcePath, &resource, s.OutputFormat, s.OnlySpec)
	if err != nil {
		return fmt.Errorf("error formatting content: %s", err)
	}

	existing, found := s.Resources.Find(NewResourceRef("Dashboard", name))
	if !found {
		return fmt.Errorf("dashboard with UID %s not found", name)
	}
	if !existing.Source.Rewritable {
		return fmt.Errorf("the source for this dashboard is not rewritable")
	}
	return os.WriteFile(existing.Source.Path, out, 0644)
}
