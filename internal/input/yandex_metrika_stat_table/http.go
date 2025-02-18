package stat_table

import (
	"fmt"
	"net/http"
	"time"

	"github.com/imroc/req/v3"
	"github.com/redpanda-data/benthos/v4/public/service"
)

const (
	apiBaseURL  = "https://api-metrika.yandex.com"
	apiKind     = "stat"
	apiVersion  = "v1"
	apiEndpoint = "data"
	pageLimit   = 1000
)

func newHttpClient(token string, logger *service.Logger) *req.Client {
	baseURL := fmt.Sprintf("%s/%s/%s", apiBaseURL, apiKind, apiVersion)

	httpClient := req.C().
		SetBaseURL(baseURL).
		SetCommonErrorResult(&apiError{}).
		SetCommonRetryCount(3).
		SetCommonRetryBackoffInterval(1*time.Second, 5*time.Second).
		SetCommonRetryCondition(func(resp *req.Response, err error) bool {
			return err != nil || resp.GetStatusCode() == http.StatusTooManyRequests
		}).
		WrapRoundTripFunc(func(rt req.RoundTripper) req.RoundTripFunc {
			return func(req *req.Request) (resp *req.Response, err error) {
				logger.
					With("url", req.URL.String()).
					Trace("Yandex.Metrika API request")

				return rt.RoundTrip(req)
			}
		}).
		OnAfterResponse(func(client *req.Client, resp *req.Response) error {
			if err, ok := resp.ErrorResult().(*apiError); ok {
				logger.
					With("error", err.Message, "code", err.Code).
					Error("Yandex.Metrika API error")

				return nil
			}

			return nil
		}).
		SetLogger(logger)

	if len(token) > 0 {
		httpClient.SetCommonBearerAuthToken(token)
	}

	return httpClient
}
