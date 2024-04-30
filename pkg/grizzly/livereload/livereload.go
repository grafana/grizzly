package livereload

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"

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

func Reload(kind, name string, spec map[string]any) error {
	log.Printf("Reloading %s/%s", kind, name)
	if kind != "Dashboard" {
		return fmt.Errorf("only dashboards supported for live reload at present")
	}
	wsHub.NotifyDashboard(name, spec)
	return nil
}
