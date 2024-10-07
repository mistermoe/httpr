package httpr

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type OTELInterceptor struct {
	tracer          trace.Tracer
	meter           metric.Meter
	requestCounter  metric.Int64Counter
	requestDuration metric.Float64Histogram
}

func NewOTELInterceptor(tracerName, meterName string) (*OTELInterceptor, error) {
	tracer := otel.GetTracerProvider().Tracer(tracerName)
	meter := otel.GetMeterProvider().Meter(meterName)

	requestCounter, err := meter.Int64Counter(
		"http.client.request_count",
		metric.WithDescription("Total number of requests sent"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request counter: %w", err)
	}

	requestDuration, err := meter.Float64Histogram(
		"http.client.duration",
		metric.WithDescription("Duration of HTTP requests"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request duration histogram: %w", err)
	}

	return &OTELInterceptor{
		tracer:          tracer,
		meter:           meter,
		requestCounter:  requestCounter,
		requestDuration: requestDuration,
	}, nil
}

func (i *OTELInterceptor) Before(_ *Client, req *http.Request) error {
	ctx := req.Context()
	ctx, span := i.tracer.Start(ctx, "HTTP "+req.Method,
		trace.WithAttributes(
			attribute.String("http.method", req.Method),
			attribute.String("http.url", req.URL.String()),
		),
	)

	fmt.Println("span start ctx:", ctx)

	ctx = trace.ContextWithSpan(ctx, span)
	fmt.Println("contextWithSpan", ctx)
	ctx = context.WithValue(ctx, "start_time", time.Now())
	fmt.Println("contextWithValue", ctx)

	req = req.WithContext(ctx)

	return nil
}

func (i *OTELInterceptor) After(_ *Client, resp *http.Response) error {
	ctx := resp.Request.Context()
	// span := trace.SpanFromContext(ctx)
	// defer span.End()

	// Record metrics
	i.requestCounter.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("http.method", resp.Request.Method),
			attribute.Int("http.status_code", resp.StatusCode),
		),
	)

	if startTime, ok := ctx.Value("start_time").(time.Time); ok {
		duration := float64(time.Since(startTime).Milliseconds())
		i.requestDuration.Record(ctx, duration,
			metric.WithAttributes(
				attribute.String("http.method", resp.Request.Method),
				attribute.Int("http.status_code", resp.StatusCode),
			),
		)

		// 	// Add duration to the span
		// span.SetAttributes(attribute.Float64("http.duration_ms", duration))
	}

	// // Add response attributes to the span
	// span.SetAttributes(
	// 	attribute.Int("http.status_code", resp.StatusCode),
	// 	attribute.Int64("http.response_content_length", resp.ContentLength),
	// )

	// if resp.ContentLength > 0 {
	// 	span.SetAttributes(attribute.Int64("http.response_body_size", resp.ContentLength))
	// }

	return nil
}
