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
)

// Use the common SetupTestTracer from the tests package

func TestDefaultSetupOptions(t *testing.T) {
	options := vayuOtel.DefaultSetupOptions()

	// Verify default values
	if options.Config.ServiceName != "vayu-service" {
		t.Errorf("Expected ServiceName to be 'vayu-service', got '%s'", options.Config.ServiceName)
	}

	if !options.EnableTracing {
		t.Error("Expected EnableTracing to be true")
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

func TestDirectSpanAttributes(t *testing.T) {
	// Create a span and add it to context
	provider, err := tests.SetupTestTracer()
	if err != nil {
		t.Fatalf("Failed to setup tracer: %v", err)
	}
	defer provider.Shutdown(context.Background())

	// Get a tracer with a specific name
	tracer := provider.Tracer("custom-test-tracer")
	_, span := tracer.Start(context.Background(), "test-span")

	// Add attributes directly to span
	span.SetAttributes(
		attribute.String("test.key", "value"),
		attribute.Int("test.count", 42),
	)

	// No direct way to verify the attributes were added
	// Just make sure it doesn't panic
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

func TestDirectSpanOperations(t *testing.T) {
	// Create a span directly
	provider, err := tests.SetupTestTracer()
	if err != nil {
		t.Fatalf("Failed to setup tracer: %v", err)
	}
	defer provider.Shutdown(context.Background())

	tracer := provider.Tracer("span-operations-tracer")

	// Create test function with error capture
	testFunction := func(ctx context.Context, name string, expectError bool) error {
		// Create child span for the function
		_, span := tracer.Start(ctx, name)
		defer span.End()

		// Simulate work
		time.Sleep(5 * time.Millisecond)

		// Handle error case
		if expectError {
			err := errors.New("test error")
			// Record error in span
			span.RecordError(err)
			span.SetStatus(1, err.Error())
			return err
		}

		return nil
	}

	// Test successful function execution
	ctx := context.Background()
	err = testFunction(ctx, "successful-operation", false)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	// Test with error
	errResult := testFunction(ctx, "error-operation", true)
	if errResult == nil {
		t.Error("Expected error, got nil")
	}
}

func TestDirectSpanInHandler(t *testing.T) {
	// Create tracer
	provider, err := tests.SetupTestTracer()
	if err != nil {
		t.Fatalf("Failed to setup tracer: %v", err)
	}
	defer provider.Shutdown(context.Background())

	// Get a tracer with a specific name for the handler
	handlerTracer := provider.Tracer("handler-tracer")

	// Create a request context
	reqCtx := context.Background()

	// Handler function that creates a span
	handlerCalled := false
	handleRequest := func() {
		// Create a span directly in the handler
		ctx, span := handlerTracer.Start(reqCtx, "handler-span")
		defer span.End()

		// Add attributes
		span.SetAttributes(
			attribute.String("handler.name", "test-handler"),
			attribute.String("http.method", "GET"),
			attribute.String("http.path", "/test"),
		)

		// Record event
		span.AddEvent("handler.executed")

		// Create a child span for database operation with a different tracer name
		dbTracer := provider.Tracer("db-tracer")
		_, dbSpan := dbTracer.Start(ctx, "database-operation")
		dbSpan.SetAttributes(attribute.String("db.operation", "query"))
		time.Sleep(1 * time.Millisecond) // Simulate work
		dbSpan.End()

		handlerCalled = true
	}

	// Execute handler
	handleRequest()

	// Verify handler was called
	if !handlerCalled {
		t.Error("Expected handler to be called")
	}
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
