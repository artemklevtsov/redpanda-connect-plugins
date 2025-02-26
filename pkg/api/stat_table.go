package api

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/artemklevtsov/redpanda-connect-yandex-metrika/pkg/misc"
	"github.com/redpanda-data/benthos/v4/public/service"
)

type StatTableService struct {
	client *Client
}

func (s *StatTableService) Get(query *StatTableQuery) (*StatTableResponse, error) {
	return s.GetWithContext(context.Background(), query)
}

func (s *StatTableService) GetWithContext(ctx context.Context, query *StatTableQuery) (*StatTableResponse, error) {
	var data StatTableResponse

	_, err := s.client.R().
		SetContext(ctx).
		SetQueryParams(query.Params()).
		SetSuccessResult(&data).
		Get("data")

	if err != nil {
		return nil, err
	}

	return &data, nil
}

// StatTableQuery represents a query for fetching data from Yandex.Metrika API stat tables.
type StatTableQuery struct {
	IDs          []int    `json:"ids"`                  // IDs is a list of counter IDs.
	Date1        string   `json:"date1,omitempty"`      // Date1 is the start date of the data range.
	Date2        string   `json:"date2,omitempty"`      // Date2 is the end date of the data range.
	Dimensions   []string `json:"dimensions,omitempty"` // Dimensions is a list of dimensions for data breakdown.
	Metrics      []string `json:"metrics"`              // Metrics is a list of metrics to be fetched.
	Sort         []string `json:"sort,omitempty"`       // Sort is a list of fields to sort the data by.
	Limit        int      `json:"limit"`                // Limit is the maximum number of rows to return.
	Offset       int      `json:"offset"`               // Offset is the offset of the first row to return.
	Filters      string   `json:"filters,omitempty"`    // Filters is a filter string to apply to the data.
	Lang         string   `json:"lang,omitempty"`       // Lang is the language for the response.
	Preset       string   `json:"preset,omitempty"`     // Preset is the preset used for the query.
	Timezone     string   `json:"timezone,omitempty"`   // Timezone is the timezone to use for the data.
	DirectLogins []string `json:"direct_client_logins"` // DirectLogins is a list of direct client logins.
}

// Map returns a map representation of the StatTableQuery.
func (q *StatTableQuery) Map() map[string]any {
	m := make(map[string]any)
	m["ids"] = q.IDs
	m["filters"] = q.Filters
	m["date1"] = q.Date1
	m["date2"] = q.Date2
	m["dimensions"] = q.Dimensions
	m["metrics"] = q.Metrics
	m["lang"] = q.Lang
	m["preset"] = q.Preset
	m["sort"] = q.Sort
	m["timezone"] = q.Timezone
	m["limit"] = q.Limit
	m["offset"] = q.Offset

	return m
}

// Params returns a map of query parameters for the StatTableQuery, suitable for API requests.
func (q *StatTableQuery) Params() map[string]string {
	m := make(map[string]string)
	m["ids"] = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(q.IDs)), ","), "[]")
	m["filters"] = q.Filters
	m["date1"] = q.Date1
	m["date2"] = q.Date2
	m["dimensions"] = strings.Join(q.Dimensions, ",")
	m["metrics"] = strings.Join(q.Metrics, ",")
	m["lang"] = q.Lang
	m["preset"] = q.Preset
	m["sort"] = strings.Join(q.Sort, ",")
	m["timezone"] = q.Timezone
	m["limit"] = strconv.Itoa(q.Limit)
	m["offset"] = strconv.Itoa(q.Offset)

	for k := range m {
		if len(m[k]) == 0 {
			delete(m, k)
		}
	}

	return m
}

// StatTableResponse represents the response from a stat table query.
type StatTableResponse struct {
	Query     *StatTableQuery          `json:"query"`      // Query contains the query parameters used to fetch this data.
	Data      []StatTableResponseEntry `json:"data"`       // Data contains the actual data returned by the query.
	TotalRows int                      `json:"total_rows"` // TotalRows is the total number of rows matching the query.
}

// StatTableResponseEntry represents a single row in the stat table data.
type StatTableResponseEntry struct {
	Dimensions []struct {
		Name string `json:"name"`         // Name is the name of the dimension.
		Id   string `json:"id,omitempty"` // Id is the optional ID of the dimension.
	} `json:"dimensions"` // Dimensions is a list of dimensions for this row.
	Metrics []float64 `json:"metrics"` // Metrics is a list of metrics for this row.
}

// Batch creates a service.MessageBatch from the StatTableResponse.
func (r *StatTableResponse) Batch() (service.MessageBatch, error) {
	if r.Data == nil {
		return nil, nil
	}

	msgs := make(service.MessageBatch, len(r.Data))

	for i, e := range r.Data {
		row := make(map[string]any)

		for di, d := range e.Dimensions {
			k := r.Query.Dimensions[di]
			k = misc.ProcessKey(k)
			row[k] = d.Name
		}

		for mi, m := range e.Metrics {
			k := r.Query.Metrics[mi]
			k = misc.ProcessKey(k)
			row[k] = m
		}

		msg := service.NewMessage(nil)
		msg.SetStructuredMut(row)
		msg.MetaSetMut("query", r.Query.Map())
		msg.MetaSetMut("limit", r.Query.Limit)
		msg.MetaSetMut("offset", r.Query.Offset)
		msg.MetaSetMut("total", r.TotalRows)

		msgs[i] = msg
	}

	return msgs, nil
}
