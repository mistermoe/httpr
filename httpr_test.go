package httpr_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/mistermoe/httpr"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/metric/metricdata/metricdatatest"

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
	rdr := metric.NewManualReader()
	// Set up test meter provider
	meterProvider := metric.NewMeterProvider(metric.WithReader(rdr))
	otel.SetMeterProvider(meterProvider)

	// Create the Observer
	observer, err := httpr.NewObserver()
	assert.NoError(t, err)

	// Set up mock HTTP server
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "https://example.com/test",
		httpmock.NewStringResponder(200, "OK"))

	// Create client with observer
	client := httpr.NewClient(httpr.Intercept(observer))
	ctx := context.Background()
	// Make a request
	resp, err := client.Get(ctx, "https://example.com/test")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var data metricdata.ResourceMetrics
	err = rdr.Collect(context.Background(), &data)
	assert.NoError(t, err)

	// Assert that metrics have the expected attributes
	metricdatatest.AssertHasAttributes(t, data,
		attribute.String("http.method", "GET"),
		attribute.Int("http.status_code", http.StatusOK),
		attribute.String("http.url", "https://example.com/test"),
		attribute.String("http.host", "example.com"),
	)

	// Assert on specific metrics
	requestCountMetric := getMetric(t, data, "httpr.requests")
	assert.NotZero(t, requestCountMetric, "httpr.requests metric not found")

	sumData, ok := requestCountMetric.Data.(metricdata.Sum[int64])
	assert.True(t, ok, "Expected client.request_count to be Sum[int64]")
	assert.Equal(t, 1, len(sumData.DataPoints), "Expected one data point for client.request_count")
	assert.Equal(t, int64(1), sumData.DataPoints[0].Value, "Expected client.request_count to be 1")

	roundtripMetric := getMetric(t, data, "httpr.roundtrip")
	assert.NotZero(t, roundtripMetric, "httpr.roundtrip metric not found")

	histogramData, ok := roundtripMetric.Data.(metricdata.Histogram[int64])
	assert.True(t, ok, "Expected httpr.roundtrip to be a Histogram")
	assert.Equal(t, 1, len(histogramData.DataPoints), "Expected one data point for httpr.roundtrip")
}

func getMetric(t *testing.T, rm metricdata.ResourceMetrics, name string) *metricdata.Metrics {
	t.Helper()
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == name {
				return &m
			}
		}
	}
	t.Logf("Metric %s not found", name)
	return nil
}
