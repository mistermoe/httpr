package httpr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/alecthomas/types/optional"
)

type ClientOption interface {
	Client(*Client)
}

type RequestOption interface {
	Request(*requestOptions)
}

type Option interface {
	ClientOption
	RequestOption
}

type requestOptions struct {
	requestBody  optional.Option[requestBodyHandler]
	responseBody optional.Option[responseBodyHandler]
	queryParams  optional.Option[url.Values]
	headers      map[string]string
	interceptors []Interceptor
}

type baseURLOption string

func (b baseURLOption) Client(c *Client) {
	c.baseURL = optional.Some(string(b))
}

func BaseURL(baseURL string) ClientOption {
	return baseURLOption(baseURL)
}

type httpClientOption http.Client

func (h httpClientOption) Client(c *Client) {
	httpClient := http.Client(h)
	c.httpClient = &httpClient
}

func HTTPClient(h http.Client) ClientOption {
	return httpClientOption(h)
}

type headerOption struct {
	key, value string
}

func (h headerOption) Client(c *Client) {
	if c.headers == nil {
		c.headers = make(map[string]string)
	}

	c.headers[h.key] = h.value
}

func (h headerOption) Request(r *requestOptions) {
	if r.headers == nil {
		r.headers = make(map[string]string)
	}

	r.headers[h.key] = h.value
}

// Header creates a new Option for setting headers.
func Header(key, value string) Option {
	return headerOption{key, value}
}

type queryParamOption struct {
	key, value string
}

func (q queryParamOption) Request(r *requestOptions) {
	queryParams, ok := r.queryParams.Get()
	if ok {
		queryParams.Add(q.key, q.value)
	} else {
		queryParams = url.Values{}
		queryParams.Add(q.key, q.value)
		r.queryParams = optional.Some(queryParams)
	}
}

func QueryParam(key, value string) RequestOption {
	return queryParamOption{key, value}
}

func Inspect() Option {
	return Intercept(Inspector())
}

type timeoutOption time.Duration

func (t timeoutOption) Client(c *Client) {
	if c.httpClient != nil {
		c.httpClient.Timeout = time.Duration(t)
	} else {
		c.httpClient = &http.Client{
			Timeout: time.Duration(t),
		}
	}
}

func Timeout(timeout time.Duration) ClientOption {
	return timeoutOption(timeout)
}

type interceptOption struct {
	Interceptor
}

func (i interceptOption) Client(c *Client) {
	c.interceptors = append(c.interceptors, i.Interceptor)
}

func (i interceptOption) Request(r *requestOptions) {
	r.interceptors = append(r.interceptors, i.Interceptor)
}

func Intercept(i Interceptor) Option {
	return interceptOption{i}
}

type requestBodyHandler func() (io.Reader, string, error)

type requestBodyOption struct {
	handler requestBodyHandler
}

func (r requestBodyOption) Request(opts *requestOptions) {
	opts.requestBody = optional.Some(r.handler)
}

func (r requestBodyOption) Client(c *Client) {
	c.requestBodyHandler = optional.Some(r.handler)
}

func RequestBody(contentType string, bodyFunc func() (io.Reader, error)) Option {
	return requestBodyOption{
		handler: func() (io.Reader, string, error) {
			body, err := bodyFunc()
			return body, contentType, err
		},
	}
}

// RequestBodyJSON json marshals whatever is passed in and sets the content type to application/json.
func RequestBodyJSON(body any) Option {
	return RequestBody("application/json", func() (io.Reader, error) {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		return bytes.NewReader(b), nil
	})
}

// RequestBodyString sets the content type to text/plain.
func RequestBodyString(body string) Option {
	return RequestBody("text/plain", func() (io.Reader, error) {
		return strings.NewReader(body), nil
	})
}

// RequestBodyForm sets the content type to application/x-www-form-urlencoded.
func RequestBodyForm(data url.Values) Option {
	return RequestBody("application/x-www-form-urlencoded", func() (io.Reader, error) {
		return strings.NewReader(data.Encode()), nil
	})
}

// RequestBodyBytes sets the content type to application/octet-stream.
func RequestBodyBytes(contentType string, body []byte) Option {
	return RequestBody(contentType, func() (io.Reader, error) {
		return bytes.NewReader(body), nil
	})
}

// RequestBodyStream sets the content type to the provided value. good for when you have a stream of data.
func RequestBodyStream(contentType string, body io.Reader) Option {
	return RequestBody(contentType, func() (io.Reader, error) {
		return body, nil
	})
}

type responseBodyHandler func(resp *http.Response) error

type responseHandlerOption struct {
	handler responseBodyHandler
}

func (r responseHandlerOption) Request(opts *requestOptions) {
	opts.responseBody = optional.Some(r.handler)
}

func (r responseHandlerOption) Client(c *Client) {
	c.responseBodyHandler = optional.Some(r.handler)
}

func ResponseBodyJSON(successBody any, errBody any) Option {
	return responseHandlerOption{handler: func(resp *http.Response) error {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}
		defer resp.Body.Close()

		var target any
		if resp.StatusCode >= http.StatusBadRequest {
			target = errBody
		} else {
			target = successBody
		}

		if target != nil {
			if err := json.Unmarshal(body, target); err != nil {
				return fmt.Errorf("failed to unmarshal %d response body: %w", resp.StatusCode, err)
			}
		}

		return nil
	}}
}

func ResponseBodyString(dest *string) Option {
	handler := func(resp *http.Response) error {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		defer resp.Body.Close()

		*dest = string(body)
		return nil
	}

	return responseHandlerOption{handler: handler}
}

func ResponseBodyBytes(dest *[]byte) Option {
	handler := func(resp *http.Response) error {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		defer resp.Body.Close()

		*dest = body
		return nil
	}

	return responseHandlerOption{handler: handler}
}

func ResponseBody(handler responseBodyHandler) Option {
	return responseHandlerOption{handler: handler}
}
