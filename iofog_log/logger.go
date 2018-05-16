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

package iofog_log

import (
	"encoding/json"
	"sync"
)

type LoggerSystem struct {
	dbManager   *DBManager
	config      *LoggingConfig
	configMutex sync.RWMutex
	RestHandler *LoggingRestHandler
	WsHandler   *LoggingWsHandler
}

func New() (*LoggerSystem, error) {
	dbManager, err := newDBManager()
	if err != nil {
		return nil, err
	}
	loggerSystem := new(LoggerSystem)
	loggerSystem.dbManager = dbManager
	loggerSystem.RestHandler = newRestHandler(dbManager)
	loggerSystem.WsHandler = newWsHandler(dbManager)
	return loggerSystem, nil

}

func (l *LoggerSystem) UpdateConfig(config map[string]interface{}) {
	configBytes, err := json.Marshal(config)
	if err != nil {
		logger.Println(err.Error())
		return
	}
	l.config = new(LoggingConfig)
	err = json.Unmarshal(configBytes, l.config)
	if err != nil {
		logger.Println(err.Error())
		return
	}

	l.RestHandler.setConfig(l.config)
	l.WsHandler.setConfig(l.config)
	cf := parseDuration(l.config.CleanFrequency, "cleaning frequency", DEFAULT_CLEAN_FREQUENCY)
	ttl := parseDuration(l.config.TTL, "ttl", DEFAULT_TTL)

	l.dbManager.setCleanInterval(cf, ttl)
}
