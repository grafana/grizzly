package notifier

import (
	"fmt"

	"github.com/fatih/color"
)

var (
	red    = color.New(color.FgRed).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
)

// NoChanges announces that nothing has changed
func NoChanges(obj fmt.Stringer) {
	fmt.Printf("%s %s\n", obj.String(), yellow("no differences"))
}

// HasChanges announces that a resource has changed, and displays the differences
func HasChanges(obj fmt.Stringer, diff string) {
	fmt.Printf("%s %s\n", obj.String(), red("changes detected:"))
	fmt.Println(diff)
}

// NotFound announces that a resource was not found on the remote endpoint
func NotFound(obj fmt.Stringer) {
	fmt.Printf("%s %s\n", obj.String(), yellow("not found"))
}

// Added announces that a resource has been added to the remote endpoint
func Added(obj fmt.Stringer) {
	fmt.Printf("%s %s\n", obj.String(), green("added"))
}

// Updated announces that a resource has been updated at the remote endpoint
func Updated(obj fmt.Stringer) {
	fmt.Printf("%s %s\n", obj.String(), green("updated"))
}

// NotSupported announces that a behaviour is not supported by a handler
func NotSupported(obj fmt.Stringer, behaviour string) {
	fmt.Printf("%s %s\n", obj.String(), red("does not support "+behaviour))
}

// Info announces a message in green
func Info(obj fmt.Stringer, msg string) {
	if obj == nil {
		fmt.Println(green(msg))
	} else {
		fmt.Printf("%s %s\n", obj.String(), green(msg))
	}
}

// Warn announces a message in yellow
func Warn(obj fmt.Stringer, msg string) {
	if obj == nil {
		fmt.Println(yellow(msg))
	} else {
		fmt.Printf("%s %s\n", obj.String(), yellow(msg))
	}
}

// Error announces a message in yellow
func Error(obj fmt.Stringer, msg string) {
	if obj == nil {
		fmt.Println(red(msg))
	} else {
		fmt.Printf("%s %s\n", obj.String(), red(msg))
	}
}

type SimpleString string

func (s SimpleString) String() string {
	return string(s)
}
