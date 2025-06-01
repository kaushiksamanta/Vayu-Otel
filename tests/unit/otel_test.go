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

func TestGetTracer(t *testing.T) {
	// Create a valid app and integration
	app := vayu.New()
	options := vayuOtel.DefaultSetupOptions()
	options.App = app
	options.Config.UseStdout = true

	integration, err := vayuOtel.Setup(options)
	if err != nil {
		t.Fatalf("Failed to set up integration: %v", err)
	}

	// Test getting a tracer with a custom name
	customTracer := integration.GetTracer("custom-tracer-name")
	if customTracer == nil {
		t.Error("Expected custom tracer to be non-nil")
	}

	// Cleanup
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := integration.Shutdown(ctx); err != nil {
		t.Errorf("Failed to shut down integration: %v", err)
	}
}

func TestStartSpan(t *testing.T) {
	// Test the StartSpan helper function
	ctx := context.Background()
	provider, err := tests.SetupTestTracer()
	if err != nil {
		t.Fatalf("Failed to setup tracer: %v", err)
	}
	defer provider.Shutdown(context.Background())

	tracer := provider.Tracer("test-tracer")
	
	// Use the StartSpan helper function
	newCtx, span := vayuOtel.StartSpan(ctx, tracer, "test-span")
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

	tracer := provider.Tracer("attributes-tracer")
	_, span := tracer.Start(context.Background(), "attributes-span")
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

	tracer := provider.Tracer("events-tracer")
	_, span := tracer.Start(context.Background(), "events-span")
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

	tracer := provider.Tracer("end-span-tracer")
	_, span := tracer.Start(context.Background(), "end-span")
	
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

	tracer := provider.Tracer("error-tracer")
	_, span := tracer.Start(context.Background(), "error-span")
	defer span.End()

	// Create a test error
	testErr := errors.New("test error")
	
	// Use the RecordSpanError helper function
	vayuOtel.RecordSpanError(span, testErr)

	// No direct way to verify error was recorded in the testing API
	// This test just ensures the function doesn't panic
}

func TestMultipleTracerNames(t *testing.T) {
	// Create a provider for testing
	provider, err := tests.SetupTestTracer()
	if err != nil {
		t.Fatalf("Failed to setup tracer: %v", err)
	}
	defer provider.Shutdown(context.Background())

	// Test creating spans with different tracer names
	ctx := context.Background()

	// Create a root span with a specific tracer name
	rootTracer := provider.Tracer("root-tracer")
	ctx, rootSpan := rootTracer.Start(ctx, "root-span")
	defer rootSpan.End()

	// Create a child span with a different tracer name
	childTracer := provider.Tracer("child-tracer")
	ctx, childSpan := childTracer.Start(ctx, "child-span")
	defer childSpan.End()

	// Create a grandchild span with yet another tracer name
	grandchildTracer := provider.Tracer("grandchild-tracer")
	_, grandchildSpan := grandchildTracer.Start(ctx, "grandchild-span")
	defer grandchildSpan.End()

	// Add some attributes and events to verify functionality
	rootSpan.SetAttributes(attribute.String("level", "root"))
	childSpan.SetAttributes(attribute.String("level", "child"))
	grandchildSpan.SetAttributes(attribute.String("level", "grandchild"))

	rootSpan.AddEvent("root-event")
	childSpan.AddEvent("child-event")
	grandchildSpan.AddEvent("grandchild-event")
}
