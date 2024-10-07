package main

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func main() {
	exporter, err := otlpmetricgrpc.New(context.Background())
	if err != nil {
		panic(err)
	}

	res, err := resource.Merge(resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("yarrrg"),
			semconv.ServiceVersion("1.0"),
		))

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exporter, metric.WithInterval(time.Second*5))),
		metric.WithResource(res),
	)

	otel.SetMeterProvider(meterProvider)

	meter := otel.Meter("test")
	counter, err := meter.Int64Counter("counter")
	if err != nil {
		panic(err)
	}

	num := 0

	for {
		counter.Add(context.Background(), 1)
		time.Sleep(time.Second)
		num++

		if num == 100 {
			break
		}
	}
}
