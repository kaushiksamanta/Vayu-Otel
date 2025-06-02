package unit

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kaushiksamanta/vayu"
	vayuOtel "github.com/kaushiksamanta/vayu-otel"
	"github.com/kaushiksamanta/vayu-otel/tests"
	"go.opentelemetry.io/otel/attribute"
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
	newCtx, span := vayuOtel.Start(ctx, "test-span")
	if span == nil {
		t.Fatal("Expected span to be non-nil")
	}

	// Verify the context contains the span
	contextSpan := trace.SpanFromContext(newCtx)
	if contextSpan != span {
		t.Error("Expected span from context to match created span")
	}

	// End the span
	span.End()
}

func TestAddSpanAttributes(t *testing.T) {
	// Test the AddSpanAttributes helper function
	provider, err := tests.SetupTestTracer()
	if err != nil {
		t.Fatalf("Failed to setup tracer: %v", err)
	}
	defer provider.Shutdown(context.Background())

	_, span := provider.Tracer("attributes-tracer").Start(context.Background(), "attributes-span")
	defer span.End()

	// Use the AddSpanAttributes helper function
	vayuOtel.AddSpanAttributes(span,
		attribute.String("test.key", "value"),
		attribute.Int("test.count", 42),
		attribute.Bool("test.enabled", true),
	)

	// No direct way to verify attributes in the testing API
	// This test just ensures the function doesn't panic
}

func TestAddSpanEvent(t *testing.T) {
	// Test the AddSpanEvent helper function
	provider, err := tests.SetupTestTracer()
	if err != nil {
		t.Fatalf("Failed to setup tracer: %v", err)
	}
	defer provider.Shutdown(context.Background())

	_, span := provider.Tracer("events-tracer").Start(context.Background(), "events-span")
	defer span.End()

	// Use the AddSpanEvent helper function
	vayuOtel.AddSpanEvent(span, "test-event",
		attribute.String("event.type", "test"),
		attribute.Int("event.count", 1),
	)

	// No direct way to verify events in the testing API
	// This test just ensures the function doesn't panic
}

func TestEndSpan(t *testing.T) {
	// Test the EndSpan helper function
	provider, err := tests.SetupTestTracer()
	if err != nil {
		t.Fatalf("Failed to setup tracer: %v", err)
	}
	defer provider.Shutdown(context.Background())

	_, span := provider.Tracer("end-span-tracer").Start(context.Background(), "end-span")

	// Use the EndSpan helper function
	vayuOtel.EndSpan(span)

	// No direct way to verify span ended in the testing API
	// This test just ensures the function doesn't panic
}

func TestRecordSpanError(t *testing.T) {
	// Test the RecordSpanError helper function
	provider, err := tests.SetupTestTracer()
	if err != nil {
		t.Fatalf("Failed to setup tracer: %v", err)
	}
	defer provider.Shutdown(context.Background())

	_, span := provider.Tracer("error-tracer").Start(context.Background(), "error-span")
	defer span.End()

	// Create a test error
	testErr := errors.New("test error")

	// Use the RecordSpanError helper function
	vayuOtel.RecordSpanError(span, testErr)

	// No direct way to verify error was recorded in the testing API
	// This test just ensures the function doesn't panic
}

func TestSpanHierarchy(t *testing.T) {
	// Create a provider for testing
	provider, err := tests.SetupTestTracer()
	if err != nil {
		t.Fatalf("Failed to setup tracer: %v", err)
	}
	defer provider.Shutdown(context.Background())

	// Test creating spans with the new API
	ctx := context.Background()

	// Set the tracer name in the context
	ctx = context.WithValue(ctx, vayuOtel.GetTracerNameKey(), vayuOtel.GetDefaultTracerName())

	// Create a root span
	ctx, rootSpan := vayuOtel.Start(ctx, "root-span",
		attribute.String("level", "root"))
	defer rootSpan.End()

	// Create a child span
	ctx, childSpan := vayuOtel.Start(ctx, "child-span",
		attribute.String("level", "child"))
	defer childSpan.End()

	// Create a grandchild span
	_, grandchildSpan := vayuOtel.Start(ctx, "grandchild-span",
		attribute.String("level", "grandchild"))
	defer grandchildSpan.End()

	// Add events to verify functionality
	vayuOtel.AddSpanEvent(rootSpan, "root-event")
	vayuOtel.AddSpanEvent(childSpan, "child-event")
	vayuOtel.AddSpanEvent(grandchildSpan, "grandchild-event")

	// This test just verifies that the API works without errors
	// The actual span hierarchy is verified by the OpenTelemetry SDK
}
