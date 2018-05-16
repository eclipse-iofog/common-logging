/*
 *******************************************************************************
 * Copyright (c) 2018 Edgeworx, Inc.
 *
 * This program and the accompanying materials are made available under the
 * terms of the Eclipse Public License v. 2.0 which is available at
 * http://www.eclipse.org/legal/epl-2.0
 *
 * SPDX-License-Identifier: EPL-2.0
 *******************************************************************************
*/

package main

import (
	"fmt"
	sdk "github.com/ioFog/iofog-go-sdk"
	"github.com/ioFog/common-logging/iofog_log"
	"log"
	"net/http"
)

var (
	client, err = sdk.NewDefaultIoFogClient()
)

func main() {
	if err != nil {
		log.Fatal(err.Error())
	}
	logSystem, err := iofog_log.New()
	if err != nil {
		log.Fatal(err.Error())
	}

	if config, err := client.GetConfig(); err != nil {
		log.Fatal("Unable to retrieve initial config", err.Error())
	} else {
		logSystem.UpdateConfig(config)
	}

	signal := client.EstablishControlWsConnection(0)
	go func() {
		for {
			select {
			case <-signal:
				config, err := client.GetConfig()
				if err != nil {
					log.Printf("Error while updating config: %s", err.Error())
				}
				logSystem.UpdateConfig(config)
			}
		}
	}()

	http.HandleFunc("/logs/get", logSystem.RestHandler.HandleGetLogs)
	http.HandleFunc("/logs/add", logSystem.RestHandler.HandlePostLog)
	http.HandleFunc("/logs/ws", logSystem.WsHandler.HandleWsConnection)
	http.ListenAndServe(fmt.Sprintf(":%d", iofog_log.LOGGER_CONTAINER_PORT), nil)
}
