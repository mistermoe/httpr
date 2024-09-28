package httpr_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/mistermoe/httpr"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"
	"github.com/jarcoal/httpmock"
)

func TestStuff(t *testing.T) {
	client := httpr.NewClient()

	type Post struct {
		UserID int    `json:"userId"`
		ID     int    `json:"id"`
		Title  string `json:"title"`
		Body   string `json:"body"`
	}

	var post Post
	resp, err := client.SendRequest(
		context.Background(),
		http.MethodGet,
		"https://jsonplaceholder.typicode.com/posts/1",
		httpr.ResponseBodyJSONInto(&post),
	)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotZero(t, post)
}

func TestBaseURL(t *testing.T) {
	client := httpr.NewClient(httpr.BaseURL("https://jsonplaceholder.typicode.com"))

	type Post struct {
		UserID int    `json:"userId"`
		ID     int    `json:"id"`
		Title  string `json:"title"`
		Body   string `json:"body"`
	}

	var post Post
	resp, err := client.SendRequest(
		context.Background(),
		http.MethodGet,
		"/posts/1",
		httpr.ResponseBodyJSONInto(&post),
	)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotZero(t, post)
}

func TestPost(t *testing.T) {
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
}

func TestResponseErrorBodyInto(t *testing.T) {
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
		httpr.ResponseBodyJSONInto(&successBody),
		httpr.ResponseErrorBodyInto(&errBody),
	)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.NotZero(t, errBody)
	assert.Zero(t, successBody)
}
