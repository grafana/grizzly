package livereload

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type connection struct {
	ws       *websocket.Conn
	send     chan []byte
	closer   sync.Once
	clientID string
}

func NewConnection(send chan []byte, ws *websocket.Conn) *connection {
	return &connection{
		send: send,
		ws:   ws,
	}
}

func (c *connection) close() {
	c.closer.Do(func() {
		close(c.send)
	})
}

type connectRequest struct {
	ID int `json:"id"`
}

type connectResponseInner struct {
	Client string `json:"client"`
	Ping   int    `json:"ping"`
	Pong   bool   `json:"pong"`
}
type connectResponse struct {
	ID      int                  `json:"id"`
	Connect connectResponseInner `json:"connect"`
}

func (c *connection) handleConnectRequest(line string) ([]byte, error) {
	// {"connect":{"name":"js"},"id":1}
	request := connectRequest{}
	err := json.Unmarshal([]byte(line), &request)
	if err != nil {
		return nil, err
	}
	c.clientID = uuid.New().String()
	// {"id":1,"connect":{"client":"5a6674c9-2450-46e4-bfff-beaa84966493","ping":25,"pong":true}}
	response := connectResponse{
		ID: request.ID,
		Connect: connectResponseInner{
			Client: c.clientID,
			Ping:   25,
			Pong:   true,
		},
	}
	j, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	return j, nil
}

type subscribeRequestInner struct {
	Channel string `json:"channel"`
}
type subscribeRequest struct {
	ID        int                   `json:"id"`
	Subscribe subscribeRequestInner `json:"subscribe"`
}

type joinResponseInfo struct {
	User   string `json:"user"`
	Client string `json:"client"`
}
type joinResponseJoin struct {
	Info joinResponseInfo `json:"info"`
}
type joinResponsePush struct {
	Channel string           `json:"channel"`
	Join    joinResponseJoin `json:"join"`
}
type joinResponse struct {
	Push joinResponsePush `json:"push"`
}

func (c *connection) handleSubscribeRequest(line string) ([]byte, error) {
	// {"subscribe":{"channel":"1/grafana/dashboard/uid/no-folder"},"id":2}
	request := subscribeRequest{}
	err := json.Unmarshal([]byte(line), &request)
	if err != nil {
		return nil, err
	}
	// {"id":2,"subscribe":{}}
	subResp := map[string]any{}
	subResp["id"] = request.ID
	subResp["subscribe"] = map[string]any{}
	j, err := json.Marshal(subResp)
	if err != nil {
		return nil, err
	}
	c.send <- j
	// {"push":{"channel":"1/grafana/dashboard/uid/no-folder","join":{"info":{"user":"1","client":"5a6674c9-2450-46e4-bfff-beaa84966493"}}}}
	joinResp := joinResponse{
		Push: joinResponsePush{
			Channel: request.Subscribe.Channel,
			Join: joinResponseJoin{
				Info: joinResponseInfo{
					User:   "1",
					Client: c.clientID,
				},
			},
		},
	}
	j, err = json.Marshal(joinResp)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (c *connection) reader() {
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			break
		}
		lines := strings.Split(string(message), "\n")
		for _, line := range lines {
			msg := map[string]any{}
			err := json.Unmarshal([]byte(line), &msg)
			if err != nil {
				log.Printf("Error parsing websocket message: %v", err)
				continue
			}
			if _, ok := msg["connect"]; ok {
				j, err := c.handleConnectRequest(line)
				if err != nil {
					log.Printf("Error handling connection request: %s", err)
					continue
				}
				c.send <- j
			} else if _, ok := msg["subscribe"]; ok {
				j, err := c.handleSubscribeRequest(line)
				if err != nil {
					log.Printf("Error handling subscribe request: %s", err)
					continue
				}
				c.send <- j
			}
		}
	}
	c.ws.Close()
}

type pushResponseUser struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
}
type pushResponseDashboard struct {
	UID      string         `json:"uid"`
	FolderID int            `json:"folderID"`
	IsFolder bool           `json:"IsFolder"`
	Data     map[string]any `json:"data"`
}

type pushResponseData struct {
	UID       string                `json:"uid"`
	Action    string                `json:"action"`
	User      pushResponseUser      `json:"user"`
	Dashboard pushResponseDashboard `json:"dashboard"`
}
type pushResponsePub struct {
	Data pushResponseData `json:"data"`
}

type pushResponsePush struct {
	Channel string          `json:"channel"`
	Pub     pushResponsePub `json:"pub"`
}

type pushResponse struct {
	Push pushResponsePush `json:"push"`
}

func (c *connection) NotifyDashboard(uid string, spec map[string]any) error {
	response := pushResponse{
		Push: pushResponsePush{
			Channel: fmt.Sprintf("1/grafana/dashboard/uid/%s", uid),
			Pub: pushResponsePub{
				Data: pushResponseData{
					UID:    uid,
					Action: "saved",
					User: pushResponseUser{
						ID:    1,
						Login: "admin",
					},
					Dashboard: pushResponseDashboard{
						UID:      uid,
						FolderID: 0,
						IsFolder: false,
						Data:     spec,
					},
				},
			},
		},
	}
	j, err := json.Marshal(response)
	if err != nil {
		return err
	}
	c.send <- j
	return nil
}

func (c *connection) writer() {
	for message := range c.send {
		err := c.ws.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			break
		}
	}
	c.ws.Close()
}
