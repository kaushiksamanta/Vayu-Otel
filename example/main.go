package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/kaushiksamanta/vayu"
	vayuOtel "github.com/kaushiksamanta/vayu-otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	// Create a new Vayu app
	app := vayu.New()

	// Configure OpenTelemetry integration
	otelConfig := vayuOtel.DefaultConfig()
	otelConfig.ServiceName = "vayu-demo"
	otelConfig.ServiceVersion = "1.0.0"

	// Configure to send traces to Jaeger via OTLP exporter
	// Don't use stdout for production, we'll send to Jaeger instead
	otelConfig.UseStdout = false
	otelConfig.OTLPEndpoint = "localhost:4317" // Jaeger OTLP gRPC endpoint
	otelConfig.Insecure = true                 // Don't use TLS for local development

	// Set up integration options
	options := vayuOtel.DefaultSetupOptions()
	options.App = app
	options.Config = otelConfig

	// Initialize OpenTelemetry
	otelIntegration, err := vayuOtel.Setup(options)
	if err != nil {
		log.Fatalf("Failed to initialize OpenTelemetry: %v", err)
	}

	// Ensure graceful shutdown
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := otelIntegration.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down OpenTelemetry: %v", err)
		}
	}()

	// Add some basic middleware
	app.Use(vayu.Logger())
	app.Use(vayu.Recovery())

	// Basic route
	app.GET("/", func(c *vayu.Context, next vayu.NextFunc) {
		// Create a span directly using a specific tracer name
		tracer := otelIntegration.GetTracer("homepage-tracer")
		ctx, span := tracer.Start(c.Request.Context(), "homepage", trace.WithAttributes(
			attribute.String("custom.attribute", "homepage"),
			attribute.Bool("demo", true),
		))
		defer span.End()

		// Update request context with the span context
		c.Request = c.Request.WithContext(ctx)

		c.JSON(http.StatusOK, map[string]string{
			"message": "Welcome to Vayu with OpenTelemetry!",
		})
	})

	// Route with custom child span
	app.GET("/users/:id", func(c *vayu.Context, next vayu.NextFunc) {
		// In a real implementation we would use route params
		// For this example, use query parameter instead
		userID := c.Query("id")
		if userID == "" {
			userID = "default"
		}

		// Create the main span for the user operation with a specific tracer name
		userTracer := otelIntegration.GetTracer("user-service-tracer")
		ctx, userSpan := userTracer.Start(c.Request.Context(), "get_user", trace.WithAttributes(
			attribute.String("user.id", userID),
			attribute.String("user.name", "user"+userID),
			attribute.String("user.email", "user"+userID+"@example.com"),
		))
		defer userSpan.End()

		// Create a child span for database operation with a different tracer name
		dbTracer := otelIntegration.GetTracer("database-service-tracer")
		_, dbSpan := dbTracer.Start(ctx, "database.query", trace.WithAttributes(
			attribute.String("db.operation", "get_user"),
			attribute.String("db.user_id", userID),
		))

		// Simulate database query
		time.Sleep(50 * time.Millisecond)
		dbSpan.End()

		// Send a JSON response using type-safe method
		c.JSON(http.StatusOK, map[string]string{
			"id":    userID,
			"name":  "User " + userID,
			"email": "user" + userID + "@example.com",
		})
	})

	// Route with error handling
	app.GET("/error", func(c *vayu.Context, next vayu.NextFunc) {
		// Create a span and record an error with a specific tracer name
		errorTracer := otelIntegration.GetTracer("error-handling-tracer")
		_, span := errorTracer.Start(c.Request.Context(), "error_handler")
		defer span.End()

		// Create an error
		err := fmt.Errorf("something went wrong")

		// Record the error in the span
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	})

	// Route for product retrieval
	app.GET("/products/:id", func(c *vayu.Context, next vayu.NextFunc) {
		// In a real implementation we would use route params
		// For this example, use query parameter instead
		productID := c.Params["id"]
		if productID == "" {
			productID = "default"
		}

		// Create the main product span with a specific tracer name
		productTracer := otelIntegration.GetTracer("product-service-tracer")
		ctx, productSpan := productTracer.Start(c.Request.Context(), "get_product", trace.WithAttributes(
			attribute.String("product.id", productID),
			attribute.String("user.ip", c.Request.RemoteAddr),
		))
		defer productSpan.End()

		// Create a validation span with a specific tracer name
		validateTracer := otelIntegration.GetTracer("validation-service-tracer")
		_, validateSpan := validateTracer.Start(ctx, "product.validate")

		// Simulate validation
		var err error
		time.Sleep(10 * time.Millisecond)
		if productID == "0" {
			err = fmt.Errorf("invalid product ID")
			// Record error in the validation span
			validateSpan.RecordError(err)
			validateSpan.SetStatus(codes.Error, err.Error())
		}
		validateSpan.End()

		if err != nil {
			c.JSON(http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
			return
		}

		// Add an event to the product span
		productSpan.AddEvent("product.accessed", trace.WithAttributes(
			attribute.String("product.id", productID),
		))

		c.JSON(http.StatusOK, map[string]string{
			"id":   productID,
			"name": "Product " + productID,
		})
	})

	// Example of complex span hierarchy
	app.GET("/span-hierarchy", func(c *vayu.Context, next vayu.NextFunc) {
		// Create our own self-contained span hierarchy with a custom tracer name
		tracer := otelIntegration.GetTracer("nested-span-demo")

		// Start with the root span (span1)
		ctx := c.Request.Context()
		ctx, span1 := tracer.Start(ctx, "/span-hierarchy", trace.WithAttributes(
			attribute.String("span.type", "root"),
		))
		defer span1.End() // Will be closed when the HTTP handler completes

		// Create span2 as a child of span1
		ctx2, span2 := tracer.Start(ctx, "/span-hierarchy/child-span2", trace.WithAttributes(
			attribute.String("span.type", "parent"),
		))
		defer span2.End()

		// Create span3 as a child of span2
		_, span3 := tracer.Start(ctx2, "/span-hierarchy/child-span3", trace.WithAttributes(
			attribute.String("span.type", "child1"),
		))
		span3.AddEvent("Processing item 1")
		time.Sleep(10 * time.Millisecond) // Simulate work
		span3.End()

		// Create span4 as another child of span2
		_, span4 := tracer.Start(ctx2, "/span-hierarchy/child-span4", trace.WithAttributes(
			attribute.String("span.type", "child2"),
		))
		span4.AddEvent("Processing item 2")
		time.Sleep(15 * time.Millisecond) // Simulate work
		span4.End()

		// Create span5 as a sibling of span2 (child of span1)
		// Use the span1 context to make it a direct child of span1
		_, span5 := tracer.Start(ctx, "/span-hierarchy/child-span5", trace.WithAttributes(
			attribute.String("span.type", "sibling"),
		))
		span5.AddEvent("Finalizing process")
		time.Sleep(5 * time.Millisecond) // Simulate work
		span5.End()

		c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Created complex span hierarchy",
			"structure": []string{
				"span1 (root)",
				"  └── span2 (parent)",
				"       ├── span3 (child of span2)",
				"       └── span4 (child of span2)",
				"  └── span5 (sibling of span2, child of span1)",
			},
		})
	})

	// Example of using custom tracer names for each span
	app.GET("/custom-tracer-names", func(c *vayu.Context, next vayu.NextFunc) {
		// Get the base context
		ctx := c.Request.Context()

		// Get a tracer with a custom name for the main operation
		mainTracer := otelIntegration.GetTracer("main-operation-tracer")
		ctx, mainSpan := mainTracer.Start(ctx, "main-operation", trace.WithAttributes(
			attribute.String("operation.type", "main"),
		))
		defer mainSpan.End()

		// Simulate some work
		time.Sleep(10 * time.Millisecond)

		// Get a different tracer for the database operation
		dbTracer := otelIntegration.GetTracer("database-tracer")
		ctx, dbSpan := dbTracer.Start(ctx, "database-query", trace.WithAttributes(
			attribute.String("db.operation", "query"),
			attribute.String("db.system", "postgresql"),
		))
		dbSpan.AddEvent("Executing query")
		time.Sleep(20 * time.Millisecond)
		dbSpan.End()

		// Get another tracer for the cache operation
		cacheTracer := otelIntegration.GetTracer("cache-tracer")
		_, cacheSpan := cacheTracer.Start(ctx, "cache-lookup", trace.WithAttributes(
			attribute.String("cache.operation", "get"),
			attribute.String("cache.system", "redis"),
		))
		cacheSpan.AddEvent("Cache hit")
		time.Sleep(5 * time.Millisecond)
		cacheSpan.End()

		// Return a response with information about the tracers used
		c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Used custom tracer names for different operations",
			"tracers": []string{
				"main-operation-tracer",
				"database-tracer",
				"cache-tracer",
			},
		})
	})

	// Start the server
	fmt.Println("Server started at http://localhost:8080")
	app.Listen(":8080")
}
