package api

import (
	"context"
	"io"
	"strconv"

	"github.com/google/go-querystring/query"
)

type LogRequestService struct {
	client *Client
}

func (s *LogRequestService) Eval(counter int, q *LogRequestQuery) (*EvalLogRequestResponse, error) {
	return s.EvalWithContext(context.Background(), counter, q)
}

func (s *LogRequestService) EvalWithContext(ctx context.Context, counter int, q *LogRequestQuery) (*EvalLogRequestResponse, error) {
	var (
		err  error
		eval EvalLogRequestResponse
	)

	values, err := query.Values(q)
	if err != nil {
		return nil, err
	}

	_, err = s.client.R().
		SetContext(ctx).
		SetPathParam("counter_id", strconv.Itoa(counter)).
		SetQueryString(values.Encode()).
		SetSuccessResult(&eval).
		Get("counter/{counter_id}/logrequests/evaluate")
	if err != nil {
		return nil, err
	}

	return &eval, nil
}

func (s *LogRequestService) Create(counter int, q *LogRequestQuery) (*LogRequestResponse, error) {
	return s.CreateWithContext(context.Background(), counter, q)
}

func (s *LogRequestService) CreateWithContext(ctx context.Context, counter int, q *LogRequestQuery) (*LogRequestResponse, error) {
	var (
		err    error
		logreq LogRequestResponse
	)

	values, err := query.Values(q)
	if err != nil {
		return nil, err
	}

	_, err = s.client.R().
		SetContext(ctx).
		SetPathParam("counter_id", strconv.Itoa(counter)).
		SetQueryString(values.Encode()).
		SetSuccessResult(&logreq).
		Post("counter/{counter_id}/logrequests")
	if err != nil {
		return nil, err
	}

	return &logreq, nil
}

func (s *LogRequestService) Get(counter int, request uint64) (*LogRequestResponse, error) {
	return s.GetWithContext(context.Background(), counter, request)
}

func (s *LogRequestService) GetWithContext(ctx context.Context, counter int, request uint64) (*LogRequestResponse, error) {
	var logreq LogRequestResponse

	_, err := s.client.R().
		SetContext(ctx).
		SetPathParam("counter_id", strconv.Itoa(counter)).
		SetPathParam("request_id", strconv.FormatUint(request, 10)).
		SetSuccessResult(&logreq).
		Get("counter/{counter_id}/logrequest/{request_id}")
	if err != nil {
		return nil, err
	}

	return &logreq, nil
}

func (s *LogRequestService) Clean(counter int, request uint64) (*LogRequestResponse, error) {
	return s.CleanWithContext(context.Background(), counter, request)
}

func (s *LogRequestService) CleanWithContext(ctx context.Context, counter int, request uint64) (*LogRequestResponse, error) {
	var logreq LogRequestResponse

	_, err := s.client.R().
		SetContext(ctx).
		SetPathParam("counter_id", strconv.Itoa(counter)).
		SetPathParam("request_id", strconv.FormatUint(request, 10)).
		SetSuccessResult(&logreq).
		Post("counter/{counter_id}/logrequest/{request_id}/clean")
	if err != nil {
		return nil, err
	}

	return &logreq, nil
}

func (s *LogRequestService) Cancel(counter int, request uint64) (*LogRequestResponse, error) {
	return s.CancelWithContext(context.Background(), counter, request)
}

func (s *LogRequestService) CancelWithContext(ctx context.Context, counter int, request uint64) (*LogRequestResponse, error) {
	var logreq LogRequestResponse

	_, err := s.client.R().
		SetContext(ctx).
		SetPathParam("counter_id", strconv.Itoa(counter)).
		SetPathParam("request_id", strconv.FormatUint(request, 10)).
		SetSuccessResult(&logreq).
		Post("counter/{counter_id}/logrequest/{request_id}/cancel")
	if err != nil {
		return nil, err
	}

	return &logreq, nil
}

func (s *LogRequestService) Download(counter int, request uint64, part int) (io.ReadCloser, error) {
	return s.DownloadWithContext(context.Background(), counter, request, part)
}

func (s *LogRequestService) DownloadWithContext(ctx context.Context, counter int, request uint64, part int) (io.ReadCloser, error) {
	resp, err := s.client.R().
		SetContext(ctx).
		SetPathParam("counter_id", strconv.Itoa(counter)).
		SetPathParam("request_id", strconv.FormatUint(request, 10)).
		SetPathParam("part", strconv.Itoa(part)).
		Get("counter/{counter_id}/logrequest/{request_id}/part/{part}/download")
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

// LogRequestQuery represents a query for requesting logs from the Yandex.Metrika API.
type LogRequestQuery struct {
	Source      string   `json:"source" url:"source"`           // Source specifies the source of the logs.
	Date1       string   `json:"date1" url:"date1"`             // Date1 is the start date for the log request in YYYY-MM-DD format.
	Date2       string   `json:"date2" url:"date2"`             // Date2 is the end date for the log request in YYYY-MM-DD format.
	Fields      []string `json:"fields" url:"fields,comma"`     // Fields is a list of fields to include in the log request.
	Attribution string   `json:"attribution" url:"attribution"` // Attribution specifies the attribution model.
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
