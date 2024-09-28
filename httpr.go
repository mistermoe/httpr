package httpr

// general purpose http client to be moved out into its own package

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/alecthomas/types/optional"
)

type Client struct {
	httpClient           *http.Client
	baseURL              optional.Option[string]
	headers              optional.Option[map[string]string]
	beforeRequestHandler optional.Option[BeforeRequestHandler]
	afterResponseHandler optional.Option[AfterResponseHandler]
}

type BeforeRequestHandler func(c *Client, req *http.Request) error
type AfterResponseHandler func(c *Client, resp *http.Response) error

func NewClient(options ...ClientOption) *Client {
	c := &Client{
		httpClient: &http.Client{},
	}

	for _, option := range options {
		option(c)
	}

	return c

}

type ClientOption func(*Client)

func BaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = optional.Some(baseURL)
	}
}

func Header(key, value string) ClientOption {
	return func(c *Client) {
		headers, ok := c.headers.Get()
		if ok {
			headers[key] = value
		} else {
			c.headers = optional.Some(map[string]string{key: value})
		}
	}
}

func BeforeRequest(handler BeforeRequestHandler) ClientOption {
	return func(c *Client) {
		c.beforeRequestHandler = optional.Some(handler)
	}
}

func AfterResponse(handler AfterResponseHandler) ClientOption {
	return func(c *Client) {
		c.afterResponseHandler = optional.Some(handler)
	}
}

func Transport(transport http.RoundTripper) ClientOption {
	return func(c *Client) {
		if c.httpClient != nil {
			c.httpClient.Transport = transport
		} else {
			c.httpClient = &http.Client{
				Transport: transport,
			}
		}
	}
}

func Timeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		if c.httpClient != nil {
			c.httpClient.Timeout = timeout
		} else {
			c.httpClient = &http.Client{
				Timeout: timeout,
			}
		}
	}
}

type requestBodyHandler = func(body any) ([]byte, error)
type ResponseBodyHandler = func(resp *http.Response) error

type requestOptions struct {
	requestBody       optional.Option[requestBodyHandler]
	responseBody      optional.Option[ResponseBodyHandler]
	responseErrorBody optional.Option[ResponseBodyHandler]
	queryParams       optional.Option[url.Values]
	headers           optional.Option[map[string]string]
}

type RequestOption func(*requestOptions)

func RequestBodyJSON(body any) RequestOption {
	return func(r *requestOptions) {
		r.requestBody = optional.Some(json.Marshal)
	}
}

func RequestBodyStr(body string) RequestOption {
	return func(r *requestOptions) {
		r.requestBody = optional.Some(func(_ any) ([]byte, error) {
			return []byte(body), nil
		})
	}
}

func RequestBody(body []byte) RequestOption {
	return func(r *requestOptions) {
		r.requestBody = optional.Some(func(_ any) ([]byte, error) {
			return body, nil
		})
	}
}

func RequestQueryParam(key string, value string) RequestOption {
	return func(r *requestOptions) {
		queryParams, ok := r.queryParams.Get()
		if ok {
			queryParams.Add(key, value)
		} else {
			queryParams = url.Values{}
			queryParams.Add(key, value)
			r.queryParams = optional.Some(queryParams)
		}
	}
}

func RequestHeader(key, value string) RequestOption {
	return func(r *requestOptions) {
		headers, ok := r.headers.Get()
		if ok {
			headers[key] = value
		} else {
			r.headers = optional.Some(map[string]string{key: value})
		}
	}
}

func ResponseBodyJSONInto(val any) RequestOption {
	return func(r *requestOptions) {
		handleResponseBody := func(resp *http.Response) error {
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read response body: %w", err)
			}

			if err := json.Unmarshal(bodyBytes, val); err != nil {
				return fmt.Errorf("failed to unmarshal response: %w", err)
			}

			return nil
		}

		r.responseBody = optional.Some(handleResponseBody)
	}
}

func ResponseErrorBodyInto(val any) RequestOption {
	return func(r *requestOptions) {
		handleResponseBody := func(resp *http.Response) error {
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read response body: %w", err)
			}

			if err := json.Unmarshal(bodyBytes, val); err != nil {
				return fmt.Errorf("failed to unmarshal response: %w", err)
			}

			return nil
		}

		r.responseErrorBody = optional.Some(handleResponseBody)
	}
}

func ResponseBodyInto(handler ResponseBodyHandler) RequestOption {
	return func(r *requestOptions) {
		r.responseBody = optional.Some(handler)
	}
}

func (c *Client) Get(ctx context.Context, url string, options ...RequestOption) (*http.Response, error) {
	return c.SendRequest(ctx, http.MethodGet, url, options...)
}

func (c *Client) SendRequest(ctx context.Context, method string, path string, options ...RequestOption) (resp *http.Response, err error) {
	opts := requestOptions{}

	for _, option := range options {
		option(&opts)
	}

	var bodyReader io.Reader
	requestBodyHandler, rbhok := opts.requestBody.Get()
	if rbhok {
		bodyBytes, rerr := requestBodyHandler(nil)
		if rerr != nil {
			return nil, fmt.Errorf("failed to prepare request body: %w", err)
		}

		bodyReader = bytes.NewReader(bodyBytes)
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

	beforeReq, brok := c.beforeRequestHandler.Get()
	if brok {
		brerr := beforeReq(c, req)
		if brerr != nil {
			return nil, fmt.Errorf("request middleware errored: %w", brerr)
		}
	}

	httpResponse, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}

	responseBodyHandler, rbhok := opts.responseBody.Get()
	errorBodyHandler, ebhok := opts.responseErrorBody.Get()

	if rbhok || ebhok {
		defer func() {
			// TODO: revisit and think through a cleaner approach. currently, if close returns an error, but there was already an error, the close error is lost
			closeErr := httpResponse.Body.Close()
			if err != nil {
				return
			}

			if closeErr != nil {
				err = fmt.Errorf("failed to close response body: %w", closeErr)
			}
		}()
	}

	if httpResponse.StatusCode >= http.StatusBadRequest {
		if ebhok {
			err := errorBodyHandler(httpResponse)
			if err != nil {
				return nil, fmt.Errorf("failed to process error body: %w", err)
			}
		}
	} else {
		if rbhok {
			err := responseBodyHandler(httpResponse)
			if err != nil {
				return nil, fmt.Errorf("failed to process response body: %w", err)
			}
		}
	}

	afterResp, ok := c.afterResponseHandler.Get()
	if ok {
		err := afterResp(c, httpResponse)
		if err != nil {
			return nil, fmt.Errorf("response middleware errored: %w", err)
		}
	}

	return httpResponse, nil
}
