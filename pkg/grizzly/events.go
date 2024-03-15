package grizzly

import (
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
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
