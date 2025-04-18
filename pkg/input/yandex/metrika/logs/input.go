package logs

import (
	"context"
	"encoding/csv"
	"io"
	"sync"
	"time"

	"github.com/Jeffail/shutdown"
	"github.com/artemklevtsov/redpanda-connect-plugins/internal/pkg/utils"
	"github.com/artemklevtsov/redpanda-connect-plugins/pkg/input/yandex/metrika/api"
	"github.com/redpanda-data/benthos/v4/public/service"
)

const (
	apiKind    = "management"
	apiVersion = "v1"
)

func init() {
	err := service.RegisterBatchInput(
		"yandex_metrika_logs", inputConfig(),
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
	part      int
	query     *api.LogRequestQuery
	eval      *api.EvalLogRequestResponseEntry
	request   *api.LogRequestResponseEntry
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

	if input.eval == nil {
		input.logger.
			With("counter_id", input.counter).
			Debug("evaluate log request")

		eval, err := input.client.LogRequest.EvalWithContext(ctx, input.counter, input.query)
		if err != nil {
			return service.ErrEndOfInput
		}

		input.eval = &eval.Result
	}

	if !input.eval.IsPossible || input.eval.MaxDays == 0 {
		input.logger.
			With(
				"days", input.eval.MaxDays,
				"possible", input.eval.IsPossible,
			).
			Error("can't evaluate log request")

		return service.ErrEndOfInput
	}

	if input.request == nil {
		input.logger.
			With("counter_id", input.counter).
			Debug("create log request")

		logreq, err := input.client.LogRequest.CreateWithContext(ctx, input.counter, input.query)
		if err != nil {
			return service.ErrEndOfInput
		}

		input.request = &logreq.Request
	}

	if input.request.Status == "created" {
		input.logger.
			With("counter_id", input.counter).
			With("request_id", input.request.RequestID).
			Debug("wait log request")

		for input.request.Status == "created" {
			time.Sleep(10 * time.Second)

			input.logger.
				With("counter_id", input.counter).
				With("request_id", input.request.RequestID).
				Trace("update log request info")

			logreq, err := input.client.LogRequest.GetWithContext(ctx, input.counter, input.request.RequestID)
			if err != nil {
				return service.ErrEndOfInput
			}

			input.request = &logreq.Request
		}
	}

	return nil
}

func (input *benthosInput) ReadBatch(ctx context.Context) (service.MessageBatch, service.AckFunc, error) {
	input.clientMut.Lock()
	defer input.clientMut.Unlock()

	if ctx.Err() != nil || input.done {
		return nil, nil, service.ErrEndOfInput
	}

	if input.request == nil {
		input.logger.
			With("counter_id", input.counter).
			Error("log request not available")

		return nil, nil, service.ErrEndOfInput
	}

	if input.request.Status != "processed" {
		input.logger.
			With("counter_id", input.counter).
			With("request_id", input.request.RequestID).
			With("status", input.request.Status).
			Error("log request process failed")

		return nil, nil, service.ErrEndOfInput
	}

	msgs := make(service.MessageBatch, 0)

	if input.request.Status == "processed" && !input.done {
		input.logger.
			With("counter_id", input.counter).
			With("request_id", input.request.RequestID).
			With("part", input.part).
			Debug("fetch log request")

		query, err := utils.StructToMap(input.request.LogRequestQuery)
		if err != nil {
			return nil, nil, service.ErrEndOfInput
		}

		httpReader, err := input.client.LogRequest.DownloadWithContext(ctx, input.counter, input.request.RequestID, input.part)
		if err != nil {
			return nil, nil, service.ErrEndOfInput
		}

		defer httpReader.Close()

		csvReader := csv.NewReader(httpReader)
		csvReader.Comma = '\t'
		csvReader.LazyQuotes = true

		csvHeader, err := csvReader.Read()
		if err != nil {
			return nil, nil, err
		}

		var rowNumber uint64 = 1

		for {
			csvRow, err := csvReader.Read()
			if err == io.EOF {
				break
			}

			if err != nil {
				return nil, nil, err
			}

			row := make(map[string]any)

			for i, key := range csvHeader {
				key = utils.ProcessKey(key)
				value := utils.ProcessValue(csvHeader[i], csvRow[i])
				row[key] = value
			}

			msg := service.NewMessage(nil)
			msg.SetStructured(row)
			msg.MetaSetMut("counter_id", input.counter)
			msg.MetaSetMut("request_id", input.request.RequestID)
			msg.MetaSetMut("total_parts", len(input.request.Parts))
			msg.MetaSetMut("current_part", input.part+1)
			msg.MetaSetMut("current_row", rowNumber)
			msg.MetaSetMut("query", query)

			msgs = append(msgs, msg)

			rowNumber++
		}

		// if err := g.Wait(); err != nil {
		// 	return nil, nil, service.ErrEndOfInput
		// }

		input.part++

		input.done = input.part == len(input.request.Parts)
	}

	ack := func(context.Context, error) error { return nil }

	return msgs, ack, nil
}

func (input *benthosInput) Close(ctx context.Context) error {
	input.clientMut.Lock()
	defer input.clientMut.Unlock()

	if input.request == nil {
		return nil
	} else {
		logreq, err := input.client.LogRequest.GetWithContext(ctx, input.counter, input.request.RequestID)
		if err != nil {
			return err
		}

		input.request = &logreq.Request
	}

	if input.request.Status == "created" {
		input.logger.
			With("counter_id", input.counter).
			With("request_id", input.request.RequestID).
			With("status", input.request.Status).
			Debug("cancel log request")

		_, err := input.client.LogRequest.CancelWithContext(ctx, input.counter, input.request.RequestID)
		if err != nil {
			return err
		}
	} else {
		input.logger.
			With("counter_id", input.counter).
			With("request_id", input.request.RequestID).
			With("status", input.request.Status).
			Debug("clean log request")

		_, err := input.client.LogRequest.CleanWithContext(ctx, input.counter, input.request.RequestID)
		if err != nil {
			return err
		}
	}

	return nil
}
