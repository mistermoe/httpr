package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mistermoe/httpr"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

func main() {
	ctx := context.Background()
	exporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to create exporter: %v", err)
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exporter, metric.WithInterval(time.Second*5))),
		metric.WithResource(resource.Default()),
	)

	otel.SetMeterProvider(meterProvider)

	observer, err := httpr.NewObserver()
	if err != nil {
		log.Fatalf("failed to create observer: %v", err)
	}

	httpc := httpr.NewClient(httpr.Intercept(observer))

	resp, err := httpc.Get(context.Background(), "https://httpbin.org/get")
	defer resp.Body.Close()

	if err != nil {
		fmt.Printf("Failed to send request: %v\n", err)
		return
	}

	fmt.Printf("Response status: %s\n", resp.Status)

	err = meterProvider.ForceFlush(ctx)
	if err != nil {
		log.Fatalf("failed to shutdown exporter: %v", err)
	}

	err = meterProvider.Shutdown(ctx)
	if err != nil {
		log.Fatalf("failed to shutdown exporter: %v", err)
	}
}
