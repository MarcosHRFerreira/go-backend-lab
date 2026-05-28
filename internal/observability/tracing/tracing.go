// Package tracing configures OpenTelemetry tracing for the application.
package tracing

import (
	"context"
	"io"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
)

const (
	defaultServiceName = "go-tweets"
	defaultEnv         = "development"
	defaultVersion     = "dev"
)

type Config struct {
	Service string
	Env     string
	Version string
	Writer  io.Writer
}

func NewProvider(cfg Config) (*sdktrace.TracerProvider, error) {
	writer := cfg.Writer
	if writer == nil {
		writer = os.Stderr
	}

	exporter, err := stdouttrace.New(
		stdouttrace.WithWriter(writer),
		stdouttrace.WithPrettyPrint(),
	)
	if err != nil {
		return nil, err
	}

	resourceInstance, err := resource.Merge(
		resource.Default(),
		resource.NewSchemaless(
			semconv.ServiceName(defaultString(cfg.Service, defaultServiceName)),
			semconv.DeploymentEnvironmentName(defaultString(cfg.Env, defaultEnv)),
			semconv.ServiceVersion(defaultString(cfg.Version, defaultVersion)),
		),
	)
	if err != nil {
		return nil, err
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resourceInstance),
	)

	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return provider, nil
}

func defaultString(value string, fallback string) string {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return fallback
	}

	return trimmedValue
}

func Shutdown(ctx context.Context, provider *sdktrace.TracerProvider) error {
	if provider == nil {
		return nil
	}

	return provider.Shutdown(ctx)
}
