package grizzly

import (
	"fmt"
	"net/http"

	"log"
)

// Pluraliser returns a string describing the count of items, with a plural 's'
// appended if the count of items is greater than one.
func Pluraliser(count int, name string) string {
	if count == 1 {
		return fmt.Sprintf("1 %s", name)
	}

	return fmt.Sprintf("%d %ss", count, name)
}

func SendError(w http.ResponseWriter, msg string, err error, code int) {
	http.Error(w, msg, 400)
	log.Printf("%d - %s: %s", code, msg, err.Error())
}
