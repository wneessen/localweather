// SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package http

import (
	"context"
	"crypto/tls"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"time"

	"github.com/wneessen/localweather/internal/log"
)

// reqParams defines the parameters for an HTTP request, including method, endpoint, query, headers, body, and timeout.
type reqParams struct {
	method   string
	endpoint string
	target   any
	query    url.Values
	headers  map[string]string
	body     io.Reader
	timeout  time.Duration
}

const (
	// DefaultTimeout is the default timeout value for the HTTPClient
	DefaultTimeout = time.Second * 10
)

var (
	// version is the version of the application (will be set at build time)
	version = "dev"
	// UserAgent is the User-Agent that the HTTP client sends with API requests
	UserAgent = fmt.Sprintf("Mozilla/5.0 (%s; %s) localweather/%s (+https://github.com/wneessen/localweather/)",
		runtime.GOOS,
		runtime.GOARCH,
		version,
	)

	ErrNonPointerTarget = errors.New("target must be a non-nil pointer")
)

// Client is a type wrapper for the Go stdlib http.Client and the Config
type Client struct {
	*http.Client
	logger *log.Logger
}

// New returns a new HTTP client
func New(logger *log.Logger) *Client {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}
	httpTransport := &http.Transport{TLSClientConfig: tlsConfig}
	httpClient := &http.Client{
		Timeout:   DefaultTimeout,
		Transport: httpTransport,
	}
	return &Client{httpClient, logger}
}

// Get performs a HTTP GET request for the given URL and json-unmarshals the response
// into target
func (h *Client) Get(ctx context.Context, endpoint string, target any, query url.Values, headers map[string]string) (int, error) {
	params := reqParams{
		method:   http.MethodGet,
		endpoint: endpoint,
		target:   target,
		query:    query,
		headers:  headers,
		timeout:  DefaultTimeout,
	}
	return h.PerformReq(ctx, params)
}

func (h *Client) GetWithTimeout(ctx context.Context, endpoint string, target any, query url.Values, headers map[string]string, timeout time.Duration) (int, error) {
	params := reqParams{
		method:   http.MethodGet,
		endpoint: endpoint,
		target:   target,
		query:    query,
		headers:  headers,
		timeout:  timeout,
	}
	return h.PerformReq(ctx, params)
}

// Post performs a HTTP POST request for the given URL and json-unmarshals the response
// into target
func (h *Client) Post(ctx context.Context, endpoint string, target any, body io.Reader, headers map[string]string) (int, error) {
	params := reqParams{
		method:   http.MethodPost,
		endpoint: endpoint,
		target:   target,
		headers:  headers,
		body:     body,
		timeout:  DefaultTimeout,
	}
	return h.PerformReq(ctx, params)
}

func (h *Client) PostWithTimeout(ctx context.Context, endpoint string, target any, body io.Reader, headers map[string]string, timeout time.Duration) (int, error) {
	params := reqParams{
		method:   http.MethodPost,
		endpoint: endpoint,
		target:   target,
		headers:  headers,
		body:     body,
		timeout:  timeout,
	}
	return h.PerformReq(ctx, params)
}

// PerformReq performs a HTTP GET or POST request for the given URL and timeout and JSON-unmarshals the
// response into target
func (h *Client) PerformReq(ctx context.Context, params reqParams) (int, error) {
	rv := reflect.ValueOf(params.target)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return 0, ErrNonPointerTarget
	}

	ctx, cancel := context.WithTimeout(ctx, params.timeout)
	defer cancel()

	// Prepare URL and query parameters
	reqURL, err := url.Parse(params.endpoint)
	if err != nil {
		return 0, fmt.Errorf("failed to parse URL: %w", err)
	}
	if len(params.query) > 0 {
		reqURL.RawQuery = params.query.Encode()
	}

	// Prepare HTTP request
	request, err := http.NewRequestWithContext(ctx, params.method, reqURL.String(), params.body)
	if err != nil {
		return 0, fmt.Errorf("failed create new HTTP request with context: %w", err)
	}
	request.Header.Set("User-Agent", UserAgent)
	for k, v := range params.headers {
		request.Header.Set(k, v)
	}

	// Execute HTTP request
	response, err := h.Do(request)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return 0, err
		}
		return 0, fmt.Errorf("failed to perform HTTP request: %w", err)
	}
	if response == nil {
		return 0, errors.New("nil response received")
	}
	defer func(body io.ReadCloser) {
		if err = body.Close(); err != nil {
			h.logger.Error("failed to close HTTP request body", log.ErrAttr(err))
		}
	}(response.Body)

	// Unmarshal the JSON API response into target
	if err = json.UnmarshalRead(response.Body, params.target); err != nil {
		return response.StatusCode, fmt.Errorf("failed to decode JSON: %w", err)
	}

	return response.StatusCode, nil
}
