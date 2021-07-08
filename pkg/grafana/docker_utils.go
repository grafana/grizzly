package grafana

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func getUrl() string {
	if os.Getenv("CI") != "" {
		return "http://grizzly-grafana:3000/"
	} else {
		return "http://localhost:3000/"
	}
}

func pingService(url string) *time.Ticker {
	ticker := time.NewTicker(1 * time.Second)
	timeoutExceeded := time.After(120 * time.Second)

	success := false
	for !success {
		select {
		case <-timeoutExceeded:
			panic("Unable to connect to grizzly-grafana:3000")

		case <-ticker.C:
			resp, _ := http.Get(url)
			fmt.Println("Response:", resp)
			if resp != nil {
				success = true
			}
		}
	}
	return ticker
}
