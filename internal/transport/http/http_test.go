package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBasicClient(t *testing.T) {
	baseURL := "https://example.com"
	client, err := New(baseURL)
	require.NoError(t, err)
	assert.Equal(t, baseURL, client.BaseURL)
}

func TestWithRetryPolicy(t *testing.T) {
	rp := RetryPolicy{Count: 3, Wait: time.Second, MaxWait: 5 * time.Second}
	client, err := New("https://example.com", WithRetryPolicy(rp))
	require.NoError(t, err)

	assert.Equal(t, 3, client.RetryCount)
	assert.Equal(t, time.Second, client.RetryWaitTime)
	assert.Equal(t, 5*time.Second, client.RetryMaxWaitTime)
}

func TestMultipleOpts(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	rp := RetryPolicy{Count: 2, Wait: 200 * time.Millisecond}
	client, err := New(
		ts.URL,
		WithRetryPolicy(rp),
	)
	require.NoError(t, err)

	resp, err := client.R().Get("/")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
}

func TestOptWithError(t *testing.T) {
	errOpt := func(c *resty.Client) error {
		return assert.AnError
	}
	client, err := New("https://example.com", errOpt)
	assert.Nil(t, client)
	assert.ErrorIs(t, err, assert.AnError)
}
