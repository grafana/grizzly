package livereload

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

// Handler is a HandlerFunc handling the livereload
// Websocket interaction.
func LiveReloadHandlerFunc(upgrader websocket.Upgrader) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		c := &connection{send: make(chan []byte, 256), ws: ws}
		wsHub.register <- c
		defer func() { wsHub.unregister <- c }()
		go c.writer()
		c.reader()
	}
}

// Initialize starts the Websocket Hub handling live reloads.
func Initialize() {
	go wsHub.run()
}

func Reload(path string) {
	log.Printf("Reloading %s", path)
	msg := fmt.Sprintf(`{"command":"reload","path":"%s","originalPath":"%s"}`, path, path)
	wsHub.broadcast <- []byte(msg)
}

// This is a patched version, see https://github.com/livereload/livereload-js/pull/84, cloned from github.com/gohugoio/hugo
//
//go:embed livereload.js
var livereloadJS []byte

func LiveReloadJSHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/javascript")
	w.Write(livereloadJS)
}

func Inject(html []byte, port int) []byte {
	inject := `<script src="/livereload.js?mindelay=10&amp;v=2&amp;path=livereload" data-no-instant="" defer=""></script>`
	//inject := fmt.Sprintf(fmtstr, port)
	return []byte(strings.ReplaceAll(string(html), "<head>", "<head>\n"+inject))
}
