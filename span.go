package vayuotel

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Span is a wrapper around trace.Span that provides a more fluent API
type Span struct {
	Span trace.Span
	ctx  context.Context
}

// convertToAttributes converts a map of interface{} values to OpenTelemetry attributes
func convertToAttributes(attributes map[string]interface{}) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, len(attributes))
	for k, v := range attributes {
		switch val := v.(type) {
		case string:
			attrs = append(attrs, StringAttribute(k, val))
		case int:
			attrs = append(attrs, IntAttribute(k, val))
		case int64:
			attrs = append(attrs, Int64Attribute(k, val))
		case float64:
			attrs = append(attrs, Float64Attribute(k, val))
		case bool:
			attrs = append(attrs, BoolAttribute(k, val))
		case time.Time:
			attrs = append(attrs, TimestampAttribute(k, val))
		}
	}
	return attrs
}

// AddAttributes adds attributes to the span and returns the span for chaining
func (s *Span) AddAttributes(attributes map[string]interface{}) *Span {
	attrs := convertToAttributes(attributes)
	s.Span.SetAttributes(attrs...)
	return s
}

// AddEvent adds an event to the span and returns the span for chaining
func (s *Span) AddEvent(name string, attributes ...map[string]interface{}) *Span {
	var attrs []attribute.KeyValue
	if len(attributes) > 0 && attributes[0] != nil {
		attrs = convertToAttributes(attributes[0])
	}
	s.Span.AddEvent(name, trace.WithAttributes(attrs...))
	return s
}

// RecordError records an error on the span and returns the span for chaining
func (s *Span) RecordError(err error) *Span {
	s.Span.RecordError(err)
	s.Span.SetStatus(codes.Error, err.Error())
	return s
}

// End ends the span
func (s *Span) End() {
	s.Span.End()
}

// Context returns the context associated with this span
func (s *Span) Context() context.Context {
	return s.ctx
}

// Start creates a span from the context and returns our wrapper Span
func Start(ctx context.Context, name string, opts ...SpanOption) *Span {
	// Get the current span from the context
	currentSpan := trace.SpanFromContext(ctx)

	// Get the tracer provider from the current span
	tracerProvider := currentSpan.TracerProvider()

	// Always get the tracer name from the context
	tracerName := ctx.Value(tracerNameKey).(string)

	// Get the tracer with the appropriate name
	tracer := tracerProvider.Tracer(tracerName)

	// Create a new child span
	newCtx, span := tracer.Start(ctx, name)

	// Apply options
	for _, opt := range opts {
		opt.Apply(span)
	}

	// Return our wrapper Span
	return &Span{
		Span: span,
		ctx:  newCtx,
	}
}
