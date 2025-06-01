package vayuotel

import (
	"context"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Constants for type-safe context keys
const (
	// SpanKey is the key used to store spans in Vayu's context
	SpanKey = "otel.span"

	// TraceIDKey is used to store the trace ID string in context
	TraceIDKey = "otel.trace_id"

	// SpanIDKey is used to store the span ID string in context
	SpanIDKey = "otel.span_id"
)

// ContextWithValue is an interface for context objects with GetValue/SetValue methods
type ContextWithValue interface {
	GetValue(key string) (interface{}, bool)
	SetValue(key string, value interface{})
}

// ContextWithString is an interface for context objects with GetString/SetString methods
type ContextWithString interface {
	GetString(key string) string
	SetString(key string, value string)
}

// ContextWithFloat is an interface for context objects with GetFloat/SetFloat methods
type ContextWithFloat interface {
	GetFloat(key string) float64
	SetFloat(key string, value float64)
}

// ContextWithInt is an interface for context objects with GetInt/SetInt methods
type ContextWithInt interface {
	GetInt(key string) int
	SetInt(key string, value int)
}

// ContextWithBool is an interface for context objects with GetBool/SetBool methods
type ContextWithBool interface {
	GetBool(key string) bool
	SetBool(key string, value bool)
}

// ContextWithStringSlice is an interface for context objects with GetStringSlice/SetStringSlice methods
type ContextWithStringSlice interface {
	GetStringSlice(key string) []string
	SetStringSlice(key string, value []string)
}

// HttpContext is an interface for context objects with HTTP request access
type HttpContext interface {
	ContextWithValue
	Request() *http.Request
}

// JSONMapContext is an interface for contexts with JSONMap method
type JSONMapContext interface {
	JSONMap(statusCode int, data map[string]interface{})
}

// StoreSpan stores a span in the context using type-safe methods
func StoreSpan(c ContextWithValue, span trace.Span) {
	// Store using general interface method for backward compatibility
	c.SetValue(SpanKey, span)

	// Store trace and span IDs as strings for easy access if the context supports it
	stringCtx, ok := c.(ContextWithString)
	if !ok {
		return
	}

	spanContext := span.SpanContext()
	if spanContext.IsValid() {
		stringCtx.SetString(TraceIDKey, spanContext.TraceID().String())
		stringCtx.SetString(SpanIDKey, spanContext.SpanID().String())
	}
}

// GetTraceSpan retrieves a span from the context using type-safe methods
// Note: This still returns interface{} and bool since the span itself isn't type-specific
func GetTraceSpan(c ContextWithValue) (trace.Span, bool) {
	if c == nil {
		return nil, false
	}

	val, exists := c.GetValue(SpanKey)
	if !exists {
		return nil, false
	}

	span, ok := val.(trace.Span)
	return span, ok
}

// GetTraceID returns the current trace ID as a string
func GetTraceID(c ContextWithString) string {
	return c.GetString(TraceIDKey)
}

// GetSpanID returns the current span ID as a string
func GetSpanID(c ContextWithString) string {
	return c.GetString(SpanIDKey)
}

// StoreRequestDuration stores the request duration in milliseconds
func StoreRequestDuration(c ContextWithFloat, duration time.Duration) {
	c.SetFloat("otel.duration_ms", float64(duration.Milliseconds()))
}

// GetRequestDuration retrieves the request duration
func GetRequestDuration(c ContextWithFloat) float64 {
	return c.GetFloat("otel.duration_ms")
}

// StoreRequestAttributes stores request attributes in the context
func StoreRequestAttributes(c interface{}, attrs []attribute.KeyValue) {
	// We need multiple interfaces to store attributes
	boolCtx, hasBool := c.(ContextWithBool)
	intCtx, hasInt := c.(ContextWithInt)
	floatCtx, hasFloat := c.(ContextWithFloat)
	stringCtx, hasString := c.(ContextWithString)
	sliceCtx, hasSlice := c.(ContextWithStringSlice)

	// If we don't have the necessary interfaces, we can't store attributes
	if !hasString || !hasSlice {
		return
	}

	// We store the names of the attributes to make them accessible
	names := make([]string, 0, len(attrs))
	for _, attr := range attrs {
		key := "otel.attr." + string(attr.Key)

		// Store each attribute based on its type
		switch attr.Value.Type() {
		case attribute.BOOL:
			if hasBool {
				boolCtx.SetBool(key, attr.Value.AsBool())
				names = append(names, string(attr.Key))
			}
		case attribute.INT64:
			if hasInt {
				intCtx.SetInt(key, int(attr.Value.AsInt64()))
				names = append(names, string(attr.Key))
			}
		case attribute.FLOAT64:
			if hasFloat {
				floatCtx.SetFloat(key, attr.Value.AsFloat64())
				names = append(names, string(attr.Key))
			}
		case attribute.STRING:
			stringCtx.SetString(key, attr.Value.AsString())
			names = append(names, string(attr.Key))
		default:
			// For complex types, fall back to string representation
			stringCtx.SetString(key, attr.Value.Emit())
			names = append(names, string(attr.Key))
		}
	}

	// Store the list of attribute names for later retrieval
	sliceCtx.SetStringSlice("otel.attributes", names)
}

// GetTracingContext returns the current context with tracing information
// This is useful for passing tracing context to background operations
func GetTracingContext(c interface{}) context.Context {
	// Try to get the request from the context
	httpCtx, ok := c.(HttpContext)
	if !ok {
		return context.Background()
	}

	req := httpCtx.Request()
	if req == nil {
		return context.Background()
	}

	return req.Context()
}

// JSONWithTracing is a helper function that adds tracing information to JSON responses
// It uses type-safe JSONMap method if available
func JSONWithTracing(c interface{}, statusCode int, data map[string]interface{}) {
	// Check if we can get trace information
	stringCtx, ok := c.(ContextWithString)
	if !ok {
		return
	}

	// Check if we can set JSON response
	jsonCtx, ok := c.(JSONMapContext)
	if !ok {
		return
	}

	// Add tracing info to the response if available
	traceID := GetTraceID(stringCtx)
	if traceID != "" {
		// Create a copy of the data map with tracing info
		dataCopy := make(map[string]interface{})
		for k, v := range data {
			dataCopy[k] = v
		}

		// Add tracing metadata
		dataCopy["_tracing"] = map[string]string{
			"trace_id": traceID,
			"span_id":  GetSpanID(stringCtx),
		}

		jsonCtx.JSONMap(statusCode, dataCopy)
		return
	}

	// No tracing info, just send the original data
	jsonCtx.JSONMap(statusCode, data)
}
