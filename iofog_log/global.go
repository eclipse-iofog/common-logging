package iofog_log

import (
	"fmt"
	"log"
	"os"
	"time"
)

const (
	DB_LOCATION              = "/log/"
	DB_NAME                  = "iofog.logs.db"
	TABLE_NAME               = "logs"
	ID_COLUMN_NAME           = "id"
	PUBLISHER_ID_COLUMN_NAME = "publisher"
	LOG_LEVEL_COLUMN_NAME    = "level"
	LOG_MESSAGE_COLUMN_NAME  = "message"
	TIMESTAMP_COLUMN_NAME    = "timestamp"
	APPLICATION_JSON         = "application/json"
	CONTENT_TYPE             = "Content-Type"
	DEFAULT_PAGE_SIZE        = 20
	DEFAULT_ORDER_BY         = TIMESTAMP_COLUMN_NAME
	DEFAULT_CLEAN_FREQUENCY  = time.Hour * 24 * 2
	DEFAULT_TTL              = time.Hour * 24 * 10

	ACCESS_TOKEN = "Access-Token"
	ASC          = "ASC"
	DESC         = "DESC"

	CRITICAL = 50
	FATAL    = CRITICAL
	ERROR    = 40
	WARNING  = 30
	WARN     = WARNING
	INFO     = 20
	DEBUG    = 10
	NOTSET   = 0

	LOGGER_CONTAINER_PORT = 10555

	CODE_ACK            = 0xB
)

var (
	logger      = log.New(os.Stderr, "", log.LstdFlags)
	_levelNames = map[interface{}]interface{}{
		CRITICAL:   "CRITICAL",
		ERROR:      "ERROR",
		WARNING:    "WARNING",
		INFO:       "INFO",
		DEBUG:      "DEBUG",
		NOTSET:     "NOTSET",
		"CRITICAL": CRITICAL,
		"FATAL":    FATAL,
		"ERROR":    ERROR,
		"WARN":     WARN,
		"WARNING":  WARNING,
		"INFO":     INFO,
		"DEBUG":    DEBUG,
		"NOTSET":   NOTSET,
		"":         NOTSET,
	}

	ORDER_BY_FIELDS = []string{PUBLISHER_ID_COLUMN_NAME, LOG_LEVEL_COLUMN_NAME,
		LOG_MESSAGE_COLUMN_NAME, TIMESTAMP_COLUMN_NAME}

	PREPARED_CREATE_TABLE = fmt.Sprintf(`create table if not exists %s(%s INTEGER PRIMARY KEY AUTOINCREMENT,
							                   %s TEXT NOT NULL CHECK(%s <> ""),
							                   %s TEXT NOT NULL CHECK(%s <> ""),
							                   %s INTEGER NOT NULL,
							                   %s INTEGER NOT NULL)`,
		TABLE_NAME, ID_COLUMN_NAME, PUBLISHER_ID_COLUMN_NAME, PUBLISHER_ID_COLUMN_NAME,
		LOG_MESSAGE_COLUMN_NAME, LOG_MESSAGE_COLUMN_NAME,
		LOG_LEVEL_COLUMN_NAME, TIMESTAMP_COLUMN_NAME)

	PREPARED_INSERT = fmt.Sprintf(`insert into %s(%s,%s,%s,%s) values(?, ?, ?, ?)`,
		TABLE_NAME, PUBLISHER_ID_COLUMN_NAME, LOG_LEVEL_COLUMN_NAME,
		LOG_MESSAGE_COLUMN_NAME, TIMESTAMP_COLUMN_NAME)

	PREPARED_DELETE = fmt.Sprintf("delete from %s where %s <= ", TABLE_NAME, TIMESTAMP_COLUMN_NAME)

	PREPARED_SELECT = fmt.Sprintf("select %s,%s,%s,%s", PUBLISHER_ID_COLUMN_NAME, LOG_MESSAGE_COLUMN_NAME,
		LOG_LEVEL_COLUMN_NAME, TIMESTAMP_COLUMN_NAME)
)

type LogMessage struct {
	Publisher string `json:"publisher"`
	TimeStamp int64  `json:"timestamp"` // will be received
	Level     string `json:"level"`
	Message   string `json:"message"`
}

type GetLogsRequest struct {
	TimeFrameStart int64    `json:"timeframestart"`
	TimeFrameEnd   int64    `json:"timeframeend"`
	Publishers     []string `json:"publishers"`
	LogLevel       string   `json:"level"`
	Message        string   `json:"message"`
	Page           int      `json:"page"`
	OrderBy        []string `json:"orderby"`
	Asc            bool     `json:"asc"`
	PageSize       int      `json:"pagesize"`
}

type GetLogsResponse struct {
	Logs     []LogMessage `json:"logs"`
	Size     int          `json:"size"`
	PageNum  int          `json:"page"`
	PageSize int          `json:"pagesize"`
}

type AddLogRequest struct {
	LogMessage
}

type LoggingConfig struct {
	AccessTokens   []string `json:"access_tokens"`
	CleanFrequency string   `json:"cleanfrequency"`
	TTL            string   `json:"ttl"`
}
