package stat_table

import (
	"context"
	"strconv"
	"sync"

	"github.com/Jeffail/shutdown"
	"github.com/imroc/req/v3"
	"github.com/redpanda-data/benthos/v4/public/service"
)

func init() {
	err := service.RegisterBatchInput(
		"yandex_metrika_stat_table", inputSpec(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.BatchInput, error) {
			return inputFromParsed(conf, mgr)
		})
	if err != nil {
		panic(err)
	}
}

type benthosInput struct {
	token      string
	fetched    int
	total      int
	formatKeys bool
	query      *query
	client     *req.Client
	logger     *service.Logger
	shutSig    *shutdown.Signaller
	clientMut  sync.Mutex
}

func (input *benthosInput) Connect(ctx context.Context) error {
	input.clientMut.Lock()
	defer input.clientMut.Unlock()

	if input.client != nil {
		return nil
	}

	input.client = newHttpClient(input.token, input.logger)

	return nil
}

func (input *benthosInput) ReadBatch(ctx context.Context) (service.MessageBatch, service.AckFunc, error) {
	input.clientMut.Lock()
	defer input.clientMut.Unlock()

	if input.total > 0 && input.fetched >= input.total {
		return nil, nil, service.ErrEndOfInput
	}

	input.logger.Info("Fetch Yandex.Metrika API data")

	var data apiResponse

	resp, err := input.client.R().
		SetContext(ctx).
		SetQueryParams(input.query.Map()).
		SetQueryParam("offset", strconv.Itoa(input.fetched+1)).
		// SetErrorResult(&errResp).
		SetSuccessResult(&data).
		Get(apiEndpoint)
	if err != nil {
		return nil, nil, err
	}

	if resp.IsErrorState() {
		return nil, nil, service.ErrEndOfInput
	}

	if data.TotalRows == 0 {
		return nil, nil, service.ErrEndOfInput
	}

	if input.total == 0 {
		input.total = data.TotalRows
	}

	input.fetched += len(data.Data)

	msgs := make(service.MessageBatch, 0, len(data.Data))

	for _, row := range data.Rows(input.formatKeys) {
		msg := service.NewMessage(nil)
		msg.SetStructuredMut(row)
		msg.MetaSetMut("total_rows", data.TotalRows)
		msg.MetaSetMut("timezone", data.Query.Timezone)
		msgs = append(msgs, msg)
	}

	ack := func(context.Context, error) error { return nil }

	return msgs, ack, nil
}

func (y *benthosInput) Close(ctx context.Context) error {
	return nil
}
