package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/kaushiksamanta/vayu"
	vayuOtel "github.com/kaushiksamanta/vayu-otel"
)

func main() {
	// Create a new Vayu app
	app := vayu.New()

	// Set up OpenTelemetry with automatic tracing
	// This automatically adds the tracing middleware to the app
	config := vayuOtel.DefaultConfig()

	config.ServiceName = "auto-trace-demo"
	integration, err := vayuOtel.TraceAllRequests(app, config)
	if err != nil {
		log.Fatalf("Failed to initialize OpenTelemetry: %v", err)
	}

	// Ensure graceful shutdown
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := integration.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down OpenTelemetry: %v", err)
		}
	}()

	app.Use(vayu.Logger())
	app.Use(vayu.Recovery())

	// Simple home route - automatically traced by the middleware
	app.GET("/", func(c *vayu.Context, next vayu.NextFunc) {
		// The request is already being traced by the middleware
		// No need to create a span manually
		c.JSON(http.StatusOK, map[string]string{
			"message": "Welcome to auto-traced Vayu!",
		})
	})

	// Route with child span for more detailed tracing
	app.GET("/users/:id", func(c *vayu.Context, next vayu.NextFunc) {
		// Get user ID from route params
		userID := c.Params["id"]
		if userID == "" {
			userID = "default"
		}

		// The parent span is already created by the middleware
		// We can create a child span for the database operation
		// Get the current context which contains the parent span
		ctx := c.Request.Context()

		// Create a child span for database operation directly from the context
		// This is simpler in auto-tracing scenarios - no need to get a tracer explicitly
		dbSpan := vayuOtel.Start(ctx, "database.get_user")
		// Add attributes using the map style API
		dbSpan.AddAttributes(map[string]interface{}{
			"db.operation": "get_user",
			"db.user_id":   userID,
		})
		defer dbSpan.End()

		// Simulate database query
		time.Sleep(50 * time.Millisecond)

		// Return the response
		c.JSON(http.StatusOK, map[string]string{
			"id":    userID,
			"name":  "User " + userID,
			"email": "user" + userID + "@example.com",
		})
	})

	// Route that demonstrates span hierarchy
	app.GET("/span-hierarchy", func(c *vayu.Context, next vayu.NextFunc) {
		ctx := c.Request.Context()

		// Create span2 as a child of span1
		span2 := vayuOtel.Start(ctx, "/span-hierarchy/child-span2")
		span2.AddAttributes(map[string]interface{}{
			"span.type": "parent",
		})
		defer span2.End()

		// Get the context from span2
		ctx2 := span2.Context()

		// Create span3 as a child of span2
		span3 := vayuOtel.Start(ctx2, "/span-hierarchy/child-span3")
		span3.AddAttributes(map[string]interface{}{
			"span.type": "child1",
		})
		span3.AddEvent("Processing item 1")
		time.Sleep(10 * time.Millisecond) // Simulate work
		span3.End()

		// Create span4 as another child of span2
		span4 := vayuOtel.Start(ctx2, "/span-hierarchy/child-span4")
		span4.AddAttributes(map[string]interface{}{
			"span.type": "child2",
		})
		span4.AddEvent("Processing item 2")
		time.Sleep(15 * time.Millisecond) // Simulate work
		span4.End()

		// Create span5 as a sibling of span2 (child of span1)
		// Use the span1 context to make it a direct child of span1
		span5 := vayuOtel.Start(ctx, "/span-hierarchy/child-span5")
		span5.AddAttributes(map[string]interface{}{
			"span.type": "sibling",
		})
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

	// Route that demonstrates error handling with auto-tracing
	app.GET("/error", func(c *vayu.Context, next vayu.NextFunc) {
		ctx := c.Request.Context()

		// Create a span for the operation that will have an error
		errorSpan := vayuOtel.Start(ctx, "/error-example/operation")
		// Add attributes using the map style API
		errorSpan.AddAttributes(map[string]interface{}{
			"operation.type": "error-demo",
		})
		defer errorSpan.End()

		// Simulate an error
		err := errors.New("something went wrong")

		// Record the error on the span using the fluent API
		errorSpan.RecordError(err)

		// The middleware will also automatically mark the parent span as error
		// based on the HTTP status code
		c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	})

	// Start the server
	log.Println("Auto-tracing example server started at http://localhost:8080")
	log.Println("Available routes:")
	log.Println("  - GET / : Home page")
	log.Println("  - GET /span-hierarchy : Span hierarchy example")
	log.Println("  - GET /users/:id : User details with child span")
	log.Println("  - GET /error : Error example")
	app.Listen(":8080")
}
