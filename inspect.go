package httpr

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
)

var _ Interceptor = (*Inspector)(nil)

type Inspector struct{}

func (i Inspector) Before(_ context.Context, _ *Client, req *http.Request) error {
	dumpReq, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return fmt.Errorf("failed to dump request to stdout for inspection: %w", err)
	}

	fmt.Printf("Request:\n%s\n", dumpReq) //nolint:forbidigo // debugging purposes

	return nil
}

func (i Inspector) After(_ context.Context, _ *Client, resp *http.Response) error {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("failed to dump response body: %v", err)
	}

	err = resp.Body.Close()
	if err != nil {
		log.Fatalf("failed to close dumped response body: %v", err)
	}

	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	dumpResp, err := httputil.DumpResponse(resp, true)
	if err != nil {
		log.Fatalf("failed to dump response: %v", err)
	}

	fmt.Printf("Response:\n%s\n", dumpResp) //nolint:forbidigo // debugging purposes

	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	return nil
}
