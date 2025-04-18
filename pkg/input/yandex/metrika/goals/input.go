package goals

import (
	"context"
	"sync"

	"github.com/Jeffail/shutdown"
	"github.com/artemklevtsov/redpanda-connect-plugins/pkg/input/yandex/metrika/api"
	"github.com/redpanda-data/benthos/v4/public/service"
)

const (
	apiKind    = "management"
	apiVersion = "v1"
)

func init() {
	err := service.RegisterBatchInput(
		"yandex_metrika_goals", inputConfig(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.BatchInput, error) {
			return inputFromConfig(conf, mgr)
		})
	if err != nil {
		panic(err)
	}
}

type benthosInput struct {
	token     string
	counter   int
	done      bool
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

	if input.done {
		return nil, nil, service.ErrEndOfInput
	}

	input.logger.Info("Fetch Yandex.Metrika API data")

	data, err := input.client.Goal.GetWithContext(ctx, input.counter)
	if err != nil {
		return nil, nil, service.ErrEndOfInput
	}

	msgs, err := data.Batch()
	if err != nil {
		return nil, nil, err
	}

	input.done = true

	ack := func(context.Context, error) error { return nil }

	return msgs, ack, nil
}

func (input *benthosInput) Close(ctx context.Context) error {
	return nil
}
