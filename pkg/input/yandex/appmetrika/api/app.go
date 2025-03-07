package api

import (
	"context"
	"encoding/json"

	"github.com/redpanda-data/benthos/v4/public/service"
)

type AppService struct {
	client *Client
}

func (s *AppService) Get(counter int) (*AppsResponse, error) {
	return s.GetWithContext(context.Background(), counter)
}

func (s *AppService) GetWithContext(ctx context.Context, counter int) (*AppsResponse, error) {
	var data AppsResponse

	_, err := s.client.R().
		SetContext(ctx).
		SetSuccessResult(&data).
		Get("applications")
	if err != nil {
		return nil, err
	}

	return &data, nil
}

// GoalsResponse represents a response containing a list of goals from the Yandex.Metrika API.
type AppsResponse struct {
	Data []AppsResponseEntry `json:"applications"` // Data is a list of goal entries.
}

// GoalsResponseEntry represents a single goal entry in a GoalsResponse.
type AppsResponseEntry struct {
	Id    uint64 `json:"id"`    // Id is the unique identifier of the goal.
	Name  string `json:"name"`  // Name is the name of the goal.
	Label string `json:"label"` // Label is name of the folder where the application is located in the web interface.
}

// Batch creates a service.MessageBatch from the GoalsResponse.
func (r *AppsResponse) Batch() (service.MessageBatch, error) {
	if r.Data == nil {
		return nil, nil
	}

	msgs := make(service.MessageBatch, len(r.Data))

	for i, row := range r.Data {
		msg := service.NewMessage(nil)

		b, err := json.Marshal(row)
		if err != nil {
			return nil, err
		}

		// can't set struct (only slice or map)
		msg.SetBytes(b)

		msgs[i] = msg
	}

	return msgs, nil
}
