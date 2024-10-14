package httpr

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
)

// Inspector is an interceptor that logs the request and response to stdout.
func Inspector() HandleFunc {
	return func(ctx context.Context, req *http.Request, next Interceptor) (*http.Response, error) {
		// Before request
		dumpReq, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			return nil, fmt.Errorf("failed to dump request to stdout for inspection: %w", err)
		}

		fmt.Printf("Request:\n%s\n", dumpReq) //nolint:forbidigo // for debugging

		// Call next interceptor
		resp, err := next.Handle(ctx, req, nil)
		if err != nil {
			return nil, err
		}

		// After response
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to dump response body: %w", err)
		}
		resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		dumpResp, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return nil, fmt.Errorf("failed to dump response: %w", err)
		}
		fmt.Printf("Response:\n%s\n", dumpResp) //nolint:forbidigo // for debugging

		resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		return resp, nil
	}
}
