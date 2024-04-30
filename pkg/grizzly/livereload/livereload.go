package livereload

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type hub struct {
	connections map[*connection]bool
	register    chan *connection
	unregister  chan *connection
}

var wsHub = hub{
	register:    make(chan *connection),
	unregister:  make(chan *connection),
	connections: make(map[*connection]bool),
}

func Register(c *connection) {
	wsHub.register <- c
}

func Unregister(c *connection) {
	wsHub.unregister <- c
}

func Initialize() {
	go func() {
		for {
			select {
			case c := <-wsHub.register:
				wsHub.connections[c] = true
			case c := <-wsHub.unregister:
				delete(wsHub.connections, c)
				c.close()
			}
		}
	}()
}

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

func Reload(kind, name string, spec map[string]any) error {
	log.Printf("Reloading %s/%s", kind, name)
	if kind != "Dashboard" {
		return fmt.Errorf("only dashboards supported for live reload at present")
	}
	for c := range wsHub.connections {
		err := c.NotifyDashboard(name, spec)
		if err != nil {
			log.Printf("Error notifying %s: %s", c.clientID, err)
		}
	}
	return nil
}
