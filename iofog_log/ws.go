package iofog_log

import (
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
)

type LoggingWsHandler struct {
	dbManager   *DBManager
	config      *LoggingConfig
	configMutex sync.RWMutex
	upgrader    websocket.Upgrader
}

func newWsHandler(dbManager *DBManager) *LoggingWsHandler {
	handler := new(LoggingWsHandler)
	handler.dbManager = dbManager
	return handler
}

func (l *LoggingWsHandler) setConfig(c *LoggingConfig) {
	l.configMutex.Lock()
	l.config = c
	l.configMutex.Unlock()
}

func (l *LoggingWsHandler) HandleWsConnection(w http.ResponseWriter, r *http.Request) {
	var config *LoggingConfig
	l.configMutex.RLock()
	config = l.config
	l.configMutex.RUnlock()
	if isAuthorized(config.AccessTokens, w, r) {
		c, err := l.upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Printf("Error in upgrade handshake: %s", err.Error())
			return
		}
		defer c.Close()
		req := new(AddLogRequest)
		for {
			err := c.ReadJSON(req)
			if err != nil {
				logger.Printf("Error while reading JSON: %s", err.Error())
				break
			}
			var response []byte
			if _, err := l.dbManager.insert(&req.LogMessage); err != nil {
				response = []byte(err.Error())
			} else {
				response = []byte{CODE_ACK}
			}

			err = c.WriteMessage(websocket.BinaryMessage, response)
			if err != nil {
				logger.Printf("Error while writing message: %s", err.Error())
				break
			}
		}

	}
}
