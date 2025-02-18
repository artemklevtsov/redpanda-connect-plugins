package stat_table

import (
	"fmt"
	"log/slog"
)

type apiResponse struct {
	Query     *query             `json:"query"`
	Data      []apiResponseEntry `json:"data"`
	TotalRows int                `json:"total_rows"`
}

type apiResponseEntry struct {
	Dimenctions []struct {
		Name string `json:"name"`
		Id   string `json:"id,omitempty"`
	} `json:"dimensions"`
	Metrics []float64 `json:"metrics"`
}

func (r *apiResponse) Rows(formayKeys bool) []map[string]any {
	data := make([]map[string]any, 0, len(r.Data))

	for _, e := range r.Data {
		row := make(map[string]any)

		for di, d := range e.Dimenctions {
			k := r.Query.Dimensions[di]
			if formayKeys {
				k = formatKey(k)
			}

			row[k] = d.Name
		}

		for mi, m := range e.Metrics {
			k := r.Query.Metrics[mi]
			if formayKeys {
				k = formatKey(k)
			}

			row[k] = m
		}

		data = append(data, row)
	}

	return data
}

// APIError API error object
// Source: https://yandex.ru/dev/metrika/doc/api2/management/concept/errors.html#errors__resp
type apiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Reasons []struct {
		ErrorType string `json:"error_type"`
		Message   string `json:"message"`
		Location  string `json:"location"`
	} `json:"errors"`
}

// API error string representation.
func (e apiError) Error() string {
	return fmt.Sprintf("Yandex.Metriika API error %d: %s", e.Code, e.Message)
}

func (e *apiError) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Int("code", e.Code),
		slog.String("message", e.Message),
	)
}
