package grizzly

import (
	"encoding/base64"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os/exec"
	"runtime"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gorilla/websocket"
	"github.com/grafana/grizzly/pkg/config"
	log "github.com/sirupsen/logrus"
)

type GrizzlyServer struct {
	proxy        *httputil.ReverseProxy
	Registry     Registry
	Parser       WatchParser
	Url          string
	User         string
	Token        string
	UserAgent    string
	ResourcePath string
	Opts         Opts
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func NewGrizzlyServer(registry Registry, parser WatchParser, resourcePath string, opts Opts) (*GrizzlyServer, error) {
	server := GrizzlyServer{
		Parser:       parser,
		UserAgent:    "grizzly",
		Opts:         opts,
		ResourcePath: resourcePath,
	}
	context, err := config.CurrentContext()
	if err != nil {
		return nil, err
	}
	server.Url = context.Grafana.URL
	server.User = context.Grafana.User
	server.Token = context.Grafana.Token
	u, err := url.Parse(server.Url)
	if err != nil {
		return nil, err
	}
	server.proxy = &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(u)

			if server.User != "" {
				header := fmt.Sprintf("%s:%s", server.User, server.Token)
				encoded := base64.StdEncoding.EncodeToString([]byte(header))
				r.Out.Header.Set("Authorization", "Bearer "+encoded)
			} else {
				r.Out.Header.Set("Authorization", "Bearer "+server.Token)
			}

			r.Out.Header.Del("Origin")
			r.Out.Header.Set("User-Agent", "Grizzly Proxy Server")
		},
	}
	return &server, nil
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

	"/api/access-control/user/actions": `{
        "dashboards:write": true,
	}`,
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

func (p *GrizzlyServer) Start() error {
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
					r.Get(endpoint.Url, endpoint.Handler)
				case "POST":
					r.Post(endpoint.Url, endpoint.Handler)
				default:
					return fmt.Errorf("Unknown endpoint method %s for handler %s", endpoint.Method, handler.Kind())
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

	r.Get("/api/live/ws", p.wsHandler)
	if p.Opts.ProxyPort == 0 {
		p.Opts.ProxyPort = 8080
	}
	if p.Opts.OpenBrowser {
		var url string
		if p.Opts.IsDir {
			url = fmt.Sprintf("http://localhost:%d", p.Opts.ProxyPort)
		} else {
			resources, err := p.Parser.Parse()
			if err != nil {
				return err
			}
			if len(resources) > 1 {
				url = fmt.Sprintf("http://localhost:%d", p.Opts.ProxyPort)
			} else if len(resources) == 0 {
				return fmt.Errorf("no resources found to proxy")
			} else {
				resource := resources[0]
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
				url = fmt.Sprintf("http://localhost:%d%s", p.Opts.ProxyPort, proxyURL)
			}
		}
		p.openBrowser(url)
	}
	fmt.Printf("Listening on http://localhost:%d\n", p.Opts.ProxyPort)
	return http.ListenAndServe(fmt.Sprintf(":%d", p.Opts.ProxyPort), r)
}

func (p *GrizzlyServer) openBrowser(url string) {
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

func (p *GrizzlyServer) blockHandler(response string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}
}

// ProxyRequestHandler handles the http request using proxy
func (p *GrizzlyServer) ProxyRequestHandler(w http.ResponseWriter, r *http.Request) {
	p.proxy.ServeHTTP(w, r)
}

func (p *GrizzlyServer) wsHandler(w http.ResponseWriter, r *http.Request) {
	_, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	p.proxy.ServeHTTP(w, r)
}

func (p *GrizzlyServer) RootHandler(w http.ResponseWriter, r *http.Request) {
	resources, err := p.Parser.Parse()
	if err != nil {
		log.Error("Error: ", err)
		http.Error(w, fmt.Sprintf("Error: %s", err), 500)
		return
	}

	templateVars := map[string]any{
		"Resources": resources,
	}
	if err := templates.ExecuteTemplate(w, "proxy/index.html.tmpl", templateVars); err != nil {
		log.Error("Error while executing template: ", err)
		http.Error(w, fmt.Sprintf("Error: %s", err), 500)
		return
	}
}
