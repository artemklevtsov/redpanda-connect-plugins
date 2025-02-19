package goals

import "github.com/redpanda-data/benthos/v4/public/service"

type apiResponse struct {
	Data []apiResponseEntry `json:"goals"`
}

type apiResponseEntry struct {
	Id         uint64              `json:"id"`
	PrevID     uint64              `json:"prev_goal_id,omitempty"`
	Name       string              `json:"name"`
	Type       string              `json:"type"`
	Source     string              `json:"goal_source"`
	Price      float64             `json:"default_price,omitempty"`
	Flag       string              `json:"flag,omitempty"`
	IsFavorite int                 `json:"is_favorite"`
	IsRetarget int                 `json:"is_retargeting"`
	Steps      []apiResponseEntry  `json:"steps,omitempty"`
	Conditions []map[string]string `json:"conditions,omitempty"`
}

func (r *apiResponse) Batch() service.MessageBatch {
	msgs := make(service.MessageBatch, len(r.Data))

	for i, row := range r.Data {
		msg := service.NewMessage(nil)
		msg.SetStructuredMut(row)

		msgs[i] = msg
	}

	return msgs
}
