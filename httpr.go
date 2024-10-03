package httpr

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/alecthomas/types/optional"
)

type Client struct {
	httpClient          *http.Client
	baseURL             optional.Option[string]
	headers             optional.Option[map[string]string]
	interceptors        []Interceptor
	requestBodyHandler  optional.Option[requestBodyHandler]
	responseBodyHandler optional.Option[responseBodyHandler]
}

func NewClient(options ...ClientOption) *Client {
	c := &Client{
		httpClient: &http.Client{},
	}

	for _, option := range options {
		option.Client(c)
	}

	return c
}

func (c *Client) Get(ctx context.Context, url string, options ...RequestOption) (*http.Response, error) {
	return c.SendRequest(ctx, http.MethodGet, url, options...)
}

func (c *Client) Post(ctx context.Context, url string, options ...RequestOption) (*http.Response, error) {
	return c.SendRequest(ctx, http.MethodPost, url, options...)
}

func (c *Client) Put(ctx context.Context, url string, options ...RequestOption) (*http.Response, error) {
	return c.SendRequest(ctx, http.MethodPut, url, options...)
}

func (c *Client) Delete(ctx context.Context, url string, options ...RequestOption) (*http.Response, error) {
	return c.SendRequest(ctx, http.MethodDelete, url, options...)
}

func (c *Client) SendRequest(ctx context.Context, method string, path string, options ...RequestOption) (resp *http.Response, err error) {
	opts := requestOptions{
		requestBody:  c.requestBodyHandler,
		responseBody: c.responseBodyHandler,
		headers:      c.headers,
		interceptors: c.interceptors,
	}

	for _, option := range options {
		option.Request(&opts)
	}

	var bodyReader io.Reader
	if requestBodyHandler, ok := opts.requestBody.Get(); ok {
		var contentType string
		var err error

		bodyReader, contentType, err = requestBodyHandler()
		if err != nil {
			return nil, fmt.Errorf("failed to get request body: %w", err)
		}

		if contentType != "" {
			Header("Content-Type", contentType).Request(&opts)
		}
	}

	url := c.baseURL.Default("") + path
	if queryParams, ok := opts.queryParams.Get(); ok {
		url += "?" + queryParams.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if headers, hok := opts.headers.Get(); hok {
		for key, value := range headers {
			req.Header.Add(key, value)
		}
	}

	for _, interceptor := range opts.interceptors {
		err := interceptor.Before(c, req)
		if err != nil {
			return nil, fmt.Errorf("request interceptor errored: %w", err)
		}
	}

	httpResponse, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}

	for _, interceptor := range opts.interceptors {
		err := interceptor.After(c, httpResponse)
		if err != nil {
			return nil, fmt.Errorf("response interceptor errored: %w", err)
		}
	}

	if responseBodyHandler, ok := opts.responseBody.Get(); ok {
		err := responseBodyHandler(httpResponse)
		if err != nil {
			return nil, fmt.Errorf("failed to handle response body: %w", err)
		}
	}

	return httpResponse, nil
}
