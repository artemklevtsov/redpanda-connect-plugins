package api

import "strings"

// LogRequestQuery represents a query for requesting logs from the Yandex.Metrika API.
type LogRequestQuery struct {
	Source      string   `json:"source"`      // Source specifies the source of the logs.
	Date1       string   `json:"date1"`       // Date1 is the start date for the log request in YYYY-MM-DD format.
	Date2       string   `json:"date2"`       // Date2 is the end date for the log request in YYYY-MM-DD format.
	Fields      []string `json:"fields"`      // Fields is a list of fields to include in the log request.
	Attribution string   `json:"attribution"` // Attribution specifies the attribution model.
}

// Map returns a map representation of the LogRequestQuery.
func (q *LogRequestQuery) Map() map[string]any {
	m := make(map[string]any)
	m["source"] = q.Source
	m["date1"] = q.Date1
	m["date2"] = q.Date2
	m["fields"] = q.Fields
	m["attribution"] = q.Attribution

	return m
}

// Params returns a map of query parameters for the LogRequestQuery, suitable for API requests.
func (q *LogRequestQuery) Params() map[string]string {
	m := make(map[string]string)
	m["source"] = q.Source
	m["date1"] = q.Date1
	m["date2"] = q.Date2
	m["fields"] = strings.Join(q.Fields, ",")
	m["attribution"] = q.Attribution

	for k := range m {
		if len(m[k]) == 0 {
			delete(m, k)
		}
	}

	return m
}

// EvalLogRequestResponse represents the response for evaluating a log request.
type EvalLogRequestResponse struct {
	Result EvalLogRequestResponseEntry `json:"log_request_evaluation"` // Result contains the evaluation result.
}

// EvalLogRequestResponseEntry represents a log request evaluation entry.
type EvalLogRequestResponseEntry struct {
	IsPossible bool `json:"possible"`                  // IsPossible indicates whether the log request is possible.
	MaxDays    int  `json:"max_possible_day_quantity"` // MaxDays is the maximum number of days for which logs can be requested.
}

// LogRequestResponse represents the response for creating or retrieving a log request.
type LogRequestResponse struct {
	Request LogRequestResponseEntry `json:"log_request"` // Request contains the log request details.
}

// LogRequestResponseEntry represents a log request entry.
type LogRequestResponseEntry struct {
	LogRequestQuery
	RequestID uint64 `json:"request_id"` // RequestID is the ID of the log request.
	CounterID uint64 `json:"counter_id"` // CounterID is the ID of the counter associated with the log request.
	Status    string `json:"status"`     // Status is the status of the log request.
	Size      uint64 `json:"size"`       // Size is the size of the log request.
	Parts     []struct {
		Number int    `json:"part_number"` // Num is the part number.
		Size   uint64 `json:"size"`        // Size is the size of the part.
	} // Parts is a list of parts of the log request.
}
