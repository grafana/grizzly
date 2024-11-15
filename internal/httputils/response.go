package httputils

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func Error(w http.ResponseWriter, msg string, err error, code int) {
	log.Warnf("%d - %s: %s", code, msg, err.Error())
	http.Error(w, msg, code)
}

func Write(w http.ResponseWriter, content []byte) {
	if _, err := w.Write(content); err != nil {
		log.Errorf("error writing response: %v", err)
	}
}

func WriteJSON(w http.ResponseWriter, content any) {
	responseJSON, err := json.Marshal(content)
	if err != nil {
		log.Errorf("error marshalling response to JSON: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	Write(w, responseJSON)
}
