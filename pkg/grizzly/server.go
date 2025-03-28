package grizzly

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httputil"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gorilla/websocket"
	"github.com/grafana/grizzly/internal/httputils"
	"github.com/grafana/grizzly/internal/livereload"
	"github.com/grafana/grizzly/internal/logger"
	"github.com/hashicorp/go-multierror"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	proxy *httputil.ReverseProxy
	// listenAddr specifies the address the server will listen on.
	// Note: if left empty, 0.0.0.0 is assumed.
	listenAddr  string
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
	WatchPaths     []string
	watchScript    string
	OnlySpec       bool
	OutputFormat   string
	watch          bool
	proxySubPath   string
}

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func NewGrizzlyServer(registry Registry, resourcePath string, listenAddr string, port int) (*Server, error) {
	prov, err := registry.GetProxyProvider()
	if err != nil {
		return nil, err
	}

	if prov == nil {
		return nil, fmt.Errorf("no proxy provider found")
	}

	proxy, subPath, err := (*prov).SetupProxy()
	if err != nil {
		return nil, err
	}

	return &Server{
		Registry:     registry,
		Resources:    NewResources(),
		UserAgent:    "grizzly",
		ResourcePath: resourcePath,
		listenAddr:   listenAddr,
		port:         port,
		proxy:        proxy,
		proxySubPath: strings.TrimSuffix(subPath, "/"),
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

func (s *Server) Watch(watchPaths []string) {
	s.watch = true
	s.WatchPaths = watchPaths
}

func (s *Server) WatchScript(script string) {
	s.watchScript = script
}

func (s *Server) SetFormatting(onlySpec bool, outputFormat string) {
	s.OnlySpec = onlySpec
	s.OutputFormat = outputFormat
}

func (s *Server) staticProxyConfig() StaticProxyConfig {
	return StaticProxyConfig{
		ProxyGet: []string{
			"/public/*",
			"/avatar/*",
		},
		MockGet: map[string]string{
			"/api/ma/events":       "[]",
			"/api/live/publish":    "[]",
			"/api/live/list":       "[]",
			"/api/user/orgs":       "[]",
			"/api/search":          "[]",
			"/api/usage/*":         "[]",
			"/api/frontend/assets": "{}",
			"/api/org/preferences": "{}",
			"/api/org/users":       "[]",

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

			"/apis/banners.grafana.app/v0alpha1/namespaces/{stack}/announcement-banners": `{
      "kind": "AnnouncementBannerList",
      "apiVersion": "banners.grafana.app/v0alpha1",
      "metadata": {"resourceVersion": "29"}
    }`,
		},
		MockPost: map[string]string{
			"/api/frontend-metrics": "[]",
			"/api/search-v2":        "[]",
			"/api/live/publish":     "{}",
			"/api/ma/events":        "null",
		},
	}
}

func (s *Server) applyStaticProxyConfig(r chi.Router, config StaticProxyConfig) {
	for _, pattern := range config.ProxyGet {
		r.Get(s.proxySubPath+pattern, s.ProxyRequestHandler)
	}
	for _, pattern := range config.ProxyPost {
		r.Post(s.proxySubPath+pattern, s.ProxyRequestHandler)
	}
	for pattern, response := range config.MockGet {
		r.Get(s.proxySubPath+pattern, s.mockHandler(response))
	}
	for pattern, response := range config.MockPost {
		r.Post(s.proxySubPath+pattern, s.mockHandler(response))
	}
}

func (s *Server) Start() error {
	assetsFS, err := fs.Sub(embedFS, "embed/assets")
	if err != nil {
		return fmt.Errorf("could not create a sub-tree from the embedded assets FS: %w", err)
	}

	r := chi.NewRouter()

	color := true
	if runtime.GOOS == "windows" {
		color = false
	}

	r.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: logger.DecorateAtLevel(log.StandardLogger(), log.DebugLevel), NoColor: !color}))
	r.Handle("/grizzly/assets/*", http.StripPrefix("/grizzly/assets/", http.FileServer(http.FS(assetsFS))))

	s.applyStaticProxyConfig(r, s.staticProxyConfig())

	for _, handler := range s.Registry.Handlers {
		proxyConfigProvider, ok := handler.(ProxyConfiguratorProvider)
		if !ok {
			continue
		}

		log.WithField("handler", handler.Kind()).Debug("registering proxy configuration")

		proxyConfig := proxyConfigProvider.ProxyConfigurator()
		for _, endpoint := range proxyConfig.Endpoints(*s) {
			switch endpoint.Method {
			case http.MethodGet:
				r.Get(s.proxySubPath+endpoint.URL, endpoint.Handler)
			case http.MethodPost:
				r.Post(s.proxySubPath+endpoint.URL, endpoint.Handler)
			case http.MethodPut:
				r.Put(s.proxySubPath+endpoint.URL, endpoint.Handler)
			default:
				return fmt.Errorf("unknown endpoint method %s for handler %s", endpoint.Method, handler.Kind())
			}
		}

		s.applyStaticProxyConfig(r, proxyConfig.StaticEndpoints())
	}

	r.Get("/", s.rootHandler)
	r.Get("/grizzly/{kind}/{name}", s.iframeHandler)
	r.Get("/livereload", livereload.Handler(upgrader))

	if s.watchScript != "" {
		var b []byte
		b, err = s.executeWatchScript()
		if err != nil {
			return err
		}
		_, err = s.ParseBytes(b)
	} else {
		_, err = s.ParseResources(s.ResourcePath)
	}
	if err != nil {
		log.Warn(err.Error())
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
		for _, path := range s.WatchPaths {
			err = watcher.Add(path)
			if err != nil {
				return err
			}
		}
		err = watcher.Watch()
		if err != nil {
			return err
		}
	}

	log.Infof("Listening on %s\n", s.url("/"))
	log.Debugf("Listening address '%s'", max(s.listenAddr, "0.0.0.0"))

	return http.ListenAndServe(fmt.Sprintf("%s:%d", s.listenAddr, s.port), r)
}

func (s *Server) ParseResources(resourcePath string) (Resources, error) {
	resources, err := s.parser.Parse(resourcePath, s.parserOpts)
	s.parserErr = err
	s.Resources.Merge(resources)
	return resources, err
}

func (s *Server) ParseBytes(b []byte) (Resources, error) {
	f, err := os.CreateTemp(".", fmt.Sprintf("*.%s", s.OutputFormat))
	if err != nil {
		return Resources{}, err
	}
	defer os.Remove(f.Name())

	_, err = f.Write(b)
	if err != nil {
		return Resources{}, err
	}

	err = f.Close()
	if err != nil {
		return Resources{}, err
	}

	resources, err := s.parser.Parse(f.Name(), s.parserOpts)
	s.parserErr = err
	s.Resources.Merge(resources)
	return resources, err
}

func (s *Server) url(path string) string {
	if len(path) == 0 || path[0] != '/' {
		path = "/" + path
	}

	return fmt.Sprintf("http://localhost:%d%s", s.port, path)
}

func (s *Server) updateWatchedResource(name string) error {
	var resources Resources
	var err error

	log.Debugf("[watcher] Updating watched resource")

	if s.watchScript != "" {
		var b []byte
		b, err = s.executeWatchScript()
		if err != nil {
			return err
		}
		resources, err = s.ParseBytes(b)
	} else {
		resources, err = s.ParseResources(name)
	}
	if errors.As(err, &UnrecognisedFormatError{}) {
		uerr := err.(UnrecognisedFormatError)
		log.Infof("[watcher] Skipping %s", uerr.File)
		return nil
	}
	if err != nil {
		log.Errorf("[watcher] Error: %s", err)
		return err
	}

	for _, resource := range resources.AsList() {
		handler, err := s.Registry.GetHandler(resource.Kind())
		if err != nil {
			log.Warnf("[watcher] Error: %s", err)
			continue
		}
		_, ok := handler.(ProxyConfigurator)
		if ok {
			log.Infof("[watcher] Changes detected. Reloading %s", resource.Ref())
			livereload.ReloadDashboard(resource.Name())
		}
	}
	return nil
}

func (s *Server) executeWatchScript() ([]byte, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	log.Debugf("[watch script] executing %s", s.watchScript)

	cmd := exec.Command("sh", "-c", s.watchScript)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if stderr.Len() > 0 {
		log.Errorf("[watch script] %s", stderr.String())
	}
	if err != nil {
		return nil, fmt.Errorf("watch script failed: %w", err)
	}
	return stdout.Bytes(), nil
}

func (s *Server) mockHandler(response string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		httputils.Write(w, []byte(response))
	}
}

// ProxyRequestHandler handles the http request using proxy
func (s *Server) ProxyRequestHandler(w http.ResponseWriter, r *http.Request) {
	s.proxy.ServeHTTP(w, r)
}

// rootHandler lists all local proxyable resources
func (s *Server) iframeHandler(w http.ResponseWriter, r *http.Request) {
	kind := chi.URLParam(r, "kind")
	name := chi.URLParam(r, "name")
	handler, err := s.Registry.GetHandler(kind)
	if err != nil {
		httputils.Error(w, fmt.Sprintf("Error getting handler for %s/%s", kind, name), err, http.StatusInternalServerError)
		return
	}

	proxyConfigProvider, ok := handler.(ProxyConfiguratorProvider)
	if !ok {
		httputils.Error(w, fmt.Sprintf("%s is not supported by the Grizzly server", kind), fmt.Errorf("%s is not supported by the Grizzly server", kind), http.StatusInternalServerError)
		return
	}

	proxyConfig := proxyConfigProvider.ProxyConfigurator()
	templateVars := map[string]any{
		"Port":           s.port,
		"IframeURL":      s.proxySubPath + proxyConfig.ProxyURL(name),
		"CurrentContext": s.CurrentContext,
	}

	if err := templates.ExecuteTemplate(w, "proxy/iframe.html.tmpl", templateVars); err != nil {
		httputils.Error(w, "Error while executing template", err, http.StatusInternalServerError)
		return
	}
}

func (s *Server) rootHandler(w http.ResponseWriter, _ *http.Request) {
	var parseErrors []error

	if s.parserErr != nil {
		if merr, ok := s.parserErr.(*multierror.Error); ok {
			parseErrors = merr.Errors
		} else {
			parseErrors = []error{s.parserErr}
		}
	}

	templateVars := map[string]any{
		"Resources":      s.Resources,
		"ParseErrors":    parseErrors,
		"ServerPort":     s.port,
		"CurrentContext": s.CurrentContext,
	}
	if err := templates.ExecuteTemplate(w, "proxy/index.html.tmpl", templateVars); err != nil {
		httputils.Error(w, "Error while executing template", err, http.StatusInternalServerError)
		return
	}
}

func (s *Server) UpdateResource(resource Resource) error {
	out, _, _, err := Format(s.Registry, s.ResourcePath, &resource, resource.Source.Format, !resource.Source.WithEnvelope)
	if err != nil {
		return fmt.Errorf("error formatting content: %s", err)
	}

	existing, found := s.Resources.Find(resource.Ref())
	if !found {
		return fmt.Errorf("%s not found", resource.Ref())
	}

	if !existing.Source.Rewritable {
		return fmt.Errorf("the source for this %s is not rewritable", resource.Kind())
	}

	return WriteFile(existing.Source.Path, out)
}
