package stat_table

import (
	"context"
	"sync"

	"github.com/Jeffail/shutdown"
	"github.com/artemklevtsov/redpanda-connect-yandex-metrika/pkg/api"
	"github.com/redpanda-data/benthos/v4/public/service"
)

const (
	apiKind    = "stat"
	apiVersion = "v1"
	pageLimit  = 1000
)

func init() {
	err := service.RegisterBatchInput(
		"yandex_metrika_stat_table", inputConfig(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.BatchInput, error) {
			return inputFromConfig(conf, mgr)
		})
	if err != nil {
		panic(err)
	}
}

type benthosInput struct {
	token     string
	fetched   int
	total     int
	query     *api.StatTableQuery
	client    *api.Client
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

	apiClient := api.NewClient(
		apiKind,
		apiVersion,
		input.token,
		input.logger,
	)

	input.client = apiClient

	return nil
}

func (input *benthosInput) ReadBatch(ctx context.Context) (service.MessageBatch, service.AckFunc, error) {
	input.clientMut.Lock()
	defer input.clientMut.Unlock()

	if input.total > 0 && input.fetched >= input.total {
		return nil, nil, service.ErrEndOfInput
	}

	input.query.Offset = input.fetched + 1

	input.logger.Info("Fetch Yandex.Metrika API data")

	var data api.StatTableResponse

	resp, err := input.client.R().
		SetContext(ctx).
		SetQueryParams(input.query.Params()).
		SetSuccessResult(&data).
		Get("data")
	if err != nil {
		return nil, nil, err
	}

	if resp.IsErrorState() {
		return nil, nil, service.ErrEndOfInput
	}

	if data.TotalRows == 0 {
		input.logger.
			Info("response return 0 rows")

		return nil, nil, service.ErrEndOfInput
	}

	if input.total == 0 {
		input.total = data.TotalRows
	}

	input.fetched += len(data.Data)

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
