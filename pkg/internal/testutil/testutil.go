package testutil

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func GetUrl() string {
	if os.Getenv("CI") != "" {
		return "http://grizzly-grafana:3000/"
	} else {
		return "http://localhost:3001/"
	}
}

func PingService(url string) *time.Ticker {
	ticker := time.NewTicker(1 * time.Second)
	timeoutExceeded := time.After(120 * time.Second)

	success := false
	for !success {
		select {
		case <-timeoutExceeded:
			panic(fmt.Sprintf("Unable to connect to %s", url))

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
