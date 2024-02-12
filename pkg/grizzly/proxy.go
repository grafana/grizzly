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

type ProxyServer struct {
	proxy        *httputil.ReverseProxy
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

func NewProxyServer(parser WatchParser, resourcePath string, opts Opts) (*ProxyServer, error) {
	server := ProxyServer{
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
	"/avatar/",
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
	"/avatar/*":    "",
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

func (p *ProxyServer) Start() error {
	assetsFS, err := fs.Sub(embedFS, "embed/assets")
	if err != nil {
		return fmt.Errorf("could not create a sub-tree from the embedded assets FS: %w", err)
	}

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Handle("/grizzly/assets/*", http.StripPrefix("/grizzly/assets/", http.FileServer(http.FS(assetsFS))))

	for _, handler := range Registry.Handlers {
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
	fmt.Printf("Listening on http://localhost:8080\n")
	if p.Opts.OpenBrowser {
		p.openBrowser("http://localhost:8080")
	}
	return http.ListenAndServe(":8080", r)
}

func (p *ProxyServer) openBrowser(url string) {
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

func (p *ProxyServer) blockHandler(response string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}
}

// ProxyRequestHandler handles the http request using proxy
func (p *ProxyServer) ProxyRequestHandler(w http.ResponseWriter, r *http.Request) {
	p.proxy.ServeHTTP(w, r)
}

func (p *ProxyServer) wsHandler(w http.ResponseWriter, r *http.Request) {
	_, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	p.proxy.ServeHTTP(w, r)
}

func (p *ProxyServer) RootHandler(w http.ResponseWriter, r *http.Request) {
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
