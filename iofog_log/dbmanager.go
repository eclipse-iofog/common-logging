package iofog_log

import (
	"database/sql"
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"strings"
	"time"
	"fmt"
)

type DBManager struct {
	db          *sql.DB
	cleanTicker *time.Ticker
	stopChannel chan int
}

func newDBManager() (*DBManager, error) {
	db, err := sql.Open("sqlite3", DB_LOCATION + DB_NAME)
	if err != nil {
		return nil, err
	}

	manager := new(DBManager)
	manager.db = db
	manager.stopChannel = make(chan int)
	_, err = db.Exec(PREPARED_CREATE_TABLE)
	if err != nil {
		return nil, err
	}

	return manager, nil
}

func (manager *DBManager) Close() {
	if err := manager.db.Close(); err != nil {
		logger.Println("Error while closing db", err)
	} else {
		logger.Println("DB successfully closed")
	}
}

func (manager *DBManager) clearDB(ttl time.Duration) (int64, error) {
	timestamp_end := (time.Now().Add(- ttl)).UnixNano() / 1000000
	logger.Printf("Edge timestamp for deletion is %d\n", timestamp_end)
	delete_stmt := fmt.Sprint(PREPARED_DELETE, timestamp_end)
	result, err := manager.db.Exec(delete_stmt)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (manager *DBManager) insert(msg *LogMessage) (int64, error) {
	stmt, err := manager.db.Prepare(PREPARED_INSERT)
	if err != nil {
		return 0, errors.New("Error while preparing instert: " + err.Error())
	}
	defer stmt.Close()
	level, ok := _levelNames[strings.ToUpper(msg.Level)]
	if !ok {
		level = NOTSET
	}
	result, err := stmt.Exec(msg.Publisher, level, msg.Message, msg.TimeStamp)
	if err != nil {
		return 0, errors.New("Error while executing instert: " + err.Error())
	}
	return result.LastInsertId()
}

func (manager *DBManager) query(request *GetLogsRequest) (*GetLogsResponse, error) {
	select_stmt, err := buildQuery(request)
	if err != nil {
		return nil, err
	}
	rows, err := manager.db.Query(select_stmt)
	if err != nil {
		logger.Println(err.Error())
		return nil, errors.New("Error while executing query: " + err.Error())
	}
	defer rows.Close()

	logs := make([]LogMessage, 0, 10)
	var response GetLogsResponse
	for rows.Next() {
		var lvl int
		var logMsg LogMessage
		err = rows.Scan(&logMsg.Publisher, &logMsg.Message, &lvl, &logMsg.TimeStamp)
		if err != nil {
			logger.Println(err)
		}
		logMsg.Level = _levelNames[lvl].(string)
		logs = append(logs, logMsg)

	}
	response.Logs = logs
	response.Size = len(logs)
	response.PageNum = request.Page
	response.PageSize = request.PageSize
	err = rows.Err()
	if err != nil {
		logger.Println(err)
	}
	return &response, nil

}

func (manager *DBManager) cleanRoutine(ttl time.Duration) {
	defer func() {
		manager.stopChannel <- 0
	}()
	defer logger.Println("Clean watcher stopped")
	for {
		select {
		case <-manager.cleanTicker.C:
			if deleted, err := manager.clearDB(ttl); err != nil {
				logger.Println("Error while cleaning db: " + err.Error())
			} else {
				logger.Printf("Deleted rows: %d\n", deleted)
			}
		case <-manager.stopChannel:
			return
		}
	}
}

func (manager *DBManager) setCleanInterval(frequency, ttl time.Duration) {
	if manager.cleanTicker != nil {
		manager.cleanTicker.Stop()
		manager.stopChannel <- 0
		<-manager.stopChannel
	}
	logger.Printf("New cleaning frequency is %v", frequency)
	logger.Printf("New ttl is %v", ttl)
	manager.cleanTicker = time.NewTicker(frequency)
	go manager.cleanRoutine(ttl)
}
