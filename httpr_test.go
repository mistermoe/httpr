package httpr_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/mistermoe/httpr"

	"github.com/alecthomas/assert/v2"
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

	fmt.Println(capturedOutput)

	assert.NotZero(t, capturedOutput)
	assert.Contains(t, capturedOutput, "Request:")
	assert.Contains(t, capturedOutput, "Response:")

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.NotZero(t, body)
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

// func TestBaseURL(t *testing.T) {
// 	client := httpr.NewClient(httpr.BaseURL("https://jsonplaceholder.typicode.com"))

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
// 		"/posts/1",
// 		httpr.ResponseBodyJSONInto(&post),
// 	)

// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusOK, resp.StatusCode)
// 	assert.NotZero(t, post)
// }

// func TestPost(t *testing.T) {
// 	client := httpr.NewClient(httpr.BaseURL("https://hehe.gov"))

// 	type Post struct {
// 		UserID int    `json:"userId"`
// 		ID     int    `json:"id"`
// 		Title  string `json:"title"`
// 		Body   string `json:"body"`
// 	}

// 	httpmock.Activate()
// 	defer httpmock.DeactivateAndReset()

// 	httpmock.RegisterResponder("POST", "https://hehe.gov/posts", func(req *http.Request) (*http.Response, error) {
// 		var p Post
// 		reqBody, err := io.ReadAll(req.Body)
// 		assert.NoError(t, err)

// 		err = json.Unmarshal(reqBody, &p)
// 		assert.NoError(t, err)

// 		return httpmock.NewBytesResponse(http.StatusCreated, nil), nil
// 	})

// 	post := Post{
// 		UserID: 1,
// 		ID:     1,
// 		Title:  "foo",
// 		Body:   "bar",
// 	}

// 	resp, err := client.SendRequest(
// 		context.Background(),
// 		http.MethodPost,
// 		"/posts",
// 		httpr.RequestBodyJSON(post),
// 	)

// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusCreated, resp.StatusCode)
// }

// func TestResponseErrorBodyInto(t *testing.T) {
// 	type Error struct {
// 		Message string                  `json:"message"`
// 		Code    optional.Option[int]    `json:"code"`
// 		Field   optional.Option[string] `json:"field"`
// 	}

// 	httpmock.Activate()
// 	defer httpmock.DeactivateAndReset()

// 	httpmock.RegisterResponder("POST", "https://hehe.gov/posts", func(*http.Request) (*http.Response, error) {
// 		errResponse := Error{
// 			Message: "bad title",
// 			Code:    optional.Some(1001),
// 			Field:   optional.Some("title"),
// 		}

// 		return httpmock.NewJsonResponse(http.StatusBadRequest, errResponse)
// 	})

// 	type Post struct {
// 		UserID int    `json:"userId"`
// 		ID     int    `json:"id"`
// 		Title  string `json:"title"`
// 		Body   string `json:"body"`
// 	}

// 	client := httpr.NewClient(httpr.BaseURL("https://hehe.gov"))
// 	var successBody Post
// 	var errBody Error

// 	resp, err := client.SendRequest(
// 		context.Background(),
// 		http.MethodPost,
// 		"/posts",
// 		httpr.RequestBodyJSON(Post{Title: "banana phone"}),
// 		httpr.ResponseBodyJSONInto(&successBody),
// 		httpr.ResponseErrorBodyInto(&errBody),
// 	)

// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
// 	assert.NotZero(t, errBody)
// 	assert.Zero(t, successBody)
// }
