package iofog_log

import (
	"sync"
	"encoding/json"
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