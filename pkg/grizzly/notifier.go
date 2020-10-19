package grizzly

import (
	"fmt"

	"github.com/fatih/color"
)

var (
	red    = color.New(color.FgRed).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
)

// Notifier provides Handlers terminal agnostic mechanisms to announce results of actions
type Notifier struct{}

// NoChanges announces that nothing has changed
func (n *Notifier) NoChanges(resource Resource) {
	fmt.Printf("%s/%s %s\n", resource.JSONPath, resource.UID, yellow("no differences"))
}

// HasChanges announces that a resource has changed, and displays the differences
func (n *Notifier) HasChanges(resource Resource, diff string) {
	fmt.Printf("%s/%s %s\n", resource.JSONPath, resource.UID, red("changes detected:"))
	fmt.Println(diff)
}

// NotFound announces that a resource was not found on the remote endpoint
func (n *Notifier) NotFound(resource Resource) {
	fmt.Printf("%s/%s %s\n", resource.JSONPath, resource.UID, yellow("not present in "+resource.Handler.GetName()))
}

// Added announces that a resource has been added to the remote endpoint
func (n *Notifier) Added(resource Resource) {
	fmt.Printf("%s/%s %s\n", resource.JSONPath, resource.UID, green("added"))
}

// Updated announces that a resource has been updated at the remote endpoint
func (n *Notifier) Updated(resource Resource) {
	fmt.Printf("%s/%s %s\n", resource.JSONPath, resource.UID, green("updated"))
}

// NotSupported announces that a behaviour is not supported by a handler
func (n *Notifier) NotSupported(resource Resource, behaviour string) {
	fmt.Printf("%s/%s %s provider %s\n", resource.JSONPath, resource.UID, resource.Handler.GetName(), red("does not support "+behaviour))
}

// Info announces a message in green
func (n *Notifier) Info(resource Resource, msg string) {
	fmt.Printf("%s/%s %s\n", resource.JSONPath, resource.UID, green(msg))
}

// Warn announces a message in yellow
func (n *Notifier) Warn(resource Resource, msg string) {
	fmt.Printf("%s/%s %s\n", resource.JSONPath, resource.UID, yellow(msg))
}

// Error announces a message in yellow
func (n *Notifier) Error(resource Resource, msg string) {
	fmt.Printf("%s/%s %s\n", resource.JSONPath, resource.UID, red(msg))
}
