package iofog_log

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"strings"
	"time"
)

type DBManager struct {
	db          *sql.DB
	cleanTicker *time.Ticker
	stopChannel chan bool
}

func NewDBManager() (*DBManager, error) {
	db, err := sql.Open("sqlite3", DB_LOCATION+DB_NAME)
	if err != nil {
		return nil, err
	}

	manager := new(DBManager)
	manager.db = db
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

func (manager *DBManager) Delete() (int64, error) {
	result, err := manager.db.Exec(PREPARED_DELETE)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (manager *DBManager) Insert(msg *LogMessage) (int64, error) {
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

func (manager *DBManager) Query(request *GetLogsRequest) (*GetLogsResponse, error) {
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

func buildQuery(request *GetLogsRequest) (string, error) {
	level, ok := _levelNames[strings.ToUpper(request.LogLevel)]
	if !ok {
		level = NOTSET
	}
	if request.Page == 0 {
		request.Page += 1
	}

	select_stmt := fmt.Sprintf(`%s from %s where %s>=%d `,
		PREPARED_SELECT, TABLE_NAME, LOG_LEVEL_COLUMN_NAME, level)

	if len(request.Publishers) > 0 {
		select_stmt += fmt.Sprintf(` and %s in ("%s") `,
			PUBLISHER_ID_COLUMN_NAME, strings.Join(request.Publishers, `", "`))
	}
	if request.TimeFrameStart != 0 {
		select_stmt += fmt.Sprintf(` and %s>=%d `, TIMESTAMP_COLUMN_NAME, request.TimeFrameStart)
	}
	if request.TimeFrameEnd != 0 {
		select_stmt += fmt.Sprintf(` and %s<=%d `, TIMESTAMP_COLUMN_NAME, request.TimeFrameEnd)
	}
	if len(request.Message) > 0 {
		select_stmt += fmt.Sprintf(` and %s like "%%%s%%"`, LOG_MESSAGE_COLUMN_NAME, request.Message)
	}
	if len(request.OrderBy) > 0 {
	outer:
		for _, order := range request.OrderBy {
			for _, field := range ORDER_BY_FIELDS {
				if order == field {
					continue outer
				}
			}
			return "", errors.New("No such column: " + order)
		}
		select_stmt += fmt.Sprintf(` order by %s `, strings.Join(request.OrderBy, `, `))
	} else {
		select_stmt += fmt.Sprintf(` order by %s `, DEFAULT_ORDER_BY)
	}
	if request.Asc {
		select_stmt += fmt.Sprintf(` %s `, ASC)
	} else {
		select_stmt += fmt.Sprintf(` %s `, DESC)

	}

	if request.PageSize == 0 {
		request.PageSize = DEFAULT_PAGE_SIZE
	}
	select_stmt += fmt.Sprintf(` limit %d offset %d `, request.PageSize, (request.Page-1)*request.PageSize)
	return select_stmt, nil
}

func (manager *DBManager) cleanRoutine(stopChannel <-chan bool) {
	defer logger.Println("Ticker stopped")
	for {
		select {
		case <-manager.cleanTicker.C:
			if deleted, err := manager.Delete(); err != nil {
				logger.Println("Error while cleaning db: " + err.Error())
			} else {
				logger.Printf("Deleted rows: %d\n", deleted)
			}
		case <-stopChannel:
			return
		}
	}
}

func (manager *DBManager) SetCleanInterval(d time.Duration) {
	if manager.cleanTicker != nil {
		manager.cleanTicker.Stop()
		close(manager.stopChannel)
	}
	logger.Printf("Setting cleaning interval of %v", d)
	manager.cleanTicker = time.NewTicker(d)
	manager.stopChannel = make(chan bool)
	go manager.cleanRoutine(manager.stopChannel)
}
