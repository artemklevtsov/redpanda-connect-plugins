package goals

import (
	"context"
	"strconv"
	"sync"

	"github.com/Jeffail/shutdown"
	"github.com/artemklevtsov/redpanda-connect-yandex-metrika/pkg/api"
	"github.com/imroc/req/v3"
	"github.com/redpanda-data/benthos/v4/public/service"
)

const (
	apiKind     = "management"
	apiVersion  = "v1"
	apiEndpoint = "counter"
)

func init() {
	err := service.RegisterBatchInput(
		"yandex_metrika_goals", inputSpec(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.BatchInput, error) {
			return inputFromParsed(conf, mgr)
		})
	if err != nil {
		panic(err)
	}
}

type benthosInput struct {
	token     string
	counter   int
	done      bool
	client    *req.Client
	logger    *service.Logger
	shutSig   *shutdown.Signaller
	clientMut sync.Mutex
}

func (input *benthosInput) Connect(ctx context.Context) error {
	input.clientMut.Lock()
	defer input.clientMut.Unlock()

	if input.client != nil {
		return nil
	}

	httpClient := api.NewClient(
		apiKind,
		apiVersion,
		input.token,
		input.logger,
	)

	input.client = httpClient

	return nil
}

func (input *benthosInput) ReadBatch(ctx context.Context) (service.MessageBatch, service.AckFunc, error) {
	input.clientMut.Lock()
	defer input.clientMut.Unlock()

	if input.done {
		return nil, nil, service.ErrEndOfInput
	}

	input.logger.Info("Fetch Yandex.Metrika API data")

	var data apiResponse

	resp, err := input.client.R().
		SetContext(ctx).
		// SetErrorResult(&errResp).
		SetSuccessResult(&data).
		SetPathParam("counter_id", strconv.Itoa(input.counter)).
		Get("counter/{counter_id}/goals")
	if err != nil {
		return nil, nil, err
	}

	if resp.IsErrorState() {
		return nil, nil, service.ErrEndOfInput
	}

	input.done = true

	msgs, err := data.Batch()
	if err != nil {
		return nil, nil, err
	}

	ack := func(context.Context, error) error { return nil }

	return msgs, ack, nil
}

func (input *benthosInput) Close(ctx context.Context) error {
	return nil
}
