package grizzly

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gorilla/websocket"
	"github.com/grafana/grizzly/pkg/config"
	log "github.com/sirupsen/logrus"
)

type ProxyServer struct {
	proxy        *httputil.ReverseProxy
	parser       WatchParser
	url          string
	user         string
	token        string
	userAgent    string
	resourcePath string
	isLegacy     bool
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func NewProxyServer(parser WatchParser, resourcePath string, isLegacyJSON bool) (*ProxyServer, error) {
	server := ProxyServer{
		parser:       parser,
		userAgent:    "grizzly",
		isLegacy:     isLegacyJSON,
		resourcePath: resourcePath,
	}
	context, err := config.CurrentContext()
	if err != nil {
		return nil, err
	}
	server.url = context.Grafana.URL
	server.user = context.Grafana.User
	server.token = context.Grafana.Token
	u, err := url.Parse(server.url)
	if err != nil {
		return nil, err
	}
	server.proxy = &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(u)
			r.Out.Host = r.In.Host // if desired
			if server.user != "" {
				header := fmt.Sprintf("%s:%s", server.user, server.token)
				encoded := base64.StdEncoding.EncodeToString([]byte(header))
				r.Out.Header.Set("Authorization", "Bearer "+encoded)
			} else {
				r.Out.Header.Set("Authorization", "Bearer "+server.token)
			}
			r.Out.Header.Set("User-Agent", "Grizzly Proxy Server")
		},
	}
	return &server, nil
}

var mustProxyGET = []string{
	"/public/*",
	"/api/datasources/proxy/*",
	"/api/datasources/*",
}
var mustProxyPOST = []string{
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
      status: "success",
      data: { groups: [] },
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
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/d/{uid}/{slug}", p.RootDashboardPageHandler)
	r.Get("/api/dashboards/uid/{uid}", p.DashboardJSONGetHandler)
	r.Post("/api/dashboards/db/", p.DashboardJSONPostHandler)
	for _, url := range mustProxyGET {
		r.Get(url, p.ProxyRequestHandler)
	}
	for _, url := range mustProxyPOST {
		r.Post(url, p.ProxyRequestHandler)
	}
	for url, response := range blockJSONget {
		r.Get(url, p.blockHandler(response))
	}
	for url, response := range blockJSONpost {
		r.Post(url, p.blockHandler(response))
	}
	r.Get("/", p.RootHandler)

	r.Get("/api/live/ws", p.wsHandler)
	fmt.Printf("Listening on http://localhost:8080\n")
	return http.ListenAndServe(":8080", r)
}

func (p *ProxyServer) blockHandler(response string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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
	resources, err := p.parser.Parse()
	if err != nil {
		log.Error("Error: ", err)
		http.Error(w, fmt.Sprintf("Error: %s", err), 500)
		return
	}
	w.Write([]byte("<h1>Available dashboards</h1><ul>"))
	for _, resource := range resources {
		if resource.Kind() == "Dashboard" {
			w.Write([]byte(fmt.Sprintf(`<li><a href="%s/d/%s/slug">%s</a></li>`, "http://localhost:8080", resource.Name(), resource.Name())))
		}
	}
	w.Write([]byte("</ul>"))
}

/*
* Note, this method avoids using `proxy.web`, implementing its own proxy
* event using Axios. This is because Grafana returns `X-Frame-Options: deny`
* which breaks our ability to place Grafana inside an iframe. `http-proxy`
* will not remove that header once it is added. Therefore we need a different
* form of proxy.
*
* This security protection does not apply to this situation - given we own
* both the connection to the backend as well as the webview. Therefore
* it is reasonable remove this header in this context.
*
* This method also doubles as connection verification. If an issue is
* encountered connecting to Grafana, rather than reporting an HTTP error,
* it returns an alternate HTML page to the user explaining the error, and
* offering a "refresh" option.
 */

func (p *ProxyServer) RootDashboardPageHandler(w http.ResponseWriter, r *http.Request) {
	req, err := http.NewRequest("GET", p.url+r.URL.Path, nil)
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(500), 500)
		return
	}
	req.Header.Set("Authorization", "Bearer "+p.token)
	req.Header.Set("User-Agent", p.userAgent)

	client := &http.Client{}
	resp, err := client.Do(req)

	if err == nil {
		w.Header().Add("Content-Type", "text/html")
		body, _ := io.ReadAll(resp.Body)
		w.Write(body)
		return
	}

	msg := ""
	if p.url == "" {
		msg += "<p><b>Error:</b> URL is not defined</p>"
	}
	if p.token == "" {
		msg += "<p><b>Warning:</b> No service account token specified.</p>"
	}

	if resp.StatusCode == 302 {
		http.Error(w, msg+"<p>Authentication error</p>", http.StatusUnauthorized)
	} else {
		body, _ := io.ReadAll(resp.Body)
		http.Error(w, msg+string(body), resp.StatusCode)
	}
}

func (p *ProxyServer) DashboardJSONGetHandler(w http.ResponseWriter, r *http.Request) {
	uid := chi.URLParam(r, "uid")
	if uid == "" {
		http.Error(w, "No UID specified", 400)
		return
	}

	// CORS!  cors(corsOptions),

	resources, err := p.parser.Parse()
	if err != nil {
		log.Error("Error: ", err)
		http.Error(w, fmt.Sprintf("Error: %s", err), 500)
		return
	}
	for _, resource := range resources {
		if resource.Kind() == "Dashboard" && resource.Name() == uid {
			meta := map[string]interface{}{
				"type":      "db",
				"isStarred": false,
				"folderID":  0,
				"folderUID": "",
				"url":       fmt.Sprintf("/d/%s/slug", uid),
			}
			wrapper := map[string]interface{}{
				"dashboard": resource.Spec(),
				"meta":      meta,
			}

			out, _ := json.Marshal(wrapper)
			w.Write(out)
			return
		}
	}
	http.Error(w, fmt.Sprintf("Dashboard with UID %s not found", uid), 404)
}

func (p *ProxyServer) DashboardJSONPostHandler(w http.ResponseWriter, r *http.Request) {
	//cors(corsOptions),

	if !p.isLegacy {
		http.Error(w, "Save only works for legacy json dashboards", 400)
		return
	}

	dash := map[string]interface{}{}
	content, _ := io.ReadAll(r.Body)
	err := json.Unmarshal(content, &dash)
	if err != nil {
		http.Error(w, "Error parsing JSON", 400)
		return
	}
	content, _ = json.MarshalIndent(dash["dashboard"], "  ", "  ")
	err = os.WriteFile(p.resourcePath, content, 0644)
	if err != nil {
		log.Print(p.resourcePath, p.isLegacy)
		http.Error(w, fmt.Sprintf("Error writing file: %s", err), 400)
	}

	log.Print(string(content))
	uid := dash["dashboard"].(map[string]interface{})["uid"].(string)
	jout := map[string]interface{}{
		"id":      1,
		"slug":    "slug",
		"status":  "success",
		"uid":     uid,
		"url":     fmt.Sprintf("/d/%s/slug", uid),
		"version": 1,
	}
	out, _ := json.Marshal(jout)
	w.Write(out)
}

/****** CORS
  const corsOptions = {
    origin: `http://localhost:${port}`,
    optionsSuccessStatus: 200,
  };
  ***/

/*** WEBSOCKETS
  server.on("upgrade", function (req, socket, head) {
    proxy.ws(req, socket, head, {});
  });
  ****/

/**** PROXY PAGES
  const sendErrorPage = (res: express.Response, message: string) => {
    const errorFile = path.join(extensionPath, "public/error.html");
    let content = fs.readFileSync(errorFile, "utf-8");
    content = content.replaceAll("${error}", message);
    res.write(content);
  };
  **/
