package goals

import (
	"github.com/Jeffail/shutdown"
	"github.com/redpanda-data/benthos/v4/public/service"
)

func inputFromConfig(conf *service.ParsedConfig, mgr *service.Resources) (service.BatchInput, error) {
	input := &benthosInput{
		logger:  mgr.Logger(),
		shutSig: shutdown.NewSignaller(),
	}

	var err error

	input.counter, err = conf.FieldInt("counter_id")
	if err != nil {
		return nil, err
	}

	if conf.Contains("token") {
		input.token, err = conf.FieldString("token")
		if err != nil {
			return nil, err
		}
	}

	return input, nil
}
