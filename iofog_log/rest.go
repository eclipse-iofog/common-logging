package iofog_log

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
	"fmt"
)

type LoggingRestHandler struct {
	dbManager   *DBManager
	config      *LoggingConfig
	configMutex sync.RWMutex
}

func NewRestHandler() (*LoggingRestHandler, error) {
	dbManager, err := NewDBManager()
	if err != nil {
		return nil, err
	}
	handler := new(LoggingRestHandler)
	handler.dbManager = dbManager
	return handler, nil

}

func (l *LoggingRestHandler) HandleGetLogs(w http.ResponseWriter, r *http.Request) {
	if l.checkMethodAndContent(http.MethodPost, APPLICATION_JSON, w, r) {
		w.Header().Set(CONTENT_TYPE, APPLICATION_JSON)
		var req GetLogsRequest
		if l.getBody(w, r, &req) && l.isAuthorized(w, r) {
			if resp, err := l.dbManager.Query(&req); err != nil {
				l.writeError(w, err)
			} else if err := json.NewEncoder(w).Encode(resp); err != nil {
				l.writeError(w, err)
			}
		}
	}
}

func (l *LoggingRestHandler) HandlePostLog(w http.ResponseWriter, r *http.Request) {
	if l.checkMethodAndContent(http.MethodPost, APPLICATION_JSON, w, r) {
		w.Header().Set(CONTENT_TYPE, APPLICATION_JSON)
		var req AddLogRequest
		if l.getBody(w, r, &req) && l.isAuthorized(w, r) {
			if _, err := l.dbManager.Insert(&req.LogMessage); err != nil {
				l.writeError(w, err)
			}
		}
	}
}

func (l *LoggingRestHandler) UpdateConfig(config map[string]interface{}) {
	configBytes, err := json.Marshal(config)
	if err != nil {
		logger.Println(err.Error())
		return
	}
	c := new(LoggingConfig)
	err = json.Unmarshal(configBytes, c)
	if err != nil {
		logger.Println(err.Error())
		return
	}
	l.configMutex.Lock()
	l.config = c
	l.configMutex.Unlock()
	if d, e := time.ParseDuration(fmt.Sprintf("%ds", c.CleanInterval)); c.CleanInterval != 0 && e == nil {
		l.dbManager.SetCleanInterval(d)
	} else {
		l.dbManager.SetCleanInterval(DEFAULT_CLEAN_DURATION)
	}
}

func (l *LoggingRestHandler) isAuthorized(w http.ResponseWriter, r *http.Request) bool {
	var config *LoggingConfig
	l.configMutex.RLock()
	config = l.config
	l.configMutex.RUnlock()
	accessToken := r.Header.Get(ACCESS_TOKEN)
	for _, token := range config.AccessTokens {
		if accessToken == token {
			return true
		}
	}
	w.WriteHeader(http.StatusUnauthorized)
	return false
}
func (l *LoggingRestHandler) checkMethodAndContent(method, contentType string,
w http.ResponseWriter, r *http.Request) bool {
	if r.Method != method {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return false
	}
	if r.Header.Get(CONTENT_TYPE) != contentType {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return false
	}
	return true
}

func (l *LoggingRestHandler) getBody(w http.ResponseWriter, r *http.Request, decoded interface{}) bool {
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(decoded)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return false
	}
	return true
}

func (l *LoggingRestHandler) writeError(w http.ResponseWriter, err error) {
	w.Header().Set(CONTENT_TYPE, TEXT_PLAN)
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(err.Error()))
}
