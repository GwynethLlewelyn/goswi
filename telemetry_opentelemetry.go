//go:build newrelic

package main

import (
	"context"
	"net/url"
	"os"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials"
)

// Environment variables can override configuration file.

var (
	serviceName  = os.Getenv("SERVICE_NAME")
	collectorURL = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	insecure     = os.Getenv("INSECURE_MODE") // Non-empty means insecure mode toggled.
)

func initTracer() func(context.Context) error {
	secureOption := otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	if len(insecure) > 0 {
		secureOption = otlptracegrpc.WithInsecure()
	}

	exporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			secureOption,
			otlptracegrpc.WithEndpoint(collectorURL),
		),
	)

	if err != nil {
		config.LogFatal(err)
	}
	resources, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attribute.String("service.name", serviceName),
			attribute.String("library.language", "go"),
		),
	)
	if err != nil {
		config.LogError("Could not set resources: ", err)
		// shouldn't we return the error here?...
	}

	otel.SetTracerProvider(
		sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(resources),
		),
	)
	return exporter.Shutdown
}

// Registers elemetry middleware for Gin with OpenTelemetry.
// While you can use OpenTelemetry with anything (including New Relic!), this implementation
// follows the setup for SigNoz.io, a free and open source Datadog/New Relic alternative written in Go.
// See https://signoz.io/blog/opentelemetry-gin/
func initTelemetry() {
	if serviceName == "" {
		serviceName = *config["OTelServiceName"]
	}
	if collectorURL == "" {
		collectorURL = *config["OTelCollectorURL"]
	}

	// Check if `collectorURL` is a valid URL:
	if _, err := url.ParseRequestURI(collectorURL); err != nil {
		config.LogError("OpenTelemetry Collector URL is invalid; telemetry aborted; error was:", err)
		return
	}

	// If unset or empty, by default OpenTelemetry will be in secure mode.
	if insecure == "" {
		insecure = *config["OTelInsecureMode"]
	}

	// check again!
	if serviceName != "" && collectorURL != "" {
		cleanup := initTracer()
		defer cleanup(context.Background())

		// router is a *gin.Engine defined as a global singleton for this application
		router.Use(otelgin.Middleware(serviceName))
	}
}
