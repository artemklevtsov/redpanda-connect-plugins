package logs

import (
	"github.com/Jeffail/shutdown"
	"github.com/artemklevtsov/redpanda-connect-yandex-metrika/pkg/api"
	"github.com/artemklevtsov/redpanda-connect-yandex-metrika/pkg/misc"
	"github.com/redpanda-data/benthos/v4/public/service"
)

func inputFromConfig(conf *service.ParsedConfig, mgr *service.Resources) (service.BatchInput, error) {
	input := &benthosInput{
		logger:  mgr.Logger(),
		shutSig: shutdown.NewSignaller(),
		query:   &api.LogRequestQuery{},
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

	if conf.Contains("source") {
		input.query.Source, err = conf.FieldString("source")
		if err != nil {
			return nil, err
		}
	}

	if conf.Contains("date1") {
		date1, err := conf.FieldString("date1")
		if err != nil {
			return nil, err
		}

		input.query.Date1, err = misc.ParseDate(date1)
		if err != nil {
			return nil, err
		}
	}

	if conf.Contains("date2") {
		date2, err := conf.FieldString("date2")
		if err != nil {
			return nil, err
		}

		input.query.Date2, err = misc.ParseDate(date2)
		if err != nil {
			return nil, err
		}
	}

	if conf.Contains("fields") {
		input.query.Fields, err = conf.FieldStringList("fields")
		if err != nil {
			return nil, err
		}
	}

	if conf.Contains("attribution") {
		input.query.Attribution, err = conf.FieldString("attribution")
		if err != nil {
			return nil, err
		}
	}

	return input, nil
}
