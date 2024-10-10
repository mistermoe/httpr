package httpr_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/mistermoe/httpr"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"
	"github.com/jarcoal/httpmock"
)

func TestQueryParam(t *testing.T) {
	t.Run("single value", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(http.MethodGet, "https://hehe.gov", func(r *http.Request) (*http.Response, error) {
			query := r.URL.Query()
			assert.Equal(t, "bar", query.Get("foo"))
			assert.Equal(t, "10", query.Get("baz"))

			return httpmock.NewBytesResponse(http.StatusOK, nil), nil
		})

		httpc := httpr.NewClient()

		resp, err := httpc.Get(
			context.Background(),
			"https://hehe.gov",
			httpr.QueryParam("foo", "bar"),
			httpr.QueryParam("baz", "10"),
		)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("multi value", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(http.MethodGet, "https://hehe.gov", func(r *http.Request) (*http.Response, error) {
			query := r.URL.Query()

			queryValues := query["foo"]
			assert.Equal(t, 2, len(queryValues))
			assert.Equal(t, queryValues[0], "bar")
			assert.Equal(t, queryValues[1], "baz")

			assert.Equal(t, "ham", query.Get("bro"))

			return httpmock.NewBytesResponse(http.StatusOK, nil), nil
		})

		httpc := httpr.NewClient()

		resp, err := httpc.Get(
			context.Background(),
			"https://hehe.gov",
			httpr.QueryParam("foo", "bar"),
			httpr.QueryParam("foo", "baz"),
			httpr.QueryParam("bro", "ham"),
		)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestHeaders(t *testing.T) {
	t.Run("default headers", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(http.MethodGet, "https://hehe.gov", func(r *http.Request) (*http.Response, error) {
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Equal(t, "application/json", r.Header.Get("Accept"))

			return httpmock.NewBytesResponse(http.StatusOK, nil), nil
		})

		httpc := httpr.NewClient(
			httpr.Header("Content-Type", "application/json"),
			httpr.Header("Accept", "application/json"),
		)

		resp, err := httpc.Get(context.Background(), "https://hehe.gov")

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("request headers", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(http.MethodGet, "https://hehe.gov", func(r *http.Request) (*http.Response, error) {
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Equal(t, "application/json", r.Header.Get("Accept"))
			assert.Equal(t, "Bearer token", r.Header.Get("Authorization"))

			return httpmock.NewBytesResponse(http.StatusOK, nil), nil
		})

		httpc := httpr.NewClient()

		resp, err := httpc.Get(
			context.Background(),
			"https://hehe.gov",
			httpr.Header("Authorization", "Bearer token"),
			httpr.Header("Content-Type", "application/json"),
			httpr.Header("Accept", "application/json"),
		)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("wombo combo", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(http.MethodGet, "https://hehe.gov", func(r *http.Request) (*http.Response, error) {
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Equal(t, "application/json", r.Header.Get("Accept"))
			assert.Equal(t, "Bearer token", r.Header.Get("Authorization"))
			assert.Equal(t, "1234", r.Header.Get("X-Request-ID"))

			return httpmock.NewBytesResponse(http.StatusOK, nil), nil
		})

		httpc := httpr.NewClient(
			httpr.Header("Content-Type", "application/json"),
			httpr.Header("Accept", "application/json"),
		)

		resp, err := httpc.Get(
			context.Background(),
			"https://hehe.gov",
			httpr.Header("Authorization", "Bearer token"),
			httpr.Header("X-Request-ID", "1234"),
		)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestInspect(t *testing.T) {
	httpc := httpr.NewClient()

	// Capture stdout
	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	resp, err := httpc.Get(context.Background(), "https://jsonplaceholder.typicode.com/posts/1", httpr.Inspect())
	assert.NoError(t, err)

	// restore stdout
	w.Close()
	os.Stdout = oldStdout

	// read captured output
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	assert.NoError(t, err)

	capturedOutput := buf.String()

	assert.NotZero(t, capturedOutput)
	assert.Contains(t, capturedOutput, "Request:")
	assert.Contains(t, capturedOutput, "Response:")

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.NotZero(t, body)
}

func TestBaseURL(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(http.MethodGet, "https://someapi.io", func(_ *http.Request) (*http.Response, error) {
		return httpmock.NewBytesResponse(http.StatusOK, nil), nil
	})

	httpc := httpr.NewClient(
		httpr.BaseURL("https://someapi.io"),
	)

	resp, err := httpc.Get(context.Background(), "")
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRequestBody(t *testing.T) {
	t.Run("RequestBodyJSON", func(t *testing.T) {
		client := httpr.NewClient(httpr.BaseURL("https://hehe.gov"))

		type Post struct {
			UserID int    `json:"userId"`
			ID     int    `json:"id"`
			Title  string `json:"title"`
			Body   string `json:"body"`
		}

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder("POST", "https://hehe.gov/posts", func(req *http.Request) (*http.Response, error) {
			var p Post
			reqBody, err := io.ReadAll(req.Body)
			assert.NoError(t, err)

			contentType := req.Header.Get("Content-Type")
			assert.Equal(t, "application/json", contentType)

			err = json.Unmarshal(reqBody, &p)
			assert.NoError(t, err)

			return httpmock.NewBytesResponse(http.StatusCreated, nil), nil
		})

		post := Post{
			UserID: 1,
			ID:     1,
			Title:  "foo",
			Body:   "bar",
		}

		resp, err := client.SendRequest(
			context.Background(),
			http.MethodPost,
			"/posts",
			httpr.RequestBodyJSON(post),
		)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("RequestBodyString", func(t *testing.T) {
		client := httpr.NewClient(httpr.BaseURL("https://hehe.gov"))

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder("POST", "https://hehe.gov/posts", func(req *http.Request) (*http.Response, error) {
			reqBody, err := io.ReadAll(req.Body)
			assert.NoError(t, err)

			contentType := req.Header.Get("Content-Type")
			assert.Equal(t, "text/plain", contentType)

			assert.Equal(t, "hello world", string(reqBody))

			return httpmock.NewBytesResponse(http.StatusCreated, nil), nil
		})

		resp, err := client.SendRequest(
			context.Background(),
			http.MethodPost,
			"/posts",
			httpr.RequestBodyString("hello world"),
		)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("RequestBodyBytes", func(t *testing.T) {
		client := httpr.NewClient(httpr.BaseURL("https://hehe.gov"))

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder("POST", "https://hehe.gov/posts", func(req *http.Request) (*http.Response, error) {
			reqBody, err := io.ReadAll(req.Body)
			assert.NoError(t, err)
			contentType := req.Header.Get("Content-Type")
			assert.Equal(t, "text/plain", contentType)

			assert.Equal(t, "hello world", string(reqBody))

			return httpmock.NewBytesResponse(http.StatusCreated, nil), nil
		})

		resp, err := client.SendRequest(
			context.Background(),
			http.MethodPost,
			"/posts",
			httpr.RequestBodyBytes("text/plain", []byte("hello world")),
		)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})
}

// func TestStuff(t *testing.T) {
// 	client := httpr.NewClient()

// 	type Post struct {
// 		UserID int    `json:"userId"`
// 		ID     int    `json:"id"`
// 		Title  string `json:"title"`
// 		Body   string `json:"body"`
// 	}

// 	var post Post
// 	resp, err := client.SendRequest(
// 		context.Background(),
// 		http.MethodGet,
// 		"https://jsonplaceholder.typicode.com/posts/1",
// 		httpr.ResponseBodyJSONInto(&post),
// 	)

// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusOK, resp.StatusCode)
// 	assert.NotZero(t, post)
// }

func TestResponseBodyJSON(t *testing.T) {
	t.Run("ResponseBodyJSONSuccess", func(t *testing.T) {
		type Error struct {
			Message string                  `json:"message"`
			Code    optional.Option[int]    `json:"code"`
			Field   optional.Option[string] `json:"field"`
		}

		type Post struct {
			UserID int    `json:"userId"`
			ID     int    `json:"id"`
			Title  string `json:"title"`
			Body   string `json:"body"`
		}

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder("POST", "https://hehe.gov/posts", func(*http.Request) (*http.Response, error) {
			post := Post{
				UserID: 1,
				ID:     10,
				Title:  "Magical Fruit",
				Body:   "Beans",
			}

			return httpmock.NewJsonResponse(http.StatusAccepted, post)
		})

		client := httpr.NewClient(httpr.BaseURL("https://hehe.gov"))
		var successBody Post
		var errBody Error

		resp, err := client.SendRequest(
			context.Background(),
			http.MethodPost,
			"/posts",
			httpr.RequestBodyJSON(Post{Title: "banana phone"}),
			httpr.ResponseBodyJSON(&successBody, &errBody),
		)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusAccepted, resp.StatusCode)
		assert.Zero(t, errBody)
		assert.NotZero(t, successBody)

		assert.Equal(t, "Beans", successBody.Body)
	})

	t.Run("ResponseBodyJSONError", func(t *testing.T) {
		type Error struct {
			Message string                  `json:"message"`
			Code    optional.Option[int]    `json:"code"`
			Field   optional.Option[string] `json:"field"`
		}

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder("POST", "https://hehe.gov/posts", func(*http.Request) (*http.Response, error) {
			errResponse := Error{
				Message: "bad title",
				Code:    optional.Some(1001),
				Field:   optional.Some("title"),
			}

			return httpmock.NewJsonResponse(http.StatusBadRequest, errResponse)
		})

		type Post struct {
			UserID int    `json:"userId"`
			ID     int    `json:"id"`
			Title  string `json:"title"`
			Body   string `json:"body"`
		}

		client := httpr.NewClient(httpr.BaseURL("https://hehe.gov"))
		var successBody Post
		var errBody Error

		resp, err := client.SendRequest(
			context.Background(),
			http.MethodPost,
			"/posts",
			httpr.RequestBodyJSON(Post{Title: "banana phone"}),
			httpr.ResponseBodyJSON(&successBody, &errBody),
		)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		assert.NotZero(t, errBody)
		assert.Zero(t, successBody)
	})
}

func TestResponseBodyString(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "https://hehe.gov/posts", func(*http.Request) (*http.Response, error) {
		return httpmock.NewStringResponse(http.StatusOK, "hello world"), nil
	})

	client := httpr.NewClient(httpr.BaseURL("https://hehe.gov"))
	var dest string

	resp, err := client.SendRequest(
		context.Background(),
		http.MethodGet,
		"/posts",
		httpr.ResponseBodyString(&dest),
	)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "hello world", dest)
}

func TestResponseBodyBytes(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "https://hehe.gov/posts", func(*http.Request) (*http.Response, error) {
		return httpmock.NewBytesResponse(http.StatusOK, []byte("hello world")), nil
	})

	client := httpr.NewClient(httpr.BaseURL("https://hehe.gov"))
	var dest []byte

	resp, err := client.SendRequest(
		context.Background(),
		http.MethodGet,
		"/posts",
		httpr.ResponseBodyBytes(&dest),
	)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "hello world", string(dest))
}

func TestObserver(t *testing.T) {
	ctx := context.Background()
	exporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithInsecure())
	assert.NoError(t, err)

	res, err := resource.Merge(resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL),
	)

	assert.NoError(t, err)

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exporter, metric.WithInterval(time.Second*5))),
		metric.WithResource(res),
	)

	otel.SetMeterProvider(meterProvider)

	observer, err := httpr.NewObserver()
	assert.NoError(t, err)

	httpc := httpr.NewClient(httpr.Intercept(observer))
	resp, err := httpc.Get(ctx, "https://jsonplaceholder.typicode.com/posts/1")
	assert.NoError(t, err)
	assert.NotEqual(t, 0, resp.StatusCode)

	err = exporter.Shutdown(ctx)
	assert.NoError(t, err)
}
