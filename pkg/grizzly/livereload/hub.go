package livereload

import "log"

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

func (h *hub) NotifyDashboard(uid string, spec map[string]any) {
	for c := range h.connections {
		err := c.NotifyDashboard(uid, spec)
		if err != nil {
			log.Printf("Error notifying %s: %s", c.clientID, err)
		}
	}
}

func (h *hub) run() {
	for {
		select {
		case c := <-h.register:
			h.connections[c] = true
		case c := <-h.unregister:
			delete(h.connections, c)
			c.close()
		}
	}
}
