package grizzly

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/grafana/grizzly/pkg/config"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type EventSeverity uint8

const (
	Info EventSeverity = iota
	Notice
	Error
)

type EventType struct {
	Severity      EventSeverity
	ID            string
	HumanReadable string
}

var (
	ResourceAdded      = EventType{ID: "resource-added", Severity: Notice, HumanReadable: "added"}
	ResourceNotChanged = EventType{ID: "resource-not-changed", Severity: Info, HumanReadable: "unchanged"}
	ResourceNotFound   = EventType{ID: "resource-not-found", Severity: Info, HumanReadable: "not found"}
	ResourceUpdated    = EventType{ID: "resource-updated", Severity: Notice, HumanReadable: "updated"}
	ResourcePulled     = EventType{ID: "resource-pulled", Severity: Notice, HumanReadable: "pulled"}
	ResourceFailure    = EventType{ID: "resource-failure", Severity: Error, HumanReadable: "failed"}
)

type Event struct {
	Type        EventType
	ResourceRef string
	Details     string
}

type EventFormatter func(event Event) string

func EventToPlainText(event Event) string {
	if event.Details == "" {
		return fmt.Sprintf("%s %s\n", event.ResourceRef, event.Type.HumanReadable)
	}

	return fmt.Sprintf("%s %s: %s\n", event.ResourceRef, event.Type.HumanReadable, event.Details)
}

func EventToColoredText(event Event) string {
	var colorFunc func(...interface{}) string

	switch event.Type.Severity {
	case Info:
		colorFunc = color.New(color.FgYellow).SprintFunc()
	case Notice:
		colorFunc = color.New(color.FgGreen).SprintFunc()
	case Error:
		colorFunc = color.New(color.FgRed).SprintFunc()
	}

	eventType := event.Type.HumanReadable
	if colorFunc != nil {
		eventType = colorFunc(eventType)
	}

	if event.Details == "" {
		return fmt.Sprintf("%s %s\n", event.ResourceRef, eventType)
	}

	return fmt.Sprintf("%s %s: %s\n", event.ResourceRef, eventType, event.Details)
}

type Summary struct {
	EventCounts map[EventType]int
}

func (summary Summary) AsString(resourceLabel string) string {
	var parts []string

	for eventType, count := range summary.EventCounts {
		if count == 0 {
			continue
		}

		parts = append(parts, fmt.Sprintf("%s %s", Pluraliser(count, resourceLabel), eventType.HumanReadable))
	}

	return strings.Join(parts, ", ")
}

type WriterRecorder struct {
	out            io.Writer
	eventFormatter EventFormatter
	summary        *Summary
}

func NewWriterRecorder(out io.Writer, eventFormatter EventFormatter) *WriterRecorder {
	return &WriterRecorder{
		out:            out,
		eventFormatter: eventFormatter,
		summary: &Summary{
			EventCounts: make(map[EventType]int),
		},
	}
}

func (recorder *WriterRecorder) Record(event Event) {
	recorder.summary.EventCounts[event.Type] += 1

	_, _ = recorder.out.Write([]byte(recorder.eventFormatter(event)))
}

func (recorder *WriterRecorder) Summary() Summary {
	return *recorder.summary
}

var _ EventsRecorder = (*WriterRecorder)(nil)

type UsageRecorder struct {
	wr       *WriterRecorder
	endpoint string
}

// Record implements EventsRecorder.
func (u *UsageRecorder) Record(event Event) {
	u.wr.Record(event)
}

// Summary implements EventsRecorder.
func (u *UsageRecorder) Summary() Summary {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	group, ctx := errgroup.WithContext(ctx)
	for op, n := range u.wr.summary.EventCounts {
		group.Go(func() error {
			return u.reportUsage(ctx, op.ID, n)
		})
	}
	err := group.Wait()
	if err != nil {
		log.Debugf("failed to send usage stats: %v", err)
	}

	return u.wr.Summary()
}

func (u *UsageRecorder) reportUsage(ctx context.Context, op string, n int) error {
	var buff bytes.Buffer
	configHash, err := config.Hash()
	if err != nil {
		configHash = "failed-to-hash-config"
	}
	err = json.NewEncoder(&buff).Encode(map[string]interface{}{
		"uuid":      configHash,
		"arch":      runtime.GOARCH,
		"os":        runtime.GOOS,
		"resources": n,
		"operation": op,
		"createdAt": time.Now(),
		"version":   config.Version,
	})
	if err != nil {
		return fmt.Errorf("encoding usage report")
	}
	req, err := http.NewRequest(http.MethodPost, u.endpoint, &buff)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	req.Header.Add("content-type", "application/json")
	req = req.WithContext(ctx)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending post request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected OK, got %s", resp.Status)
	}
	return nil
}

var _ EventsRecorder = (*UsageRecorder)(nil)

func NewUsageRecorder(wr *WriterRecorder) *UsageRecorder {
	return &UsageRecorder{
		wr:       wr,
		endpoint: "https://stats.grafana.org/grizzly-usage-report",
	}
}
