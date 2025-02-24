package logs

import (
	"github.com/Jeffail/shutdown"
	"github.com/redpanda-data/benthos/v4/public/service"
)

func inputFromConfig(conf *service.ParsedConfig, mgr *service.Resources) (service.BatchInput, error) {
	input := &benthosInput{
		logger:  mgr.Logger(),
		shutSig: shutdown.NewSignaller(),
		query:   &apiQuery{},
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
		input.query.Date1, err = conf.FieldString("date1")
		if err != nil {
			return nil, err
		}
	}

	if conf.Contains("date2") {
		input.query.Date2, err = conf.FieldString("date2")
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
