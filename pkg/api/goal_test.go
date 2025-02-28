package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/redpanda-data/benthos/v4/public/service"
	"github.com/stretchr/testify/assert"
)

func TestGoalService_GetWithContext(t *testing.T) {
	testCases := []struct {
		name           string
		mockResponse   string
		mockStatusCode int
		expectedGoals  *GoalsResponse
		expectedError  error
	}{
		{
			name: "Successful Request",
			mockResponse: `{
				"goals": [
					{
						"id": 123,
						"name": "Goal 1",
						"type": "visit",
						"goal_source": "test",
						"is_favorite": 1,
						"is_retargeting": 0
					},
					{
						"id": 124,
						"name": "Goal 2",
						"type": "visit",
						"goal_source": "test",
						"is_favorite": 1,
						"is_retargeting": 0,
						"conditions": [{ "type": "value 1", "url": "url 1" }]
					}
				]
			}`,
			mockStatusCode: http.StatusOK,
			expectedGoals: &GoalsResponse{
				Data: []GoalsResponseEntry{
					{
						Id:         123,
						Name:       "Goal 1",
						Type:       "visit",
						Source:     "test",
						IsFavorite: 1,
						IsRetarget: 0,
					},
					{
						Id:         124,
						Name:       "Goal 2",
						Type:       "visit",
						Source:     "test",
						IsFavorite: 1,
						IsRetarget: 0,
						Conditions: []map[string]string{
							{
								"type": "value 1",
								"url":  "url 1",
							},
						},
					},
				},
			},
			expectedError: nil,
		},
		{
			name:           "Error Response",
			mockResponse:   `{"message": "Something went wrong", "code": 1}`,
			mockStatusCode: http.StatusBadRequest,
			expectedGoals:  nil,
			expectedError: &APIError{
				Message: "Something went wrong",
				Code:    1,
			},
		},
		{
			name:           "Empty Response",
			mockResponse:   `{"goals": []}`,
			mockStatusCode: http.StatusOK,
			expectedGoals: &GoalsResponse{
				Data: []GoalsResponseEntry{},
			},
			expectedError: nil,
		},
		{
			name:           "Invalid JSON Response",
			mockResponse:   `{invalid}`,
			mockStatusCode: http.StatusOK,
			expectedGoals:  nil,
			expectedError:  errors.New(`invalid character 'i' looking for beginning of object key string`),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/counter/1/goals", r.URL.Path)
				w.WriteHeader(tc.mockStatusCode)
				fmt.Fprint(w, tc.mockResponse)
			}))
			defer server.Close()

			// Create a client
			client := NewClient("management", "v1", "test_token", nil)
			client.client.SetBaseURL(server.URL)

			// Call GetWithContext
			goals, err := client.Goal.GetWithContext(context.Background(), 1)

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
				assert.Equal(t, tc.expectedGoals, goals)
			}
		})
	}
}

func TestGoalsResponse_Batch(t *testing.T) {
	testCases := []struct {
		name           string
		goalsResponse  GoalsResponse
		expectedBatch  service.MessageBatch
		expectedError  error
		expectedLength int
	}{
		{
			name: "Successful batch creation",
			goalsResponse: GoalsResponse{
				Data: []GoalsResponseEntry{
					{
						Id:         1,
						Name:       "Goal 1",
						Type:       "visit",
						Source:     "test",
						IsFavorite: 1,
						IsRetarget: 0,
					},
					{
						Id:         2,
						Name:       "Goal 2",
						Type:       "visit",
						Source:     "test",
						IsFavorite: 1,
						IsRetarget: 0,
					},
				},
			},
			expectedBatch: func() service.MessageBatch {
				msg1 := service.NewMessage(nil)
				msg1.SetStructuredMut(
					map[string]any{
						"id":             json.Number("1"),
						"name":           "Goal 1",
						"type":           "visit",
						"goal_source":    "test",
						"is_favorite":    json.Number("1"),
						"is_retargeting": json.Number("0"),
					})

				msg2 := service.NewMessage(nil)
				msg2.SetStructuredMut(map[string]any{
					"id":   json.Number("2"),
					"name": "Goal 2",
					"type": "visit", "goal_source": "test",
					"is_favorite":    json.Number("1"),
					"is_retargeting": json.Number("0"),
				})

				return service.MessageBatch{msg1, msg2}
			}(),
			expectedError:  nil,
			expectedLength: 2,
		},
		{
			name:           "Empty Data",
			goalsResponse:  GoalsResponse{Data: []GoalsResponseEntry{}},
			expectedBatch:  nil,
			expectedError:  nil,
			expectedLength: 0,
		},
		{
			name:           "Nil Data",
			goalsResponse:  GoalsResponse{},
			expectedBatch:  nil,
			expectedError:  nil,
			expectedLength: 0,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			batch, err := tc.goalsResponse.Batch()

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
				}
			}
		})
	}
}
