// Copyright 2015 The Hugo Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package livereload

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

// Initialize starts the Websocket Hub handling live reloads.
func Initialize() {
	go wsHub.run()
}

// Handler is a HandlerFunc handling the livereload
// Websocket interaction.
func Handler(upgrader *websocket.Upgrader) http.HandlerFunc {
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

func ReloadDashboard(uid string) {
	msg := fmt.Sprintf(`{"command": "reload", "path": "/grizzly/Dashboard/%s"}`, uid)
	wsHub.broadcast <- []byte(msg)
}
