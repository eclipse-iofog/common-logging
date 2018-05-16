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
	"net/http"
	"sync"
)

type LoggingRestHandler struct {
	dbManager   *DBManager
	config      *LoggingConfig
	configMutex sync.RWMutex
}

func newRestHandler(dbManager *DBManager) *LoggingRestHandler {
	handler := new(LoggingRestHandler)
	handler.dbManager = dbManager
	return handler
}

func (l *LoggingRestHandler) HandleGetLogs(w http.ResponseWriter, r *http.Request) {
	if checkMethodAndContent(http.MethodPost, APPLICATION_JSON, w, r) {
		w.Header().Set(CONTENT_TYPE, APPLICATION_JSON)
		var req GetLogsRequest
		var config *LoggingConfig
		l.configMutex.RLock()
		config = l.config
		l.configMutex.RUnlock()
		if isAuthorized(config.AccessTokens, w, r) && getJsonBody(w, r, &req) {
			if resp, err := l.dbManager.query(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			} else if err := json.NewEncoder(w).Encode(resp); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
		}
	}
}

func (l *LoggingRestHandler) HandlePostLog(w http.ResponseWriter, r *http.Request) {
	if checkMethodAndContent(http.MethodPost, APPLICATION_JSON, w, r) {
		w.Header().Set(CONTENT_TYPE, APPLICATION_JSON)
		var req AddLogRequest
		var config *LoggingConfig
		l.configMutex.RLock()
		config = l.config
		l.configMutex.RUnlock()
		if isAuthorized(config.AccessTokens, w, r) && getJsonBody(w, r, &req) {
			if _, err := l.dbManager.insert(&req.LogMessage); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
		}
	}
}

func (l *LoggingRestHandler) setConfig(c *LoggingConfig) {
	l.configMutex.Lock()
	l.config = c
	l.configMutex.Unlock()
}
