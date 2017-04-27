package iofog_log

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

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

func getJsonBody(w http.ResponseWriter, r *http.Request, decoded interface{}) bool {
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(decoded)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}
	return true
}

func checkMethodAndContent(method, contentType string, w http.ResponseWriter, r *http.Request) bool {
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

func isAuthorized(accessTokens []string, w http.ResponseWriter, r *http.Request) bool {
	accessToken := r.Header.Get(ACCESS_TOKEN)
	for _, token := range accessTokens {
		if accessToken == token {
			return true
		}
	}
	w.WriteHeader(http.StatusUnauthorized)
	return false
}

func parseDuration(duration, description string, defaultDuration time.Duration) time.Duration {
	useDefault := false
	var parsed time.Duration
	var err error
	if len(duration) > 0 {
		parsed, err = time.ParseDuration(fmt.Sprintf("%s", duration))
		if err != nil {
			logger.Printf("Error while parsing %s: %s.", description, err.Error())
			useDefault = true
		} else if parsed <= 0 {
			logger.Printf("Unable to specify %s as %s\n", duration, description)
			useDefault = true
		}
	} else {
		useDefault = true
	}
	if useDefault {
		logger.Printf("Using default value for %s\n", description)
		parsed = defaultDuration
	}
	return parsed
}