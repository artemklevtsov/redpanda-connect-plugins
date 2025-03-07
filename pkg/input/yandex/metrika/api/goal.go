package api

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/redpanda-data/benthos/v4/public/service"
)

type GoalService struct {
	client *Client
}

func (s *GoalService) Get(counter int) (*GoalsResponse, error) {
	return s.GetWithContext(context.Background(), counter)
}

func (s *GoalService) GetWithContext(ctx context.Context, counter int) (*GoalsResponse, error) {
	var data GoalsResponse

	_, err := s.client.R().
		SetContext(ctx).
		SetSuccessResult(&data).
		SetPathParam("counter_id", strconv.Itoa(counter)).
		Get("counter/{counter_id}/goals")
	if err != nil {
		return nil, err
	}

	return &data, nil
}

// GoalsResponse represents a response containing a list of goals from the Yandex.Metrika API.
type GoalsResponse struct {
	Data []GoalsResponseEntry `json:"goals"` // Data is a list of goal entries.
}

// GoalsResponseEntry represents a single goal entry in a GoalsResponse.
type GoalsResponseEntry struct {
	Id         uint64               `json:"id"`                      // Id is the unique identifier of the goal.
	PrevID     uint64               `json:"prev_goal_id,omitempty"`  // PrevID is the ID of the previous goal.
	Name       string               `json:"name"`                    // Name is the name of the goal.
	Type       string               `json:"type"`                    // Type is the type of the goal.
	Source     string               `json:"goal_source"`             // Source is the source of the goal.
	Price      float64              `json:"default_price,omitempty"` // Price is the default price associated with the goal.
	Flag       string               `json:"flag,omitempty"`          // Flag is an optional flag associated with the goal.
	IsFavorite int                  `json:"is_favorite"`             // IsFavorite indicates whether the goal is a favorite.
	IsRetarget int                  `json:"is_retargeting"`          // IsRetarget indicates whether the goal is for retargeting.
	Steps      []GoalsResponseEntry `json:"steps,omitempty"`         // Steps is a list of steps associated with the goal.
	Conditions []map[string]string  `json:"conditions,omitempty"`    // Conditions is a list of conditions associated with the goal.
}

// Batch creates a service.MessageBatch from the GoalsResponse.
func (r *GoalsResponse) Batch() (service.MessageBatch, error) {
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
