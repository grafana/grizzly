package grizzly

import (
	"fmt"
	"io"

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
	ResourceNotChanged = EventType{ID: "resource-not-changed", Severity: Info, HumanReadable: "no differences"}
	ResourceNotFound   = EventType{ID: "resource-not-found", Severity: Info, HumanReadable: "not found"}
	ResourceUpdated    = EventType{ID: "resource-updated", Severity: Notice, HumanReadable: "updated"}
	ResourcePulled     = EventType{ID: "resource-pulled", Severity: Notice, HumanReadable: "pulled"}
	ResourceFailure    = EventType{ID: "resource-failure", Severity: Error, HumanReadable: "failure"}
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
	EventCounts map[string]int
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
			EventCounts: make(map[string]int),
		},
	}
}

func (recorder *WriterRecorder) Record(event Event) {
	recorder.summary.EventCounts[event.Type.ID] += 1

	_, _ = recorder.out.Write([]byte(recorder.eventFormatter(event)))
}

func (recorder *WriterRecorder) Summary() Summary {
	return *recorder.summary
}
