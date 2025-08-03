package schedule

import (
	"time"

	"github.com/redpanda-data/benthos/v4/public/service"
)

func inputFromConfig(conf *service.ParsedConfig) (*benthosInput, error) {
	interval, err := conf.FieldDuration("interval")
	if err != nil {
		return nil, err
	}

	child, err := conf.FieldInput("input")
	if err != nil {
		return nil, err
	}

	input := &benthosInput{
		child:    child,
		timer:    time.NewTicker(interval),
		interval: interval,
	}

	return input, nil
}
