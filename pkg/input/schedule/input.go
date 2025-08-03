package schedule

import (
	"context"
	"time"

	"github.com/redpanda-data/benthos/v4/public/service"
)

func init() {
	err := service.RegisterBatchInput(
		"schedule",
		inputConfig(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.BatchInput, error) {
			return inputFromConfig(conf)
		})
	if err != nil {
		panic(err)
	}
}

type benthosInput struct {
	child    *service.OwnedInput
	timer    *time.Ticker
	interval time.Duration
}

func (input *benthosInput) Connect(ctx context.Context) error {
	return nil
}

func (input *benthosInput) ReadBatch(ctx context.Context) (service.MessageBatch, service.AckFunc, error) {
	select {
	case <-ctx.Done():
		return nil, nil, service.ErrEndOfInput
	case <-input.timer.C:
		return input.child.ReadBatch(ctx)
	default:
		return nil, func(ctx context.Context, err error) error { return nil }, nil
	}
}

func (input *benthosInput) Close(ctx context.Context) error {
	input.timer.Stop()

	return input.child.Close(ctx)
}
