package iofog_log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func buildQuery(request *GetLogsRequest) (string, error) {
	level, ok := levelNames[strings.ToUpper(request.LogLevel)]
	if !ok {
		level = NOTSET
	}
	if request.Page == 0 {
		request.Page += 1
	}
	if request.PageSize == 0 {
		request.PageSize = DEFAULT_PAGE_SIZE
	}

	var select_stmt bytes.Buffer
	fmt.Fprintf(&select_stmt, `%s where %s>=%d`, PREPARED_SELECT, LOG_LEVEL_COLUMN_NAME, level)
	if request.Publishers != nil {
		fmt.Fprintf(&select_stmt, ` and %s in ("%s") `, PUBLISHER_ID_COLUMN_NAME, strings.Join(request.Publishers, `", "`))
	}
	if request.TimeFrameStart != 0 {
		fmt.Fprintf(&select_stmt, ` and %s>=%d `, TIMESTAMP_COLUMN_NAME, request.TimeFrameStart)
	}
	if request.TimeFrameEnd != 0 {
		fmt.Fprintf(&select_stmt, ` and %s<=%d `, TIMESTAMP_COLUMN_NAME, request.TimeFrameEnd)
	}
	if len(request.Message) > 0 {
		fmt.Fprintf(&select_stmt, ` and %s like "%%%s%%" `, LOG_MESSAGE_COLUMN_NAME, request.Message)
	}
	var orderBy string
	order := DESC
	if request.OrderBy != nil {
		orderBy = strings.Join(request.OrderBy, `, `)
	} else {
		orderBy = DEFAULT_ORDER_BY
	}
	if request.Asc {
		order = ASC
	}

	fmt.Fprintf(&select_stmt, ` order by %s %s limit %d offset %d`, orderBy, order, request.PageSize, (request.Page-1)*request.PageSize)
	return select_stmt.String(), nil
}

func getJsonBody(w http.ResponseWriter, r *http.Request, decoded interface{}) bool {
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(decoded)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		logger.Println(err.Error())
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
