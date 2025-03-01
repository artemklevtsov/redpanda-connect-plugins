package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-querystring/query"
	"github.com/redpanda-data/benthos/v4/public/service"
	"github.com/stretchr/testify/assert"
)

func TestStatTableService_GetWithContext(t *testing.T) {
	testCases := []struct {
		name           string
		mockResponse   string
		mockStatusCode int
		query          *StatTableQuery
		expectedData   *StatTableResponse
		expectedError  error
	}{
		{
			name: "Successful Request",
			mockResponse: `{
				"query": {
					"ids": [123],
					"metrics": ["ym:s:visits"],
					"dimensions": ["ym:s:date"]
				},
				"data": [{ "dimensions": [{ "name": "2023-10-26" }], "metrics": [100] }],
				"total_rows": 1
			}`,
			mockStatusCode: http.StatusOK,
			query: &StatTableQuery{
				IDs:        []int{123},
				Metrics:    []string{"ym:s:visits"},
				Dimensions: []string{"ym:s:date"},
			},
			expectedData: &StatTableResponse{
				Query: &StatTableQuery{
					IDs:        []int{123},
					Metrics:    []string{"ym:s:visits"},
					Dimensions: []string{"ym:s:date"},
				},
				Data: []StatTableResponseEntry{
					{
						Dimensions: []struct {
							Name string `json:"name"`
							Id   string `json:"id,omitempty"`
						}{
							{Name: "2023-10-26"},
						},
						Metrics: []float64{100},
					},
				},
				TotalRows: 1,
			},
			expectedError: nil,
		},
		{
			name:           "Error Response",
			mockResponse:   `{"message": "Something went wrong", "code": 1}`,
			mockStatusCode: http.StatusBadRequest,
			query: &StatTableQuery{
				IDs:        []int{123},
				Metrics:    []string{"ym:s:visits"},
				Dimensions: []string{"ym:s:date"},
			},
			expectedData: nil,
			expectedError: &APIError{
				Message: "Something went wrong",
				Code:    1,
			},
		},
		{
			name: "Empty Response",
			mockResponse: `{
				"query": {
					"ids": [123],
					"metrics": ["ym:s:visits"],
					"dimensions": ["ym:s:date"]
				},
				"data": [],
				"total_rows": 0
			}`,
			mockStatusCode: http.StatusOK,
			query: &StatTableQuery{
				IDs:        []int{123},
				Metrics:    []string{"ym:s:visits"},
				Dimensions: []string{"ym:s:date"},
			},
			expectedData: &StatTableResponse{
				Query: &StatTableQuery{
					IDs:        []int{123},
					Metrics:    []string{"ym:s:visits"},
					Dimensions: []string{"ym:s:date"},
				},
				Data:      []StatTableResponseEntry{},
				TotalRows: 0,
			},
			expectedError: nil,
		},
		{
			name:           "Invalid JSON Response",
			mockResponse:   `{invalid}`,
			mockStatusCode: http.StatusOK,
			query: &StatTableQuery{
				IDs:        []int{123},
				Metrics:    []string{"ym:s:visits"},
				Dimensions: []string{"ym:s:date"},
			},
			expectedData:  nil,
			expectedError: errors.New(`invalid character 'i' looking for beginning of object key string`),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/data", r.URL.Path)
				assert.Equal(t, http.MethodGet, r.Method)
				w.WriteHeader(tc.mockStatusCode)

				exptectedQuery, _ := query.Values(tc.query)
				// realQuery, _ := query.Values(r.URL.Query())
				assert.Equal(t, exptectedQuery.Encode(), r.URL.RawQuery)

				fmt.Fprint(w, tc.mockResponse)
			}))
			defer server.Close()

			// Create a client
			client := NewClient("stat", "v1", "test_token", nil)
			client.client.SetBaseURL(server.URL)

			// Call GetWithContext
			data, err := client.StatTable.GetWithContext(context.Background(), tc.query)

			// Assertions
			if tc.expectedError != nil {
				assert.Error(t, err)

				//nolint:errorlint
				if _, ok := tc.expectedError.(*APIError); ok {
					var apiErr *APIError

					assert.ErrorAs(t, err, &apiErr)
					assert.Equal(t, tc.expectedError, err)
				} else {
					assert.EqualError(t, err, tc.expectedError.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedData, data)
			}
		})
	}
}

func TestStatTableResponse_Batch(t *testing.T) {
	testCases := []struct {
		name           string
		statResponse   StatTableResponse
		expectedBatch  service.MessageBatch
		expectedError  error
		expectedLength int
	}{
		{
			name: "Successful batch creation",
			statResponse: StatTableResponse{
				Query: &StatTableQuery{
					IDs:        []int{123},
					Metrics:    []string{"ym:s:visits", "ym:s:pageviews"},
					Dimensions: []string{"ym:s:date"},
					Date1:      "2023-10-26",
					Date2:      "2023-10-27",
					Limit:      10,
					Offset:     0,
				},
				Data: []StatTableResponseEntry{
					{
						Dimensions: []struct {
							Name string `json:"name"`
							Id   string `json:"id,omitempty"`
						}{
							{Name: "2023-10-26"},
						},
						Metrics: []float64{100, 200},
					},
					{
						Dimensions: []struct {
							Name string `json:"name"`
							Id   string `json:"id,omitempty"`
						}{
							{Name: "2023-10-27"},
						},
						Metrics: []float64{150, 250},
					},
				},
				TotalRows: 2,
			},
			expectedBatch: func() service.MessageBatch {
				msg1 := service.NewMessage(nil)
				msg1.SetStructuredMut(
					map[string]any{
						"date":      "2023-10-26",
						"visits":    100.0,
						"pageviews": 200.0,
					})
				msg1.MetaSetMut("query", map[string]any{
					"ids":        []int{123},
					"metrics":    []string{"ym:s:visits", "ym:s:pageviews"},
					"dimensions": []string{"ym:s:date"},
					"limit":      10,
					"offset":     0,
					"filters":    "",
					"date1":      "2023-10-26",
					"date2":      "2023-10-27",
					"accuracy":   "",
					"lang":       "",
					"preset":     "",
					"sort":       []string(nil),
					"timezone":   "",
				})
				msg1.MetaSetMut("limit", 10)
				msg1.MetaSetMut("offset", 0)
				msg1.MetaSetMut("total", 2)

				msg2 := service.NewMessage(nil)
				msg2.SetStructuredMut(map[string]any{
					"date":      "2023-10-27",
					"visits":    150.0,
					"pageviews": 250.0,
				})
				msg2.MetaSetMut("query", map[string]any{
					"ids":        []int{123},
					"metrics":    []string{"ym:s:visits", "ym:s:pageviews"},
					"dimensions": []string{"ym:s:date"},
					"limit":      10,
					"offset":     0,
					"filters":    "",
					"date1":      "2023-10-26",
					"date2":      "2023-10-27",
					"accuracy":   "",
					"lang":       "",
					"preset":     "",
					"sort":       []string(nil),
					"timezone":   "",
				})
				msg2.MetaSetMut("limit", 10)
				msg2.MetaSetMut("offset", 0)
				msg2.MetaSetMut("total", 2)

				return service.MessageBatch{msg1, msg2}
			}(),
			expectedError:  nil,
			expectedLength: 2,
		},
		{
			name: "Empty Data",
			statResponse: StatTableResponse{
				Query: &StatTableQuery{
					IDs:        []int{123},
					Metrics:    []string{"ym:s:visits"},
					Dimensions: []string{"ym:s:date"},
				},
				Data:      []StatTableResponseEntry{},
				TotalRows: 0,
			},
			expectedBatch:  nil,
			expectedError:  nil,
			expectedLength: 0,
		},
		{
			name: "Nil Data",
			statResponse: StatTableResponse{
				Query: &StatTableQuery{
					IDs:        []int{123},
					Metrics:    []string{"ym:s:visits"},
					Dimensions: []string{"ym:s:date"},
				},
				TotalRows: 0,
			},
			expectedBatch:  nil,
			expectedError:  nil,
			expectedLength: 0,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			batch, err := tc.statResponse.Batch()

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedLength, len(batch))

				for i := range tc.expectedBatch {
					expectedMsg, _ := tc.expectedBatch[i].AsStructured()
					realMsg, _ := batch[i].AsStructured()
					assert.Equal(t, expectedMsg, realMsg)

					meta := []string{"query", "limit", "offset", "total"}
					for _, m := range meta {
						expectedMeta, _ := tc.expectedBatch[i].MetaGetMut(m)
						realMeta, _ := batch[i].MetaGetMut(m)
						assert.Equal(t, expectedMeta, realMeta)
					}
				}
			}
		})
	}
}
