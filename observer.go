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
	metricPrefix      string
	requestCtr        metric.Int64Counter
	roundtripDuration metric.Int64Histogram
}

var _ Interceptor = (*Observer)(nil)

type ObserverOption func(*Observer)

func WithMetricPrefix(prefix string) ObserverOption {
	return func(o *Observer) {
		o.metricPrefix = prefix
	}
}

func NewObserver(opts ...ObserverOption) (*Observer, error) {
	o := &Observer{
		metricPrefix: "httpr", // Default prefix
	}

	for _, opt := range opts {
		opt(o)
	}

	o.meter = otel.GetMeterProvider().Meter(o.metricPrefix)

	requestCtr, err := o.meter.Int64Counter(
		fmt.Sprintf("%s.requests", o.metricPrefix),
		metric.WithDescription("Total number of requests sent"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request counter: %w", err)
	}

	roundtripDuration, err := o.meter.Int64Histogram(
		fmt.Sprintf("%s.roundtrip", o.metricPrefix),
		metric.WithDescription("Duration of HTTP requests"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request duration histogram: %w", err)
	}

	o.requestCtr = requestCtr
	o.roundtripDuration = roundtripDuration

	return o, nil
}

func (o *Observer) Handle(ctx context.Context, req *http.Request, next Interceptor) (*http.Response, error) {
	startTime := time.Now()

	// Call next interceptor
	resp, err := next.Handle(ctx, req, nil)

	duration := time.Since(startTime).Milliseconds()

	// Prepare attributes
	attrs := []attribute.KeyValue{
		attribute.String("http.method", req.Method),
		attribute.String("http.url", req.URL.String()),
		attribute.String("http.host", req.URL.Host),
	}

	if err != nil {
		attrs = append(attrs, attribute.Bool("error", true))
	} else {
		attrs = append(attrs, attribute.Int("http.status_code", resp.StatusCode))
	}

	// Record metrics
	o.requestCtr.Add(ctx, 1, metric.WithAttributes(attrs...))
	o.roundtripDuration.Record(ctx, duration, metric.WithAttributes(attrs...))

	return resp, err
}
