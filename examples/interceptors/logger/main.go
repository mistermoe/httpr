package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/mistermoe/httpr"
)

type RequestLogger struct {
	Logger *log.Logger
}

func (l *RequestLogger) Handle(ctx context.Context, req *http.Request, next httpr.Interceptor) (*http.Response, error) {
	l.Logger.Printf("Sending %s request to %s\n", req.Method, req.URL)
	return next.Handle(ctx, req, next)
}

func main() {
	logger := &RequestLogger{Logger: log.New(os.Stdout, "", 0)}

	client := httpr.NewClient(httpr.Intercept(logger))

	resp, err := client.Get(context.Background(), "https://httpbin.org/get")
	if err != nil {
		fmt.Printf("Error: %v\n", err) //nolint:forbidigo // example
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Response: %s\n", body) //nolint:forbidigo // example
}
