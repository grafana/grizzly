package grizzly

import (
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httputil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gorilla/websocket"
	"github.com/grafana/grizzly/pkg/grizzly/livereload"
	"github.com/hashicorp/go-multierror"
	log "github.com/sirupsen/logrus"
	"gopkg.in/fsnotify.v1"
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

	httpEndpointConfig := HttpEndpointConfig{
		Port: p.port,
	}
	for _, handler := range p.Registry.Handlers {
		proxyHandler, ok := handler.(ProxyHandler)
		if ok {
			for _, endpoint := range proxyHandler.GetProxyEndpoints(*p, httpEndpointConfig) {
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

	livereload.Initialize()
	r.Get("/livereload.js", livereload.LiveReloadJSHandler)
	r.Get("/livereload", livereload.LiveReloadHandlerFunc(upgrader))
	r.Get("/api/live/ws", p.wsHandler)

	if err := p.ParseResources(p.ResourcePath); err != nil {
		return err
	}

	if p.openBrowser {
		path := "/"

		stat, err := os.Stat(p.ResourcePath)
		if err != nil {
			return err
		}

		if !stat.IsDir() && p.Resources.Len() == 0 {
			return fmt.Errorf("no resources found to proxy")
		}

		if !stat.IsDir() && p.Resources.Len() == 1 {
			resource := p.Resources.First()
			handler, err := p.Registry.GetHandler(resource.Kind())
			if err != nil {
				return err
			}
			proxyHandler, ok := handler.(ProxyHandler)
			if !ok {
				uid, err := handler.GetUID(resource)
				if err != nil {
					return err
				}
				return fmt.Errorf("kind %s (for resource %s) does not support proxying", resource.Kind(), uid)
			}
			proxyURL, err := proxyHandler.ProxyURL(resource)
			if err != nil {
				return err
			}
			path = proxyURL
		}

		p.openInBrowser(p.URL(path))
	}
	if p.Watch {
		err := p.setupWatch()
		if err != nil {
			return err
		}
	}

	fmt.Printf("Listening on %s\n", p.URL("/"))
	return http.ListenAndServe(fmt.Sprintf(":%d", p.port), r)
}

func (p *Server) ParseResources(resourcesPath string) error {
	resources, err := p.parser.Parse(resourcesPath, p.parserOpts)
	p.parserErr = err
	if err != nil {
		return err
	}

	p.Resources.Merge(resources)

	return nil
}

func (p *Server) URL(path string) string {
	if len(path) == 0 || path[0] != '/' {
		path = "/" + path
	}

	return fmt.Sprintf("http://localhost:%d%s", p.port, path)
}

func (p *Server) openInBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}

func (p *Server) setupWatch() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	//done := make(chan bool)
	go func() {
		log.Info("Watching for changes")
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Info("Changes detected. Parsing ")
					resources, err := p.parser.Parse(event.Name, p.parserOpts)
					if err != nil {
						log.Error("Error: ", err)
					}
					resource, bool := resources.FindByFilename(event.Name)
					if bool {
						handler, err := p.Registry.GetHandler(resource.Kind())
						if err != nil {
							log.Printf("Error: %v", err)
						} else {
							proxyHandler, ok := handler.(ProxyHandler)
							if ok {
								u, err := proxyHandler.ProxyURL(resource)
								if err != nil {
									log.Print(err)
								} else {
									livereload.Reload(u)
								}
							}
						}
					} else {
						log.Printf("%s not found in resources", event.Name)
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

	err = filepath.WalkDir(p.ResourcePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return watcher.Add(path)
		}

		return nil
	})
	if err != nil {
		return err
	}

	//<-done

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

func (p *Server) wsHandler(w http.ResponseWriter, r *http.Request) {
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
