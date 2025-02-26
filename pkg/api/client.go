package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/imroc/req/v3"
	"github.com/redpanda-data/benthos/v4/public/service"
)

const (
	// BaseURL is the base URL for Yandex.Metrika API requests.
	APIBaseURL = "https://api-metrika.yandex.com"
)

// Client is a wrapper around req.Client for interacting with the Yandex.Metrika API.
// It provides methods for creating and sending API requests, handling errors, and managing authentication.
type Client struct {
	client     *req.Client
	logger     *service.Logger
	Goal       *GoalService
	StatTable  *StatTableService
	LogRequest *LogRequestService
}

// R creates and returns a new req.Request instance.
// This method is a convenient way to start building a new API request
// with the pre-configured client settings, such as base URL, authentication,
// retry policy, and error handling.
func (c *Client) R() *req.Request {
	return c.client.R()
}

// NewClient creates a new req.Client configured for interacting with the Yandex.Metrika API.
// It sets the base URL, common error result, retry policy, logging, and authentication.
func NewClient(kind, version, token string, logger *service.Logger) *Client {
	baseURL := fmt.Sprintf("%s/%s/%s", APIBaseURL, kind, version)

	httpClient := req.C().
		SetBaseURL(baseURL).
		SetCommonErrorResult(&APIError{}).
		SetCommonRetryCount(3).
		SetCommonRetryBackoffInterval(1*time.Second, 5*time.Second).
		SetCommonRetryCondition(func(resp *req.Response, err error) bool {
			return resp.GetStatusCode() == http.StatusTooManyRequests
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

				resp.Err = err

				return nil
			}

			if !resp.IsSuccessState() {
				// Neither a success response nor a error response, record details to help troubleshooting
				resp.Err = fmt.Errorf("Yandex.Metrika API unknown error: %s\nraw content:\n%s", resp.Status, resp.Dump())

				return nil
			}

			return nil
		})

	if logger != nil {
		httpClient.SetLogger(logger)
	}

	if len(token) > 0 {
		httpClient.SetCommonBearerAuthToken(token)
	}

	c := &Client{
		client: httpClient,
		logger: logger,
	}

	c.Goal = &GoalService{client: c}
	c.LogRequest = &LogRequestService{client: c}
	c.StatTable = &StatTableService{client: c}

	return c
}
