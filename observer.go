package httpr

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type Observer struct {
	meter             metric.Meter
	requestCtr        metric.Int64Counter
	roundtripDuration metric.Int64Histogram
}

func NewObserver() (*Observer, error) {
	meter := otel.GetMeterProvider().Meter("httpr")

	requestCtr, err := meter.Int64Counter(
		"client.request_count",
		metric.WithDescription("Total number of requests sent"),
	)
	if err != nil {
		return nil, err
	}

	roundtripDuration, err := meter.Int64Histogram(
		"http.client.roundtrip",
		metric.WithDescription("Duration of HTTP requests"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request duration histogram: %w", err)
	}

	return &Observer{
		meter:             meter,
		requestCtr:        requestCtr,
		roundtripDuration: roundtripDuration,
	}, nil
}

func (o *Observer) Handle(ctx context.Context, req *http.Request, next Interceptor) (*http.Response, error) {
	startTime := time.Now()

	// Call next interceptor
	resp, err := next.Handle(ctx, req, nil)
	if err != nil {
		return nil, err
	}

	duration := time.Since(startTime).Milliseconds()

	// After response
	attrs := metric.WithAttributes(
		attribute.String("http.method", req.Method),
		attribute.Int("http.status_code", resp.StatusCode),
		attribute.String("http.url", req.URL.String()),
		attribute.String("http.domain", req.URL.Host),
	)

	o.requestCtr.Add(ctx, 1, attrs)
	o.roundtripDuration.Record(ctx, duration, attrs)

	return resp, nil
}
