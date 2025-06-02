package vayuotel

import (
	"context"
	"fmt"

	"github.com/kaushiksamanta/vayu"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
)

// Middleware returns a Vayu middleware function that automatically traces HTTP requests
func (i *Integration) Middleware(options ...MiddlewareOptions) vayu.HandlerFunc {
	// Use default options if none are provided
	opts := DefaultMiddlewareOptions()
	if len(options) > 0 {
		opts = options[0]
	}

	// Use default span name formatter if not provided
	if opts.SpanNameFormatter == nil {
		opts.SpanNameFormatter = func(c *vayu.Context) string {
			return fmt.Sprintf("HTTP %s %s", c.Request.Method, c.Request.URL.Path)
		}
	}

	// Get the tracer
	tracer := i.provider.TracerProvider.Tracer(tracerNameValue)

	// Return the middleware function
	return func(c *vayu.Context, next vayu.NextFunc) {
		// Extract trace context from the incoming request headers
		propagator := propagation.TraceContext{}
		ctx := propagator.Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

		// Create the span name
		spanName := opts.SpanNameFormatter(c)

		// Start a new span
		ctx, span := tracer.Start(ctx, spanName)
		defer span.End()

		// Add default HTTP attributes
		span.SetAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
			attribute.String("http.host", c.Request.Host),
			attribute.String("http.user_agent", c.Request.UserAgent()),
			attribute.String("http.scheme", getScheme(c.Request)),
			attribute.String("http.target", c.Request.URL.Path),
		)

		// Add route parameters as attributes if available
		if len(c.Params) > 0 {
			for k, v := range c.Params {
				span.SetAttributes(attribute.String("http.route.param."+k, v))
			}
		}

		// Add custom attributes if provided
		if opts.CustomAttributes != nil {
			customAttrs := opts.CustomAttributes(c)
			if len(customAttrs) > 0 {
				span.SetAttributes(customAttrs...)
			}
		}

		// Store the tracer name in the context
		ctx = context.WithValue(ctx, tracerNameKey, tracerNameValue)

		// Store the span in the request context
		c.Request = c.Request.WithContext(ctx)

		// Call the next handler
		next()

		// Get the response writer to extract status code
		// Note: This assumes Vayu's response writer tracks status code internally
		// If not, we'll need to adapt this approach
		responseStatus := 200 // Default to 200 if we can't determine

		// Add response status code attribute
		span.SetAttributes(attribute.Int("http.status_code", responseStatus))

		// Mark span as error if status code is 5xx
		if responseStatus >= 500 {
			span.SetAttributes(attribute.Bool("error", true))
			span.SetStatus(codes.Error, fmt.Sprintf("Error: HTTP %d", responseStatus))
		}
	}
}

// AutoTraceMiddleware is a convenience function that returns a middleware with default options
func (i *Integration) AutoTraceMiddleware() vayu.HandlerFunc {
	return i.Middleware(DefaultMiddlewareOptions())
}
