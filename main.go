package main

import (
	"github.com/iotracks/logging-system-container/iofog_log"
	sdk "github.com/iotracks/container-sdk-go"
	"net/http"
	"fmt"
	"log"
)

var (
	client, err = sdk.NewDefaultIoFogClient()
)

func main() {
	if err != nil {
		log.Fatal(err.Error())
	}
	handler, err := iofog_log.NewRestHandler()
	if err != nil {
		log.Fatal(err.Error())
	}

	if config, err := client.GetConfig(); err != nil {
		log.Fatal("Unable to retrieve initial config", err.Error())
	} else {
		handler.UpdateConfig(config)
	}

	signal := client.EstablishControlWsConnection(0)
	go func() {
		select {
		case <-signal:
			config, err := client.GetConfig()
			if err != nil {
				log.Printf("Error while updating config: %s", err.Error())
			}
			handler.UpdateConfig(config)
		}
	}()

	http.HandleFunc("/logs/get", handler.HandleGetLogs)
	http.HandleFunc("/logs/add", handler.HandlePostLog)
	http.ListenAndServe(fmt.Sprintf(":%d", iofog_log.LOGGER_CONTAINER_PORT), nil)
}