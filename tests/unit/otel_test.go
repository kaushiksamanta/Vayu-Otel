package unit

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kaushiksamanta/vayu"
	vayuOtel "github.com/kaushiksamanta/vayu-otel"
	"github.com/kaushiksamanta/vayu-otel/tests"
	"go.opentelemetry.io/otel/trace"
)

func TestDefaultSetupOptions(t *testing.T) {
	options := vayuOtel.DefaultSetupOptions()

	// Verify default values
	if options.Config.ServiceName != "vayu-service" {
		t.Errorf("Expected ServiceName to be 'vayu-service', got '%s'", options.Config.ServiceName)
	}
}

func TestSetup(t *testing.T) {
	// Test with nil app
	options := vayuOtel.DefaultSetupOptions()
	options.App = nil

	_, err := vayuOtel.Setup(options)
	if err != vayuOtel.ErrNoApp {
		t.Errorf("Expected ErrNoApp, got %v", err)
	}

	// Test with valid app
	app := vayu.New()
	options.App = app
	options.Config.UseStdout = true // For testing

	integration, err := vayuOtel.Setup(options)
	if err != nil {
		t.Fatalf("Failed to set up integration: %v", err)
	}

	if integration == nil {
		t.Fatal("Expected integration to be non-nil")
	}

	// Cleanup
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := integration.Shutdown(ctx); err != nil {
		t.Errorf("Failed to shut down integration: %v", err)
	}
}

func TestStart(t *testing.T) {
	// Test the Start helper function
	ctx := context.Background()
	provider, err := tests.SetupTestTracer()
	if err != nil {
		t.Fatalf("Failed to setup tracer: %v", err)
	}
	defer provider.Shutdown(context.Background())

	// Set up a context with the tracer name value
	ctx = context.WithValue(ctx, vayuOtel.GetTracerNameKey(), vayuOtel.GetDefaultTracerName())

	// Use the Start helper function
	spanWrapper := vayuOtel.Start(ctx, "test-span")
	if spanWrapper == nil {
		t.Fatal("Expected span wrapper to be non-nil")
	}

	// Get the context from the span wrapper
	newCtx := spanWrapper.Context()

	// Verify the context contains the span
	contextSpan := trace.SpanFromContext(newCtx)
	if contextSpan != spanWrapper.Span {
		t.Error("Expected span from context to match created span")
	}

	// End the span
	spanWrapper.End()
}

func TestAddSpanAttributes(t *testing.T) {
	// Test the AddAttributes method on the Span struct
	provider, err := tests.SetupTestTracer()
	if err != nil {
		t.Fatalf("Failed to setup tracer: %v", err)
	}
	defer provider.Shutdown(context.Background())

	// Create a context with the tracer name
	ctx := context.Background()
	ctx = context.WithValue(ctx, vayuOtel.GetTracerNameKey(), vayuOtel.GetDefaultTracerName())

	// Create a span using the new API
	spanWrapper := vayuOtel.Start(ctx, "attributes-span")
	defer spanWrapper.End()

	// Use the AddAttributes method on the Span struct
	spanWrapper.AddAttributes(map[string]interface{}{
		"test.key":     "value",
		"test.count":   42,
		"test.enabled": true,
	})

	// No direct way to verify attributes in the testing API
	// This test just ensures the method doesn't panic
}

func TestAddSpanEvent(t *testing.T) {
	// Test the AddEvent method on the Span struct
	provider, err := tests.SetupTestTracer()
	if err != nil {
		t.Fatalf("Failed to setup tracer: %v", err)
	}
	defer provider.Shutdown(context.Background())

	// Create a context with the tracer name
	ctx := context.Background()
	ctx = context.WithValue(ctx, vayuOtel.GetTracerNameKey(), vayuOtel.GetDefaultTracerName())

	// Create a span using the new API
	spanWrapper := vayuOtel.Start(ctx, "events-span")
	defer spanWrapper.End()

	// Use the AddEvent method on the Span struct
	spanWrapper.AddEvent("test-event", map[string]interface{}{
		"event.type":  "test",
		"event.count": 1,
	})

	// No direct way to verify events in the testing API
	// This test just ensures the method doesn't panic
}

func TestEndSpan(t *testing.T) {
	// Test the End method on the Span struct
	provider, err := tests.SetupTestTracer()
	if err != nil {
		t.Fatalf("Failed to setup tracer: %v", err)
	}
	defer provider.Shutdown(context.Background())

	// Create a context with the tracer name
	ctx := context.Background()
	ctx = context.WithValue(ctx, vayuOtel.GetTracerNameKey(), vayuOtel.GetDefaultTracerName())

	// Create a span using the new API
	spanWrapper := vayuOtel.Start(ctx, "end-span")

	// Use the End method on the Span struct
	spanWrapper.End()

	// No direct way to verify span ended in the testing API
	// This test just ensures the method doesn't panic
}

func TestRecordSpanError(t *testing.T) {
	// Test the RecordError method on the Span struct
	provider, err := tests.SetupTestTracer()
	if err != nil {
		t.Fatalf("Failed to setup tracer: %v", err)
	}
	defer provider.Shutdown(context.Background())

	// Create a context with the tracer name
	ctx := context.Background()
	ctx = context.WithValue(ctx, vayuOtel.GetTracerNameKey(), vayuOtel.GetDefaultTracerName())

	// Create a span using the new API
	spanWrapper := vayuOtel.Start(ctx, "error-span")
	defer spanWrapper.End()

	// Create a test error
	testErr := errors.New("test error")

	// Use the RecordError method on the Span struct
	spanWrapper.RecordError(testErr)

	// No direct way to verify error was recorded in the testing API
	// This test just ensures the method doesn't panic
}

func TestSpanHierarchy(t *testing.T) {
	// Create a provider for testing
	provider, err := tests.SetupTestTracer()
	if err != nil {
		t.Fatalf("Failed to setup tracer: %v", err)
	}
	defer provider.Shutdown(context.Background())

	// Create a context with the tracer name
	ctx := context.Background()
	ctx = context.WithValue(ctx, vayuOtel.GetTracerNameKey(), vayuOtel.GetDefaultTracerName())

	// Create a parent span
	parentSpan := vayuOtel.Start(ctx, "parent-span")
	// Add attributes using the fluent API
	parentSpan.AddAttributes(map[string]interface{}{
		"level": "root",
	})
	defer parentSpan.End()

	// Get the context from the parent span
	parentCtx := parentSpan.Context()

	// Create a child span
	childSpan := vayuOtel.Start(parentCtx, "child-span")
	// Add attributes and an event using the fluent API
	childSpan.AddAttributes(map[string]interface{}{
		"level": "child",
	}).AddEvent("child-event")
	defer childSpan.End()

	// Get the context from the child span
	childCtx := childSpan.Context()

	// Create a grandchild span
	grandchildSpan := vayuOtel.Start(childCtx, "grandchild-span")
	// Add attributes and an event using the fluent API
	grandchildSpan.AddAttributes(map[string]interface{}{
		"level": "grandchild",
	}).AddEvent("grandchild-event")
	grandchildSpan.End()

	// Add an event to the parent span using the fluent API
	parentSpan.AddEvent("parent-event")

	// This test just verifies that the API works without errors
	// The actual span hierarchy is verified by the OpenTelemetry SDK
}
