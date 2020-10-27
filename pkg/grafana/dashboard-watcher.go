package grafana

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/centrifugal/centrifuge-go"
	"github.com/grafana/grizzly/pkg/grizzly"
)

type eventHandler struct {
	filename string
	url      string
	stop     bool
}

func (h *eventHandler) OnConnect(c *centrifuge.Client, e centrifuge.ConnectEvent) {
	log.Println("Connected to", h.url)
	return
}

func (h *eventHandler) OnError(c *centrifuge.Client, e centrifuge.ErrorEvent) {
	log.Printf("Error: %s", e.Message)
	return
}

func (h *eventHandler) OnDisconnect(c *centrifuge.Client, e centrifuge.DisconnectEvent) {
	log.Println("Disconnected from", h.url)
	h.stop = true
	return
}
func (h *eventHandler) OnSubscribeSuccess(sub *centrifuge.Subscription, e centrifuge.SubscribeSuccessEvent) {
	log.Printf("Subscribed to channel %s, resubscribed: %v, recovered: %v", sub.Channel(), e.Resubscribed, e.Recovered)
}

func (h *eventHandler) OnSubscribeError(sub *centrifuge.Subscription, e centrifuge.SubscribeErrorEvent) {
	log.Printf("Failed to subscribe to channel %s, error: %s", sub.Channel(), e.Error)
}

func (h *eventHandler) OnUnsubscribe(sub *centrifuge.Subscription, e centrifuge.UnsubscribeEvent) {
	log.Printf("Unsubscribed from channel %s", sub.Channel())
}

func (h *eventHandler) OnPublish(sub *centrifuge.Subscription, e centrifuge.PublishEvent) {
	response := struct {
		UID    string `json:"uid"`
		Action string `json:"action"`
		UserID int64  `json:"userId"`
	}{}
	err := json.Unmarshal(e.Data, &response)
	if err != nil {
		log.Println(err)
		return
	}
	if response.Action != "saved" {
		log.Println("Unknown action received", string(e.Data))
	}
	dashboard, err := getRemoteDashboard(response.UID)
	if err != nil {
		log.Println(err)
		return
	}
	dashboardJSON, err := dashboard.toJSON()
	if err != nil {
		log.Println(err)
		return
	}
	ioutil.WriteFile(h.filename, []byte(dashboardJSON), 0644)
	log.Printf("%s updated from dashboard %s", h.filename, response.UID)
}

func (h *eventHandler) WaitForStop() {
	for {
		time.Sleep(time.Second)
		if h.stop {
			log.Println("Stopping.")
			os.Exit(1)
		}
	}
}
func watchDashboard(notifier grizzly.Notifier, UID, filename string) error {
	wsURL, token, err := getWSGrafanaURL("live/ws?format=json")
	if err != nil {
		return err
	}

	c := centrifuge.New(wsURL, centrifuge.DefaultConfig())
	handler := &eventHandler{
		filename: filename,
		url:      wsURL,
	}
	c.OnConnect(handler)
	c.OnError(handler)
	c.OnDisconnect(handler)
	c.SetToken(token)

	channel := fmt.Sprintf("grafana/dashboard/%s", UID)
	sub, err := c.NewSubscription(channel)
	if err != nil {
		return err
	}

	sub.OnSubscribeSuccess(handler)
	sub.OnSubscribeError(handler)
	sub.OnUnsubscribe(handler)
	sub.OnPublish(handler)

	err = sub.Subscribe()
	if err != nil {
		return err
	}

	err = c.Connect()
	if err != nil {
		return err
	}

	go handler.WaitForStop()
	// Run until CTRL+C.
	select {}
}
