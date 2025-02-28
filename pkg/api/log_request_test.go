package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-querystring/query"
	"github.com/stretchr/testify/assert"
)

func TestLogRequestService_EvalWithContext(t *testing.T) {
	testCases := []struct {
		name           string
		mockResponse   string
		mockStatusCode int
		query          *LogRequestQuery
		expectedData   *EvalLogRequestResponse
		expectedError  error
	}{
		{
			name: "Successful Request",
			mockResponse: `{
				"log_request_evaluation": {
					"possible": true,
					"max_possible_day_quantity": 30
				}
			}`,
			mockStatusCode: http.StatusOK,
			query: &LogRequestQuery{
				Source:      "visits",
				Date1:       "2023-01-01",
				Date2:       "2023-01-02",
				Fields:      []string{"field1", "field2"},
				Attribution: "last",
			},
			expectedData: &EvalLogRequestResponse{
				Result: EvalLogRequestResponseEntry{
					IsPossible: true,
					MaxDays:    30,
				},
			},
			expectedError: nil,
		},
		{
			name:           "Error Response",
			mockResponse:   `{"message": "Something went wrong", "code": 1}`,
			mockStatusCode: http.StatusBadRequest,
			query: &LogRequestQuery{
				Source:      "visits",
				Date1:       "2023-01-01",
				Date2:       "2023-01-02",
				Fields:      []string{"field1", "field2"},
				Attribution: "last",
			},
			expectedData: nil,
			expectedError: &APIError{
				Message: "Something went wrong",
				Code:    1,
			},
		},
		{
			name:           "Empty Response",
			mockResponse:   `{"log_request_evaluation": {"possible": false, "max_possible_day_quantity": 0}}`,
			mockStatusCode: http.StatusOK,
			query: &LogRequestQuery{
				Source:      "visits",
				Date1:       "2023-01-01",
				Date2:       "2023-01-02",
				Fields:      []string{"field1", "field2"},
				Attribution: "last",
			},
			expectedData: &EvalLogRequestResponse{
				Result: EvalLogRequestResponseEntry{
					IsPossible: false,
					MaxDays:    0,
				},
			},
			expectedError: nil,
		},
		{
			name:           "Invalid JSON Response",
			mockResponse:   `{invalid}`,
			mockStatusCode: http.StatusOK,
			query: &LogRequestQuery{
				Source:      "visits",
				Date1:       "2023-01-01",
				Date2:       "2023-01-02",
				Fields:      []string{"field1", "field2"},
				Attribution: "last",
			},
			expectedData:  nil,
			expectedError: errors.New(`invalid character 'i' looking for beginning of object key string`),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/counter/123/logrequests/evaluate", r.URL.Path)
				assert.Equal(t, http.MethodGet, r.Method)

				exptectedQuery, _ := query.Values(tc.query)
				assert.Equal(t, exptectedQuery.Encode(), r.URL.RawQuery)

				w.WriteHeader(tc.mockStatusCode)
				fmt.Fprint(w, tc.mockResponse)
			}))
			defer server.Close()

			// Create a client
			client := NewClient("management", "v1", "test_token", nil)
			client.client.SetBaseURL(server.URL)

			// Call GetWithContext
			data, err := client.LogRequest.EvalWithContext(context.Background(), 123, tc.query)

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

func TestLogRequestService_CreateWithContext(t *testing.T) {
	testCases := []struct {
		name           string
		mockResponse   string
		mockStatusCode int
		query          *LogRequestQuery
		expectedData   *LogRequestResponse
		expectedError  error
	}{
		{
			name: "Successful Request",
			mockResponse: `{
				"log_request": {
					"source": "visits",
					"date1": "2023-01-01",
					"date2": "2023-01-02",
					"fields": ["field1", "field2"],
					"attribution": "last",
					"request_id": 1,
					"counter_id": 123,
					"status": "created"
				}
			}`,
			mockStatusCode: http.StatusOK,
			query: &LogRequestQuery{
				Source:      "visits",
				Date1:       "2023-01-01",
				Date2:       "2023-01-02",
				Fields:      []string{"field1", "field2"},
				Attribution: "last",
			},
			expectedData: &LogRequestResponse{
				Request: LogRequestResponseEntry{
					LogRequestQuery: LogRequestQuery{
						Source:      "visits",
						Date1:       "2023-01-01",
						Date2:       "2023-01-02",
						Fields:      []string{"field1", "field2"},
						Attribution: "last",
					},
					RequestID: 1,
					CounterID: 123,
					Status:    "created",
				},
			},
			expectedError: nil,
		},
		{
			name:           "Error Response",
			mockResponse:   `{"message": "Something went wrong", "code": 1}`,
			mockStatusCode: http.StatusBadRequest,
			query: &LogRequestQuery{
				Source:      "visits",
				Date1:       "2023-01-01",
				Date2:       "2023-01-02",
				Fields:      []string{"field1", "field2"},
				Attribution: "last",
			},
			expectedData: nil,
			expectedError: &APIError{
				Message: "Something went wrong",
				Code:    1,
			},
		},
		{
			name:           "Invalid JSON Response",
			mockResponse:   `{invalid}`,
			mockStatusCode: http.StatusOK,
			query: &LogRequestQuery{
				Source:      "visits",
				Date1:       "2023-01-01",
				Date2:       "2023-01-02",
				Fields:      []string{"field1", "field2"},
				Attribution: "last",
			},
			expectedData:  nil,
			expectedError: errors.New(`invalid character 'i' looking for beginning of object key string`),
		},
		{
			name:           "Empty params",
			mockResponse:   `{"log_request": {"request_id": 1, "counter_id": 123, "status": "created"}}`,
			mockStatusCode: http.StatusOK,
			query: &LogRequestQuery{
				Source:      "",
				Date1:       "",
				Date2:       "",
				Fields:      []string{},
				Attribution: "",
			},
			expectedData: &LogRequestResponse{
				Request: LogRequestResponseEntry{
					LogRequestQuery: LogRequestQuery{
						Source:      "",
						Date1:       "",
						Date2:       "",
						Fields:      []string(nil),
						Attribution: "",
					},
					RequestID: 1,
					CounterID: 123,
					Status:    "created",
				},
			},
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/counter/123/logrequests", r.URL.Path)
				assert.Equal(t, http.MethodPost, r.Method)

				w.WriteHeader(tc.mockStatusCode)
				fmt.Fprint(w, tc.mockResponse)
			}))
			defer server.Close()

			client := NewClient("management", "v1", "test_token", nil)
			client.client.SetBaseURL(server.URL)

			data, err := client.LogRequest.CreateWithContext(context.Background(), 123, tc.query)

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

func TestLogRequestService_GetWithContext(t *testing.T) {
	testCases := []struct {
		name           string
		mockResponse   string
		mockStatusCode int
		counter        int
		request        uint64
		expectedData   *LogRequestResponse
		expectedError  error
	}{
		{
			name: "Successful Request",
			mockResponse: `{
				"log_request": {
					"source": "visits",
					"date1": "2023-01-01",
					"date2": "2023-01-02",
					"fields": ["field1", "field2"],
					"attribution": "last",
					"request_id": 1,
					"counter_id": 123,
					"status": "processed",
					"size": 1024,
					"parts": [{"part_number": 1, "size": 512}, {"part_number": 2, "size": 512}]
				}
			}`,
			mockStatusCode: http.StatusOK,
			counter:        123,
			request:        1,
			expectedData: &LogRequestResponse{
				Request: LogRequestResponseEntry{
					LogRequestQuery: LogRequestQuery{
						Source:      "visits",
						Date1:       "2023-01-01",
						Date2:       "2023-01-02",
						Fields:      []string{"field1", "field2"},
						Attribution: "last",
					},
					RequestID: 1,
					CounterID: 123,
					Status:    "processed",
					Size:      1024,
					Parts: []struct {
						Number int    `json:"part_number"`
						Size   uint64 `json:"size"`
					}{
						{Number: 1, Size: 512},
						{Number: 2, Size: 512},
					},
				},
			},
			expectedError: nil,
		},
		{
			name:           "Error Response",
			mockResponse:   `{"message": "Something went wrong", "code": 1}`,
			mockStatusCode: http.StatusBadRequest,
			counter:        123,
			request:        1,
			expectedData:   nil,
			expectedError: &APIError{
				Message: "Something went wrong",
				Code:    1,
			},
		},
		{
			name:           "Invalid JSON Response",
			mockResponse:   `{invalid}`,
			mockStatusCode: http.StatusOK,
			counter:        123,
			request:        1,
			expectedData:   nil,
			expectedError:  errors.New(`invalid character 'i' looking for beginning of object key string`),
		},
		{
			name:           "Empty response",
			mockResponse:   `{"log_request": {}}`,
			mockStatusCode: http.StatusOK,
			counter:        123,
			request:        1,
			expectedData: &LogRequestResponse{
				Request: LogRequestResponseEntry{},
			},
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, fmt.Sprintf("/counter/%d/logrequest/%d", tc.counter, tc.request), r.URL.Path)
				assert.Equal(t, http.MethodGet, r.Method)

				w.WriteHeader(tc.mockStatusCode)
				fmt.Fprint(w, tc.mockResponse)
			}))
			defer server.Close()

			client := NewClient("management", "v1", "test_token", nil)
			client.client.SetBaseURL(server.URL)

			data, err := client.LogRequest.GetWithContext(context.Background(), tc.counter, tc.request)

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

func TestLogRequestService_CancelWithContext(t *testing.T) {
	testCases := []struct {
		name           string
		mockResponse   string
		mockStatusCode int
		counter        int
		request        uint64
		expectedData   *LogRequestResponse
		expectedError  error
	}{
		{
			name: "Successful Request",
			mockResponse: `{
				"log_request": {
					"source": "visits",
					"date1": "2023-01-01",
					"date2": "2023-01-02",
					"fields": ["field1", "field2"],
					"attribution": "last",
					"request_id": 1,
					"counter_id": 123,
					"status": "canceled"
				}
			}`,
			mockStatusCode: http.StatusOK,
			counter:        123,
			request:        1,
			expectedData: &LogRequestResponse{
				Request: LogRequestResponseEntry{
					LogRequestQuery: LogRequestQuery{
						Source:      "visits",
						Date1:       "2023-01-01",
						Date2:       "2023-01-02",
						Fields:      []string{"field1", "field2"},
						Attribution: "last",
					},
					RequestID: 1,
					CounterID: 123,
					Status:    "canceled",
				},
			},
			expectedError: nil,
		},
		{
			name:           "Error Response",
			mockResponse:   `{"message": "Something went wrong", "code": 1}`,
			mockStatusCode: http.StatusBadRequest,
			counter:        123,
			request:        1,
			expectedData:   nil,
			expectedError: &APIError{
				Message: "Something went wrong",
				Code:    1,
			},
		},
		{
			name:           "Invalid JSON Response",
			mockResponse:   `{invalid}`,
			mockStatusCode: http.StatusOK,
			counter:        123,
			request:        1,
			expectedData:   nil,
			expectedError:  errors.New(`invalid character 'i' looking for beginning of object key string`),
		},
		{
			name: "Empty response",
			mockResponse: `{
				"log_request": {}
			}`,
			mockStatusCode: http.StatusOK,
			counter:        123,
			request:        1,
			expectedData: &LogRequestResponse{
				Request: LogRequestResponseEntry{},
			},
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, fmt.Sprintf("/counter/%d/logrequest/%d/cancel", tc.counter, tc.request), r.URL.Path)
				assert.Equal(t, http.MethodPost, r.Method)

				w.WriteHeader(tc.mockStatusCode)
				fmt.Fprint(w, tc.mockResponse)
			}))
			defer server.Close()

			client := NewClient("management", "v1", "test_token", nil)
			client.client.SetBaseURL(server.URL)

			data, err := client.LogRequest.CancelWithContext(context.Background(), tc.counter, tc.request)

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

func TestLogRequestService_CleanWithContext(t *testing.T) {
	testCases := []struct {
		name           string
		mockResponse   string
		mockStatusCode int
		counter        int
		request        uint64
		expectedData   *LogRequestResponse
		expectedError  error
	}{
		{
			name: "Successful Request",
			mockResponse: `{
				"log_request": {
					"source": "visits",
					"date1": "2023-01-01",
					"date2": "2023-01-02",
					"fields": ["field1", "field2"],
					"attribution": "last",
					"request_id": 1,
					"counter_id": 123,
					"status": "cleaned_by_user"
				}
			}`,
			mockStatusCode: http.StatusOK,
			counter:        123,
			request:        1,
			expectedData: &LogRequestResponse{
				Request: LogRequestResponseEntry{
					LogRequestQuery: LogRequestQuery{
						Source:      "visits",
						Date1:       "2023-01-01",
						Date2:       "2023-01-02",
						Fields:      []string{"field1", "field2"},
						Attribution: "last",
					},
					RequestID: 1,
					CounterID: 123,
					Status:    "cleaned_by_user",
				},
			},
			expectedError: nil,
		},
		{
			name:           "Error Response",
			mockResponse:   `{"message": "Something went wrong", "code": 1}`,
			mockStatusCode: http.StatusBadRequest,
			counter:        123,
			request:        1,
			expectedData:   nil,
			expectedError: &APIError{
				Message: "Something went wrong",
				Code:    1,
			},
		},
		{
			name:           "Invalid JSON Response",
			mockResponse:   `{invalid}`,
			mockStatusCode: http.StatusOK,
			counter:        123,
			request:        1,
			expectedData:   nil,
			expectedError:  errors.New(`invalid character 'i' looking for beginning of object key string`),
		},
		{
			name: "Empty response",
			mockResponse: `{
				"log_request": {}
			}`,
			mockStatusCode: http.StatusOK,
			counter:        123,
			request:        1,
			expectedData: &LogRequestResponse{
				Request: LogRequestResponseEntry{},
			},
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, fmt.Sprintf("/counter/%d/logrequest/%d/clean", tc.counter, tc.request), r.URL.Path)
				assert.Equal(t, http.MethodPost, r.Method)

				w.WriteHeader(tc.mockStatusCode)
				fmt.Fprint(w, tc.mockResponse)
			}))
			defer server.Close()

			client := NewClient("management", "v1", "test_token", nil)
			client.client.SetBaseURL(server.URL)

			data, err := client.LogRequest.CleanWithContext(context.Background(), tc.counter, tc.request)

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

func TestLogRequestService_DownloadWithContext(t *testing.T) {
	testCases := []struct {
		name           string
		mockResponse   string
		mockStatusCode int
		counter        int
		request        uint64
		part           int
		expectedData   string
		expectedError  error
	}{
		{
			name:           "Successful Request",
			mockResponse:   `col1\tcol2\nval1\tval2`,
			mockStatusCode: http.StatusOK,
			counter:        123,
			request:        1,
			part:           1,
			expectedData:   `col1\tcol2\nval1\tval2`,
			expectedError:  nil,
		},
		{
			name:           "Error Response",
			mockResponse:   `{"message": "Something went wrong", "code": 1}`,
			mockStatusCode: http.StatusBadRequest,
			counter:        123,
			request:        1,
			part:           1,
			expectedData:   "",
			expectedError:  &APIError{Message: "Something went wrong", Code: 1},
		},
		{
			name:           "Empty Response",
			mockResponse:   "",
			mockStatusCode: http.StatusOK,
			counter:        123,
			request:        1,
			part:           1,
			expectedData:   "",
			expectedError:  nil,
		},
		{
			name:           "Headers only",
			mockResponse:   "",
			mockStatusCode: http.StatusTeapot,
			counter:        123,
			request:        1,
			part:           1,
			expectedData:   "",
			expectedError:  errors.New(`Yandex.Metrika API unknown error: 418 I'm a teapot\nraw content:\n`),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, fmt.Sprintf("/counter/%d/logrequest/%d/part/%d/download", tc.counter, tc.request, tc.part), r.URL.Path)
				assert.Equal(t, http.MethodGet, r.Method)

				w.WriteHeader(tc.mockStatusCode)
				fmt.Fprint(w, tc.mockResponse)
			}))
			defer server.Close()

			client := NewClient("management", "v1", "test_token", nil)
			client.client.SetBaseURL(server.URL)

			body, err := client.LogRequest.DownloadWithContext(context.Background(), tc.counter, tc.request, tc.part)
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

				if body != nil {
					defer body.Close()
					content, _ := io.ReadAll(body)
					assert.Equal(t, tc.expectedData, string(content))
				} else {
					assert.Equal(t, "", tc.expectedData)
				}
			}
		})
	}
}
