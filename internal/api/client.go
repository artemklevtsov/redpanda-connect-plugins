package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/imroc/req/v3"
	"github.com/redpanda-data/benthos/v4/public/service"
)

const (
	APIBaseURL = "https://api-metrika.yandex.com"
)

func NewClient(kind, version, token string, logger *service.Logger) *req.Client {
	baseURL := fmt.Sprintf("%s/%s/%s", APIBaseURL, kind, version)

	httpClient := req.C().
		SetBaseURL(baseURL).
		SetCommonErrorResult(&APIError{}).
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
			if err, ok := resp.ErrorResult().(*APIError); ok {
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

// APIError API error object
// Source: https://yandex.ru/dev/metrika/doc/api2/management/concept/errors.html#errors__resp
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Reasons []struct {
		ErrorType string `json:"error_type"`
		Message   string `json:"message"`
		Location  string `json:"location"`
	} `json:"errors"`
}

// API error string representation.
func (e APIError) Error() string {
	return fmt.Sprintf("Yandex.Metriika API error %d: %s", e.Code, e.Message)
}

func (e *APIError) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Int("code", e.Code),
		slog.String("message", e.Message),
	)
}
