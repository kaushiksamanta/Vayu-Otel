package tests

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

// Provider is a simplified version of the Provider in the main package
type Provider struct {
	TracerProvider trace.TracerProvider
	Exporter       sdktrace.SpanExporter
}

// Shutdown cleans up resources used by the provider
func (p *Provider) Shutdown(ctx context.Context) error {
	if p == nil || p.TracerProvider == nil {
		return errors.New("provider is nil")
	}

	// Cast to SDK TracerProvider to access Shutdown
	if sdkProvider, ok := p.TracerProvider.(*sdktrace.TracerProvider); ok {
		return sdkProvider.Shutdown(ctx)
	}

	return errors.New("tracer provider is not an SDK TracerProvider")
}

// Tracer returns a tracer with the given name
func (p *Provider) Tracer(name string) trace.Tracer {
	if p == nil || p.TracerProvider == nil {
		return trace.NewNoopTracerProvider().Tracer("")
	}
	return p.TracerProvider.Tracer(name)
}

// SetupTestTracer creates a test tracer provider that writes to stdout
func SetupTestTracer() (*Provider, error) {
	// Create stdout exporter
	exporter, err := stdouttrace.New(
		stdouttrace.WithPrettyPrint(),
	)
	if err != nil {
		return nil, err
	}

	// Create resource
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String("test-service"),
		semconv.ServiceVersionKey.String("0.1.0"),
	)

	// Create trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// Set as global provider
	otel.SetTracerProvider(tp)

	return &Provider{
		TracerProvider: tp,
		Exporter:       exporter,
	}, nil
}
