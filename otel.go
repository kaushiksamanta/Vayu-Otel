// Package vayuotel provides OpenTelemetry integration for the Vayu web framework.
package vayuotel

import (
	"context"

	"github.com/kaushiksamanta/vayu"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Integration provides an easy-to-use integration with the Vayu framework
type Integration struct {
	provider *Provider
	app      *vayu.App
}

// SetupOptions contains the options for setting up the integration
type SetupOptions struct {
	// App is the Vayu application instance
	App *vayu.App

	// Config is the OpenTelemetry configuration
	Config Config
}

// DefaultSetupOptions returns default setup options
func DefaultSetupOptions() SetupOptions {
	return SetupOptions{
		Config: DefaultConfig(),
	}
}

// Setup sets up OpenTelemetry integration with Vayu
func Setup(options SetupOptions) (*Integration, error) {
	if options.App == nil {
		return nil, ErrNoApp
	}

	integration := &Integration{
		app: options.App,
	}

	// Initialize provider
	provider, err := NewProvider(options.Config)
	if err != nil {
		return nil, err
	}
	integration.provider = provider

	return integration, nil
}

// Shutdown gracefully shuts down the OpenTelemetry integration
func (i *Integration) Shutdown(ctx context.Context) error {
	if i.provider != nil {
		return i.provider.Shutdown(ctx)
	}
	return nil
}

// AddSpanAttributes adds attributes to the given span
func AddSpanAttributes(span trace.Span, attributes ...attribute.KeyValue) {
	span.SetAttributes(attributes...)
}

// AddSpanEvent adds an event to the given span
func AddSpanEvent(span trace.Span, name string, attributes ...attribute.KeyValue) {
	span.AddEvent(name, trace.WithAttributes(attributes...))
}

// EndSpan ends the given span
func EndSpan(span trace.Span) {
	span.End()
}

// RecordSpanError records an error on the given span
func RecordSpanError(span trace.Span, err error) {
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

// Start creates a span from the context
func Start(ctx context.Context, name string, attributes ...attribute.KeyValue) (context.Context, trace.Span) {
	// Get the current span from the context
	currentSpan := trace.SpanFromContext(ctx)

	// Get the tracer provider from the current span
	tracerProvider := currentSpan.TracerProvider()

	// Always get the tracer name from the context
	tracerName := ctx.Value(tracerNameKey).(string)

	// Get the tracer with the appropriate name
	tracer := tracerProvider.Tracer(tracerName)

	// Create a new child span
	ctx, span := tracer.Start(ctx, name)

	// Add attributes if provided
	if len(attributes) > 0 {
		span.SetAttributes(attributes...)
	}

	return ctx, span
}
