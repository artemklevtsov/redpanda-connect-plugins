package logs

import (
	"context"
	"encoding/csv"
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/Jeffail/shutdown"
	"github.com/artemklevtsov/redpanda-connect-yandex-metrika/pkg/api"
	"github.com/artemklevtsov/redpanda-connect-yandex-metrika/pkg/misc"
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
	query     *apiQuery
	eval      *evalLogRequest
	request   *apiLogRequest
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

	if input.eval == nil {
		input.logger.
			With("counter_id", input.counter).
			Debug("evaluate log request")

		var eval apiEvalLogRequestResponse

		resp, err := input.client.R().
			SetContext(ctx).
			SetPathParam("counter_id", strconv.Itoa(input.counter)).
			SetQueryParams(input.query.Params()).
			SetSuccessResult(&eval).
			Get("counter/{counter_id}/logrequests/evaluate")
		if err != nil {
			return service.ErrEndOfInput
		}

		if resp.IsErrorState() {
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

		var logreq apiLogRequestResponse

		resp, err := input.client.R().
			SetContext(ctx).
			SetPathParam("counter_id", strconv.Itoa(input.counter)).
			SetQueryParams(input.query.Params()).
			SetSuccessResult(&logreq).
			Post("counter/{counter_id}/logrequests")
		if err != nil {
			return service.ErrEndOfInput
		}

		if resp.IsErrorState() {
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

			var logreq apiLogRequestResponse

			resp, err := input.client.R().
				SetContext(ctx).
				SetPathParam("counter_id", strconv.Itoa(input.counter)).
				SetPathParam("request_id", strconv.FormatUint(input.request.RequestID, 10)).
				SetQueryParams(input.query.Params()).
				SetSuccessResult(&logreq).
				Get("counter/{counter_id}/logrequest/{request_id}")
			if err != nil {
				return service.ErrEndOfInput
			}

			if resp.IsErrorState() {
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

		resp, err := input.client.R().
			SetContext(ctx).
			SetPathParam("counter_id", strconv.Itoa(input.counter)).
			SetPathParam("request_id", strconv.FormatUint(input.request.RequestID, 10)).
			SetPathParam("part", strconv.Itoa(input.part)).
			SetQueryParams(input.query.Params()).
			Get("counter/{counter_id}/logrequest/{request_id}/part/{part}/download")
		if err != nil {
			return nil, nil, service.ErrEndOfInput
		}

		if resp.IsErrorState() {
			return nil, nil, service.ErrEndOfInput
		}

		defer resp.Body.Close()

		reader := csv.NewReader(resp.Body)
		reader.Comma = '\t'
		reader.LazyQuotes = true

		header, err := reader.Read()
		if err != nil {
			return nil, nil, err
		}

		var rowNumber uint64 = 1

		for {
			record, err := reader.Read()
			if err == io.EOF {
				break
			}

			if err != nil {
				return nil, nil, err
			}

			row := make(map[string]any)

			for i, field := range header {
				field = misc.ProcessKey(field)
				row[field] = misc.ProcessValue(header[i], record[i])
			}

			msg := service.NewMessage(nil)
			msg.SetStructured(row)
			msg.MetaSetMut("counter_id", input.counter)
			msg.MetaSetMut("request_id", input.request.RequestID)
			msg.MetaSetMut("total_parts", len(input.request.Parts))
			msg.MetaSetMut("current_part", input.part+1)
			msg.MetaSetMut("current_row", rowNumber)
			msg.MetaSetMut("query", input.query.Map())

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
		var logreq apiLogRequestResponse

		resp, err := input.client.R().
			SetContext(ctx).
			SetPathParam("counter_id", strconv.Itoa(input.counter)).
			SetPathParam("request_id", strconv.FormatUint(input.request.RequestID, 10)).
			SetQueryParams(input.query.Params()).
			SetSuccessResult(&logreq).
			Get("counter/{counter_id}/logrequest/{request_id}")
		if err != nil {
			return err
		}

		if resp.IsErrorState() {
			return resp.Err
		}

		input.request = &logreq.Request
	}

	if input.request.Status == "created" {
		input.logger.
			With("counter_id", input.counter).
			With("request_id", input.request.RequestID).
			With("status", input.request.Status).
			Debug("cancel log request")

		resp, err := input.client.R().
			SetContext(ctx).
			SetPathParam("counter_id", strconv.Itoa(input.counter)).
			SetPathParam("request_id", strconv.FormatUint(input.request.RequestID, 10)).
			SetQueryParams(input.query.Params()).
			Post("counter/{counter_id}/logrequest/{request_id}/cancel")
		if err != nil {
			return err
		}

		if resp.IsErrorState() {
			return resp.Err
		}
	} else {
		input.logger.
			With("counter_id", input.counter).
			With("request_id", input.request.RequestID).
			With("status", input.request.Status).
			Debug("clean log request")

		resp, err := input.client.R().
			SetContext(ctx).
			SetPathParam("counter_id", strconv.Itoa(input.counter)).
			SetPathParam("request_id", strconv.FormatUint(input.request.RequestID, 10)).
			SetQueryParams(input.query.Params()).
			Post("counter/{counter_id}/logrequest/{request_id}/clean")
		if err != nil {
			return err
		}

		if resp.IsErrorState() {
			return resp.Err
		}
	}

	return nil
}
