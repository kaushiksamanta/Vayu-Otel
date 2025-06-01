// Package vayuotel provides OpenTelemetry integration for the Vayu web framework.
package vayuotel

import (
	"context"

	"github.com/kaushiksamanta/vayu"
	"go.opentelemetry.io/otel"
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

	// EnableTracing determines whether to enable distributed tracing
	EnableTracing bool
}

// DefaultSetupOptions returns default setup options
func DefaultSetupOptions() SetupOptions {
	return SetupOptions{
		Config:        DefaultConfig(),
		EnableTracing: true,
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

// GetTracer returns a tracer with the specified name
func (i *Integration) GetTracer(tracerName string) trace.Tracer {
	if i.provider == nil || i.provider.TracerProvider == nil {
		return NewNoopTracer()
	}
	return i.provider.TracerProvider.Tracer(tracerName)
}

// NewNoopTracer returns a no-op tracer for when tracing is disabled
func NewNoopTracer() trace.Tracer {
	return otel.GetTracerProvider().Tracer("noop")
}

// SpanFromContext retrieves the current span from the context
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// StartSpan starts a new span with the given name and options
func StartSpan(ctx context.Context, tracer trace.Tracer, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return tracer.Start(ctx, name, opts...)
}

// StartSpanWithTracer starts a new span with a tracer of the specified name
func StartSpanWithTracer(ctx context.Context, provider *Provider, tracerName string, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if provider == nil || provider.TracerProvider == nil {
		return ctx, trace.SpanFromContext(ctx)
	}
	tracer := provider.TracerProvider.Tracer(tracerName)
	return tracer.Start(ctx, spanName, opts...)
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
