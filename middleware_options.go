package vayuotel

import (
	"fmt"

	"github.com/kaushiksamanta/vayu"
	"go.opentelemetry.io/otel/attribute"
)

// MiddlewareOptions contains configuration options for the tracing middleware
type MiddlewareOptions struct {
	// SpanNameFormatter is a function that formats the span name for a request
	// If nil, the span name will be "HTTP {method} {path}"
	SpanNameFormatter func(c *vayu.Context) string

	// CustomAttributes is a function that adds custom attributes to the span
	// This is called in addition to the default HTTP attributes
	CustomAttributes func(c *vayu.Context) []attribute.KeyValue
}

// DefaultMiddlewareOptions returns the default options for the tracing middleware
func DefaultMiddlewareOptions() MiddlewareOptions {
	return MiddlewareOptions{
		SpanNameFormatter: func(c *vayu.Context) string {
			return fmt.Sprintf("HTTP %s %s", c.Request.Method, c.Request.URL.Path)
		},
		CustomAttributes: nil,
	}
}

// TraceAllRequests is a convenience function that sets up the integration and returns a middleware
func TraceAllRequests(app *vayu.App, config Config) (*Integration, error) {
	// Set up integration options
	options := DefaultSetupOptions()
	options.App = app
	options.Config = config

	// Initialize OpenTelemetry
	integration, err := Setup(options)
	if err != nil {
		return nil, err
	}

	// Add the middleware to the application
	app.Use(integration.AutoTraceMiddleware())

	return integration, nil
}
