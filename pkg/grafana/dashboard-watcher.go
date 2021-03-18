package grafana

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/centrifugal/centrifuge-go"
	"github.com/grafana/grizzly/pkg/grizzly"
)

type eventHandler struct {
	filename string
	url      string
	stop     bool
	notifier grizzly.Notifier
}

func (h *eventHandler) OnConnect(c *centrifuge.Client, e centrifuge.ConnectEvent) {
	h.notifier.Info(nil, fmt.Sprintf("Connected to %s", h.url))
}

func (h *eventHandler) OnError(c *centrifuge.Client, e centrifuge.ErrorEvent) {
	h.notifier.Error(nil, fmt.Sprintf("Error: %s", e.Message))
}

func (h *eventHandler) OnDisconnect(c *centrifuge.Client, e centrifuge.DisconnectEvent) {
	h.notifier.Error(nil, fmt.Sprintf("Disconnected from %s", h.url))
	h.stop = true
}
func (h *eventHandler) OnSubscribeSuccess(sub *centrifuge.Subscription, e centrifuge.SubscribeSuccessEvent) {
	h.notifier.Info(nil, fmt.Sprintf("Subscribed to channel %s", sub.Channel()))
}

func (h *eventHandler) OnSubscribeError(sub *centrifuge.Subscription, e centrifuge.SubscribeErrorEvent) {
	h.notifier.Error(nil, fmt.Sprintf("Failed to subscribe to channel %s, error: %s", sub.Channel(), e.Error))
}

func (h *eventHandler) OnUnsubscribe(sub *centrifuge.Subscription, e centrifuge.UnsubscribeEvent) {
	h.notifier.Info(nil, fmt.Sprintf("Unsubscribed from channel %s", sub.Channel()))
}

func (h *eventHandler) OnPublish(sub *centrifuge.Subscription, e centrifuge.PublishEvent) {
	response := struct {
		UID    string `json:"uid"`
		Action string `json:"action"`
		UserID int64  `json:"userId"`
	}{}
	err := json.Unmarshal(e.Data, &response)
	if err != nil {
		h.notifier.Error(nil, fmt.Sprintf("Error: %s", err))
		return
	}
	if response.Action != "saved" {
		h.notifier.Warn(nil, fmt.Sprintf("Unknown action received: %s", string(e.Data)))
	}
	dashboard, err := getRemoteDashboard(response.UID)
	if err != nil {
		h.notifier.Error(nil, fmt.Sprintf("Error: %s", err))
		return
	}
	dashboardJSON, err := dashboard.SpecAsJSON()
	if err != nil {
		h.notifier.Error(nil, fmt.Sprintf("Error: %s", err))
		return
	}
	ioutil.WriteFile(h.filename, []byte(dashboardJSON), 0644)
	t := time.Now()
	now := fmt.Sprintf(t.Format("2006-01-02-15:04:05"))
	h.notifier.Info(nil, fmt.Sprintf("%s: %s updated from dashboard %s", now, h.filename, response.UID))
}

func (h *eventHandler) WaitForStop() {
	for {
		time.Sleep(time.Second)
		if h.stop {
			h.notifier.Warn(nil, "Stopping.")
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
		notifier: notifier,
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
		msg := err.Error()
		if strings.Contains(msg, "bad handshake") {
			notifier.Error(nil, "Your Grafana Version does not support listening for dashboard changes.")
			notifier.Warn(nil, "The feature is available in Grafana 7.3 when the 'live' feature flag is enabled.")
			notifier.Warn(nil, "If running Grafana in Docker, set envvar GF_FEATURE_TOGGLES_ENABLE=live")
			notifier.Warn(nil, "Otherwise, add 'enable = live' to your [feature_toggles] section in grafana.ini.")
			return nil
		}
		return err
	}

	go handler.WaitForStop()
	// Run until CTRL+C.
	select {}
}
