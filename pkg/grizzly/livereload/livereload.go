package livereload

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

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
	//go RegularRefresh(10)
}

func RegularRefresh(t int) {
	for {
		time.Sleep(time.Duration(t) * time.Second)
		Reload("/d/qEYZMimVz/slug")
	}
}
func Reload(path string) {
	// Tell livereload a file has changed - will force a hard refresh if not CSS or an image
	log.Println("\n\nRELOAD.....\n\n")
	msg := fmt.Sprintf(`{"command":"reload","path":"%s","originalPath":"","liveCSS":true,"liveImg":true}`, path)
	wsHub.broadcast <- []byte(msg)
}

// This is a patched version, see https://github.com/livereload/livereload-js/pull/84
//
//go:embed livereload.js
var livereloadJS []byte

func LiveReloadJSHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/javascript")
	w.Write(livereloadJS)
}

func Inject(html []byte, port int) []byte {
	fmtstr := `<script src="/livereload.js?mindelay=10&amp;v=2&amp;port=%d&amp;path=livereload" data-no-instant="" defer=""></script>`
	inject := fmt.Sprintf(fmtstr, port)
	return []byte(strings.ReplaceAll(string(html), "<head>", "<head>\n"+inject))
}
