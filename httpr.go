package httpr

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/alecthomas/types/optional"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type Client struct {
	httpClient          *http.Client
	baseURL             optional.Option[string]
	headers             optional.Option[map[string]string]
	interceptors        []Interceptor
	inspect             optional.Option[bool]
	requestBodyHandler  optional.Option[requestBodyHandler]
	responseBodyHandler optional.Option[responseBodyHandler]
	tracer              trace.Tracer
	meter               metric.Meter

	telemetryEnabled bool
}

func NewClient(options ...ClientOption) *Client {
	c := &Client{
		httpClient: &http.Client{},
		tracer:     otel.GetTracerProvider().Tracer("httpr"),
		meter:      otel.GetMeterProvider().Meter("httpr"),
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
	ctx, span := c.tracer.Start(ctx, "SendRequest",
		trace.WithAttributes(
			attribute.String("http.method", method),
			attribute.String("http.url", path),
		),
	)
	defer span.End()

	opts := requestOptions{
		inspect:      c.inspect,
		requestBody:  c.requestBodyHandler,
		responseBody: c.responseBodyHandler,
		headers:      c.headers,
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
	queryParams, ok := opts.queryParams.Get()
	if ok {
		url += "?" + queryParams.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if headers, hok := c.headers.Get(); hok {
		for key, value := range headers {
			req.Header.Add(key, value)
		}
	}

	if headers, hok := opts.headers.Get(); hok {
		for key, value := range headers {
			req.Header.Add(key, value)
		}
	}

	if _, ok := opts.inspect.Get(); ok {
		var bodyBytes []byte
		if req.Body != nil {
			bodyBytes, _ = io.ReadAll(req.Body)
			req.Body.Close()
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		dumpReq, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			return nil, err
		}
		fmt.Printf("Request:\n%s\n", dumpReq)

		// Restore the request body
		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
	}

	httpResponse, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if _, ok := opts.inspect.Get(); ok {
		bodyBytes, err := io.ReadAll(httpResponse.Body)
		if err != nil {
			log.Fatalf("failed to dump response body: %v", err)
		}

		err = httpResponse.Body.Close()
		if err != nil {
			log.Fatalf("failed to close dumped response body: %v", err)
		}

		httpResponse.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		dumpResp, err := httputil.DumpResponse(httpResponse, true)
		if err != nil {
			log.Fatalf("failed to dump response: %v", err)
		}

		fmt.Printf("Response:\n%s\n", dumpResp)

		httpResponse.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	if responseBodyHandler, ok := opts.responseBody.Get(); ok {
		err := responseBodyHandler(httpResponse)
		if err != nil {
			return nil, fmt.Errorf("failed to handle response body: %w", err)
		}
	}

	return httpResponse, nil
}
