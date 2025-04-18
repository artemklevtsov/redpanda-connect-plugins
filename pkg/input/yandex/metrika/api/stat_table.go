package api

import (
	"context"

	"github.com/artemklevtsov/redpanda-connect-plugins/internal/pkg/utils"
	"github.com/google/go-querystring/query"
	"github.com/redpanda-data/benthos/v4/public/service"
)

type StatTableService struct {
	client *Client
}

func (s *StatTableService) Get(q *StatTableQuery) (*StatTableResponse, error) {
	return s.GetWithContext(context.Background(), q)
}

func (s *StatTableService) GetWithContext(ctx context.Context, q *StatTableQuery) (*StatTableResponse, error) {
	var (
		err  error
		data StatTableResponse
	)

	values, err := query.Values(q)
	if err != nil {
		return nil, err
	}

	_, err = s.client.R().
		SetContext(ctx).
		SetQueryString(values.Encode()).
		SetSuccessResult(&data).
		Get("data")
	if err != nil {
		return nil, err
	}

	return &data, nil
}

// StatTableQuery represents a query for fetching data from Yandex.Metrika API stat tables.
type StatTableQuery struct {
	IDs          []int    `json:"ids" url:"ids,comma"`                                                       // IDs is a list of counter IDs.
	Date1        string   `json:"date1,omitempty" url:"date1,omitempty"`                                     // Date1 is the start date of the data range.
	Date2        string   `json:"date2,omitempty" url:"date2,omitempty"`                                     // Date2 is the end date of the data range.
	Dimensions   []string `json:"dimensions,omitempty" url:"dimensions,comma,omitempty"`                     // Dimensions is a list of dimensions for data breakdown.
	Metrics      []string `json:"metrics,omitempty" url:"metrics,comma,omitempty"`                           // Metrics is a list of metrics to be fetched.
	Sort         []string `json:"sort,omitempty" url:"sort,comma,omitempty"`                                 // Sort is a list of fields to sort the data by.
	Accuracy     string   `json:"accuracy,omitempty" url:"accuracy,omitempty"`                               // Accuracy is a sample size for the report.
	Limit        int      `json:"limit,omitempty" url:"limit,omitempty"`                                     // Limit is the maximum number of rows to return.
	Offset       int      `json:"offset,omitempty" url:"offset,omitempty"`                                   // Offset is the offset of the first row to return.
	Filters      string   `json:"filters,omitempty" url:"filters,omitempty"`                                 // Filters is a filter string to apply to the data.
	Lang         string   `json:"lang,omitempty" url:"lang,omitempty"`                                       // Lang is the language for the response.
	Preset       string   `json:"preset,omitempty" url:"preset,omitempty"`                                   // Preset is the preset used for the query.
	Timezone     string   `json:"timezone,omitempty" url:"timezone,omitempty"`                               // Timezone is the timezone to use for the data.
	DirectLogins []string `json:"direct_client_logins,omitempty" url:"direct_client_logins,comma,omitempty"` // DirectLogins is a list of direct client logins.
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
			k = utils.ProcessKey(k)
			row[k] = d.Name
		}

		for mi, m := range e.Metrics {
			k := r.Query.Metrics[mi]
			k = utils.ProcessKey(k)
			row[k] = m
		}

		query, err := utils.StructToMap(r.Query)
		if err != nil {
			return nil, err
		}

		msg := service.NewMessage(nil)
		msg.SetStructuredMut(row)
		msg.MetaSetMut("query", query)
		msg.MetaSetMut("limit", r.Query.Limit)
		msg.MetaSetMut("offset", r.Query.Offset)
		msg.MetaSetMut("total", r.TotalRows)

		msgs[i] = msg
	}

	return msgs, nil
}
