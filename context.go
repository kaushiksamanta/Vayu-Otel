package vayuotel

// contextKey is a private type for context keys used by the vayuotel package
type contextKey int

// Context keys for storing OpenTelemetry-related values in the request context
const (
	tracerNameKey   contextKey = iota
	tracerNameValue string     = "vayu-http"
)

// GetTracerNameKey returns the context key used for storing the tracer name
// This is primarily used for testing
func GetTracerNameKey() contextKey {
	return tracerNameKey
}

// GetDefaultTracerName returns the default tracer name used by the middleware
func GetDefaultTracerName() string {
	return tracerNameValue
}
