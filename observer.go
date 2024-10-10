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

var _ Interceptor = (*Observer)(nil)

type startTimeKey struct{}

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

func (o *Observer) Before(_ context.Context, _ *Client, _ *http.Request) error {
	return nil
}

func (o *Observer) After(ctx context.Context, _ *Client, resp *http.Response) error {
	attrs := metric.WithAttributes(
		attribute.String("http.method", resp.Request.Method),
		attribute.Int("http.status_code", resp.StatusCode),
		attribute.String("http.url", resp.Request.URL.String()),
		attribute.String("http.domain", resp.Request.URL.Host),
	)

	o.requestCtr.Add(ctx, 1, attrs)

	if startTime, ok := ctx.Value(startTimeKey{}).(time.Time); ok {
		duration := time.Since(startTime).Milliseconds()
		o.roundtripDuration.Record(ctx, duration, attrs)
	}

	return nil
}
