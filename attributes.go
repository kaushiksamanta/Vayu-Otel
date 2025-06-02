package vayuotel

import (
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Attribute creation helper functions to avoid direct OpenTelemetry imports

// StringAttribute creates a string attribute
func StringAttribute(key, value string) attribute.KeyValue {
	return attribute.String(key, value)
}

// IntAttribute creates an int attribute
func IntAttribute(key string, value int) attribute.KeyValue {
	return attribute.Int(key, value)
}

// Int64Attribute creates an int64 attribute
func Int64Attribute(key string, value int64) attribute.KeyValue {
	return attribute.Int64(key, value)
}

// Float64Attribute creates a float64 attribute
func Float64Attribute(key string, value float64) attribute.KeyValue {
	return attribute.Float64(key, value)
}

// BoolAttribute creates a bool attribute
func BoolAttribute(key string, value bool) attribute.KeyValue {
	return attribute.Bool(key, value)
}

// TimestampAttribute creates a timestamp attribute
func TimestampAttribute(key string, value time.Time) attribute.KeyValue {
	return attribute.Int64(key, value.UnixNano())
}

// SpanOption is an interface for applying options to a span
type SpanOption interface {
	Apply(span trace.Span)
}

// WithAttributes returns a SpanOption that sets attributes on a span
type WithAttributes []attribute.KeyValue

// Apply implements SpanOption
func (w WithAttributes) Apply(span trace.Span) {
	span.SetAttributes([]attribute.KeyValue(w)...)
}

// WithEvent returns a SpanOption that adds an event to a span
type WithEvent struct {
	Name       string
	Attributes []attribute.KeyValue
}

// Apply implements SpanOption
func (w WithEvent) Apply(span trace.Span) {
	span.AddEvent(w.Name, trace.WithAttributes(w.Attributes...))
}

// WithStringAttribute creates a span option with a string attribute
func WithStringAttribute(key, value string) SpanOption {
	return WithAttributes{StringAttribute(key, value)}
}

// WithIntAttribute creates a span option with an int attribute
func WithIntAttribute(key string, value int) SpanOption {
	return WithAttributes{IntAttribute(key, value)}
}

// WithInt64Attribute creates a span option with an int64 attribute
func WithInt64Attribute(key string, value int64) SpanOption {
	return WithAttributes{Int64Attribute(key, value)}
}

// WithFloat64Attribute creates a span option with a float64 attribute
func WithFloat64Attribute(key string, value float64) SpanOption {
	return WithAttributes{Float64Attribute(key, value)}
}

// WithBoolAttribute creates a span option with a bool attribute
func WithBoolAttribute(key string, value bool) SpanOption {
	return WithAttributes{BoolAttribute(key, value)}
}

// WithTimestampAttribute creates a span option with a timestamp attribute
func WithTimestampAttribute(key string, value time.Time) SpanOption {
	return WithAttributes{TimestampAttribute(key, value)}
}

// WithEventName creates a span option that adds an event with the given name
func WithEventName(name string) SpanOption {
	return WithEvent{Name: name}
}
