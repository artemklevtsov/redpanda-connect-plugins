package api

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/imroc/req/v3"
	"github.com/redpanda-data/benthos/v4/public/service"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	t.Run("default case", func(t *testing.T) {
		client := NewClient("management", "v1", "test_token", &service.Logger{})

		assert.NotNil(t, client)
		assert.NotNil(t, client.client)
		assert.NotNil(t, client.logger)
		assert.NotNil(t, client.Goal)
		assert.NotNil(t, client.LogRequest)
		assert.NotNil(t, client.StatTable)
		assert.Equal(t, "https://api-metrika.yandex.ru/management/v1", client.client.BaseURL)
		assert.Equal(t, "Bearer test_token", client.client.Headers.Get("Authorization"))
	})

	t.Run("nil logger case", func(t *testing.T) {
		client := NewClient("management", "v1", "test_token", nil)

		assert.NotNil(t, client)
		assert.NotNil(t, client.client)
		assert.Nil(t, client.logger) // logger should be nil
		assert.NotNil(t, client.Goal)
		assert.NotNil(t, client.LogRequest)
		assert.NotNil(t, client.StatTable)
		assert.Equal(t, "https://api-metrika.yandex.ru/management/v1", client.client.BaseURL)
		assert.Equal(t, "Bearer test_token", client.client.Headers.Get("Authorization"))
	})

	t.Run("empty token case", func(t *testing.T) {
		client := NewClient("management", "v1", "", &service.Logger{})

		assert.NotNil(t, client)
		assert.NotNil(t, client.client)
		assert.NotNil(t, client.logger)
		assert.NotNil(t, client.Goal)
		assert.NotNil(t, client.LogRequest)
		assert.NotNil(t, client.StatTable)
		assert.Equal(t, "https://api-metrika.yandex.ru/management/v1", client.client.BaseURL)
		// Verify that the token is not set in the headers
		req := client.client.R()
		assert.Empty(t, req.Headers.Get("Authorization"))
	})

	t.Run("different kind and version", func(t *testing.T) {
		client := NewClient("stat", "v2", "test_token", &service.Logger{})

		assert.NotNil(t, client)
		assert.NotNil(t, client.client)
		assert.NotNil(t, client.logger)
		assert.NotNil(t, client.Goal)
		assert.NotNil(t, client.LogRequest)
		assert.NotNil(t, client.StatTable)
		assert.Equal(t, "https://api-metrika.yandex.ru/stat/v2", client.client.BaseURL)
	})
}

func TestClientR(t *testing.T) {
	client := NewClient("management", "v1", "test_token", nil)
	r := client.R()
	assert.NotNil(t, r)
}

func TestClient_ErrorResult(t *testing.T) {
	t.Run("json error", func(t *testing.T) {
		client := NewClient("management", "v1", "test_token", nil)

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"code":1, "message":"some err"}`)
		}))

		defer ts.Close()

		client.client.SetBaseURL(ts.URL)

		var apiErr *APIError

		resp, err := client.client.R().Get("/")
		assert.Error(t, err)
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, 1, apiErr.Code)
		assert.Equal(t, "some err", apiErr.Message)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("unknown error", func(t *testing.T) {
		client := NewClient("management", "v1", "test_token", nil)

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `error message`)
		}))

		defer ts.Close()

		client.client.SetBaseURL(ts.URL)

		resp, err := client.client.R().EnableDump().Get("/")
		assert.Error(t, err)
		assert.Equal(t, errors.New(`Yandex.Metrika API unknown error: 400 Bad Request\nraw content:\nerror message`), err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestClient_TooManyRequest(t *testing.T) {
	client := NewClient("management", "v1", "test_token", nil)

	var counter int

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		counter += 1
		if counter < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))

	defer ts.Close()

	client.client.SetBaseURL(ts.URL)

	_, err := client.client.R().Get("/")
	assert.NoError(t, err)
	assert.Equal(t, counter, 3)
}

func TestClient_WrapRoundTripFunc(t *testing.T) {
	client := NewClient("management", "v1", "test_token", nil)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	defer ts.Close()

	client.client.SetBaseURL(ts.URL)

	_, err := client.client.R().Get("/")
	assert.NoError(t, err)
}

func TestClient_WrapRoundTripFuncError(t *testing.T) {
	client := NewClient("management", "v1", "test_token", nil)
	client.client.WrapRoundTripFunc(func(rt req.RoundTripper) req.RoundTripFunc {
		return func(req *req.Request) (resp *req.Response, err error) {
			return nil, errors.New("some error")
		}
	})

	_, err := client.client.R().Get("/")
	assert.Error(t, err)
	assert.Equal(t, errors.New("some error"), err)
}
