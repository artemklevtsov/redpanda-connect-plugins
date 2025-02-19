package input

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"cuelang.org/go/pkg/regexp"
	"github.com/Jeffail/shutdown"
	"github.com/iancoleman/strcase"
	"github.com/imroc/req/v3"
	"github.com/redpanda-data/benthos/v4/public/service"
)

const (
	apiURL      = "https://api-metrika.yandex.net/stats/v1"
	apiLimit    = 1000
	apiKind     = "stat"
	apiVersion  = "v1"
	apiEndpoint = "data"
	pageLimit   = 1000
)

func init() {
	err := service.RegisterBatchInput(
		"yandex_metrika_report", inputSpec(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.BatchInput, error) {
			return inputFromParsed(conf, mgr)
		})
	if err != nil {
		panic(err)
	}
}

func inputSpec() *service.ConfigSpec {
	return service.NewConfigSpec().
		Beta().
		Categories("api", "http", "yandex").
		Summary("Creates an input that fetch Yandex.Metrika API report.").
		Fields(
			service.NewIntListField("ids").
				Description("Yandex.Metrika Counter ID").
				Example([]int{44147844, 2215573}),
			service.NewStringListField("metrics").
				Description("A list of metrics.").
				Example([]string{"ym:s:pageviews", "ym:s:visits", "ym:s:users"}),
			service.NewStringField("token").
				Description("Yandex.Metrika API token").
				Optional(),
			service.NewStringListField("dimensions").
				Description("A list of dimensions.").
				Optional(),
			service.NewStringField("filters").
				Description("Segmentation filter.").
				Optional(),
			service.NewStringField("date1").
				Description("Start date of the sample period in YYYY-MM-DD format.").
				Default("6daysAgo").
				Optional(),
			service.NewStringField("date2").
				Description("End date of the sample period in YYYY-MM-DD format.").
				Default("today").
				Optional(),
			service.NewStringField("lang").
				Description("Language.").
				Optional(),
			service.NewStringField("preset").
				Description("Report preset.").
				Optional(),
			service.NewStringListField("sort").
				Description("A list of dimensions and metrics to use for sorting.").
				Optional(),
			service.NewStringField("timezone").
				Description("Time zone in ±hh:mm format within the range of [-23:59; +23:59]").
				Optional(),
			service.NewBoolField("format_keys").
				Description("Format output JSON keys.").
				Optional(),
		)
}

type yandexMetrikaReportInput struct {
	token      string
	fetched    int
	total      int
	formatKeys bool
	query      *yandexMetrikaReportQuery
	client     *req.Client
	logger     *service.Logger
	shutSig    *shutdown.Signaller
	clientMut  sync.Mutex
}

func inputFromParsed(conf *service.ParsedConfig, mgr *service.Resources) (service.BatchInput, error) {
	y := &yandexMetrikaReportInput{
		logger:  mgr.Logger(),
		shutSig: shutdown.NewSignaller(),
		query:   &yandexMetrikaReportQuery{},
	}

	var err error

	y.query.IDs, err = conf.FieldIntList("ids")
	if err != nil {
		return nil, err
	}

	y.query.Metrics, err = conf.FieldStringList("metrics")
	if err != nil {
		return nil, err
	}
	// if len(y.query.Metrics) == 0 {
	// 	return nil, errors.New("metrics not defined")
	// }

	if conf.Contains("token") {
		y.token, err = conf.FieldString("token")
		if err != nil {
			return nil, err
		}
	}

	if conf.Contains("dimensions") {
		y.query.Dimensions, err = conf.FieldStringList("dimensions")
		if err != nil {
			return nil, err
		}
	}

	if conf.Contains("filters") {
		y.query.Filters, err = conf.FieldString("filters")
		if err != nil {
			return nil, err
		}
	}

	if conf.Contains("sort") {
		y.query.Sort, err = conf.FieldStringList("sort")
		if err != nil {
			return nil, err
		}
	}

	if conf.Contains("date1") {
		y.query.Date1, err = conf.FieldString("date1")
		if err != nil {
			return nil, err
		}
	}

	if conf.Contains("date2") {
		y.query.Date2, err = conf.FieldString("date2")
		if err != nil {
			return nil, err
		}
	}

	if conf.Contains("lang") {
		y.query.Lang, err = conf.FieldString("lang")
		if err != nil {
			return nil, err
		}
	}

	if conf.Contains("preset") {
		y.query.Preset, err = conf.FieldString("preset")
		if err != nil {
			return nil, err
		}
	}

	if conf.Contains("timezone") {
		y.query.Timezone, err = conf.FieldString("timezone")
		if err != nil {
			return nil, err
		}
	}

	if conf.Contains("format_keys") {
		y.formatKeys, err = conf.FieldBool("format_keys")
		if err != nil {
			return nil, err
		}
	}

	return y, nil
}

func (y *yandexMetrikaReportInput) Connect(ctx context.Context) error {
	y.clientMut.Lock()
	defer y.clientMut.Unlock()

	if y.client != nil {
		return nil
	}

	y.client = req.C().
		SetLogger(y.logger).
		SetBaseURL(apiURL).
		SetCommonErrorResult(&yandexMetrikaReportError{}).
		WrapRoundTripFunc(func(rt req.RoundTripper) req.RoundTripFunc {
			return func(req *req.Request) (resp *req.Response, err error) {
				y.logger.
					With("url", req.URL.String()).
					Debug("Yandex.Metrika API request")

				return rt.RoundTrip(req)
			}
		}).
		OnAfterResponse(func(client *req.Client, resp *req.Response) error {
			if err, ok := resp.ErrorResult().(*yandexMetrikaReportError); ok {
				y.logger.
					With("error", err.Message, "code", err.Code).
					Error("Yandex.Metrika API error")

				return nil
			}

			return nil
		}).
		SetCommonRetryCount(5).
		SetCommonRetryBackoffInterval(5*time.Second, 60*time.Second).
		AddCommonRetryCondition(func(resp *req.Response, err error) bool {
			return err != nil || resp.GetStatusCode() == http.StatusTooManyRequests
		})

	if len(y.token) > 0 {
		y.client.SetCommonBearerAuthToken(y.token)
	}

	return nil
}

func (y *yandexMetrikaReportInput) ReadBatch(ctx context.Context) (service.MessageBatch, service.AckFunc, error) {
	y.clientMut.Lock()
	defer y.clientMut.Unlock()

	if y.total > 0 && y.fetched >= y.total {
		return nil, nil, service.ErrEndOfInput
	}

	y.logger.Info("Fetch Yandex.Metrika API data")

	var dataResp yandexMetrikaReportResponse
	// var errResp yandexMetrikaReportError

	resp, err := y.client.R().
		SetContext(ctx).
		SetQueryParams(y.query.Map()).
		SetQueryParam("offset", strconv.Itoa(y.fetched+1)).
		// SetErrorResult(&errResp).
		SetSuccessResult(&dataResp).
		Get("data")
	if err != nil {
		return nil, nil, err
	}

	if resp.IsErrorState() {
		return nil, nil, service.ErrEndOfInput
	}

	if y.total == 0 {
		y.total = dataResp.TotalRows
	}

	y.fetched += len(dataResp.Data)

	msgs := make(service.MessageBatch, 0, len(dataResp.Data))

	for _, row := range dataResp.Rows(y.formatKeys) {
		msg := service.NewMessage(nil)
		msg.SetStructuredMut(row)
		msgs = append(msgs, msg)
	}

	ack := func(context.Context, error) error { return nil }

	return msgs, ack, nil
}

func (y *yandexMetrikaReportInput) Close(ctx context.Context) error {
	return nil
}

type yandexMetrikaReportQuery struct {
	IDs        []int    `json:"ids"`
	Date1      string   `json:"date1,omitempty"`
	Date2      string   `json:"date2,omitempty"`
	Dimensions []string `json:"dimensions,omitempty"`
	Metrics    []string `json:"metrics"`
	Sort       []string `json:"sort,omitempty"`
	Filters    string   `json:"filters,omitempty"`
	Lang       string   `json:"lang,omitempty"`
	Preset     string   `json:"preset,omitempty"`
	Timezone   string   `json:"timezone,omitempty"`
}

func (q *yandexMetrikaReportQuery) Map() map[string]string {
	m := make(map[string]string)
	m["ids"] = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(q.IDs)), ","), "[]")
	m["filters"] = q.Filters
	m["date1"] = q.Date1
	m["date2"] = q.Date2
	m["dimensions"] = strings.Join(q.Dimensions, ",")
	m["metrics"] = strings.Join(q.Metrics, ",")
	m["lang"] = q.Lang
	m["preset"] = q.Preset
	m["sort"] = strings.Join(q.Sort, ",")
	m["timezone"] = q.Timezone
	m["limit"] = strconv.Itoa(apiLimit)

	for k := range m {
		if len(m[k]) == 0 {
			delete(m, k)
		}
	}

	return m
}

type yandexMetrikaReportResponse struct {
	Query     *yandexMetrikaReportQuery      `json:"query"`
	Data      []yandexMetrikaReportDataEntry `json:"data"`
	TotalRows int                            `json:"total_rows"`
}

type yandexMetrikaReportDataEntry struct {
	Dimenctions []struct {
		Name string `json:"name"`
		Id   string `json:"id,omitempty"`
	} `json:"dimensions"`
	Metrics []float64 `json:"metrics"`
}

func (r *yandexMetrikaReportResponse) Rows(formayKeys bool) []map[string]any {
	data := make([]map[string]any, 0, len(r.Data))

	for _, e := range r.Data {
		row := make(map[string]any)

		for di, d := range e.Dimenctions {
			k := r.Query.Dimensions[di]
			if formayKeys {
				k = formatKey(k)
			}

			row[k] = d.Name
		}

		for mi, m := range e.Metrics {
			k := r.Query.Metrics[mi]
			if formayKeys {
				k = formatKey(k)
			}

			row[k] = m
		}

		data = append(data, row)
	}

	return data
}

type yandexMetrikaReportError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Errors  []struct {
		ErrorType string `json:"error_type"`
		Message   string `json:"message"`
	} `json:"errors"`
}

// API error string representation.
func (e yandexMetrikaReportError) Error() string {
	return fmt.Sprintf("Yandex.Metriika API error %d: %s", e.Code, e.Message)
}

func formatKey(s string) string {
	s, _ = regexp.ReplaceAll("^ym:.*:", s, "")
	s = strcase.ToSnake(s)

	return s
}
