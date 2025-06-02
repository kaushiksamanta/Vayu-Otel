# Vayu OpenTelemetry Integration

This package provides a focused OpenTelemetry tracing integration for the [Vayu web framework](https://github.com/kaushiksamanta/vayu), enabling distributed tracing and enhanced observability for your applications.

## Features

- **Distributed Tracing**: Automatically trace HTTP requests and create child spans for operations
- **Automatic Middleware**: One-line setup for end-to-end request tracing
- **Zero OpenTelemetry Imports**: End users don't need to import any OpenTelemetry packages directly
- **Fluent API**: Chainable methods for span operations (attributes, events, errors)
- **Type-Safe Integration**: Leverages Vayu's type-safe context methods for clean integration
- **Minimal Configuration**: Simple API with sensible defaults
- **Flexible Options**: Fine-grained control over what gets traced
- **Modular Code Structure**: Well-organized code split into focused files
- **Compatible with all OpenTelemetry backends**: Works with Jaeger, Zipkin, and any OpenTelemetry collector

## Screenshots

View your distributed traces in Jaeger UI or any other OpenTelemetry-compatible visualization tool:

![Trace Overview](screenshots/1.png)

Detailed view of spans with timing and attributes:

![Detailed Span View](screenshots/2.png)

## Installation

To use the OpenTelemetry integration with Vayu, first install both packages:

```bash
go get github.com/kaushiksamanta/vayu
go get github.com/kaushiksamanta/vayu-otel
```

## Quick Start

### Example Usage

The following examples are taken from the included example application in `example/main.go`:

```go
package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/kaushiksamanta/vayu"
	vayuOtel "github.com/kaushiksamanta/vayu-otel"
)

func main() {
	// Create a new Vayu app
	app := vayu.New()

	// Set up OpenTelemetry with automatic tracing in one line
	// This automatically adds the tracing middleware to the app
	integration, err := vayuOtel.TraceAllRequests(app, "auto-trace-demo")
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

	// Add other middleware if needed
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
		dbSpan := vayuOtel.Start(ctx, "database.get_user",
			vayuOtel.WithStringAttribute("db.operation", "get_user"),
			vayuOtel.WithStringAttribute("db.user_id", userID),
		)
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

	app.Listen(":8080")
}```

### Error Handling Example

The middleware automatically marks spans as errors based on HTTP status codes. Here's an example from the example application:

```go
// Route that demonstrates error handling with auto-tracing
app.GET("/error", func(c *vayu.Context, next vayu.NextFunc) {
	// Simulate an error
	// The middleware will automatically mark the span as error
	// based on the HTTP status code
	c.JSON(http.StatusInternalServerError, map[string]string{
		"error": "Something went wrong",
	})
})
```

## Configuration Options

### Provider Configuration

```go
config := vayuOtel.DefaultConfig()
config.ServiceName = "my-service"           // Required: Name of your service
config.ServiceVersion = "1.2.3"             // Optional: Version of your service
config.Environment = "production"           // Optional: Deployment environment
config.OTLPEndpoint = "collector:4317"      // Optional: OTLP endpoint for your collector
config.UseStdout = true                     // Optional: Print traces to stdout for development
config.Insecure = true                      // Optional: Use insecure connection to collector
```

### Complete Setup Example

```go
options := vayuOtel.DefaultSetupOptions()
options.App = app
options.Config = config                  // From above

options.EnableTracing = true             // Enable distributed tracing

otel, err := vayuOtel.Setup(options)
if err != nil {
  log.Fatalf("Failed to initialize OpenTelemetry: %v", err)
}
```

## Type-Safe Features

Vayu-OTel leverages Vayu's type-safe context methods to provide enhanced integration:

### Creating Spans with Dynamic Tracer Names

```go
app.GET("/users/:id", func(c *vayu.Context, next vayu.NextFunc) {
  userID := c.Param("id")

  // Get a tracer with a specific name for user operations
  tracer := otel.GetTracer("user-service")
  
  // Create a span for this handler
  ctx, span := tracer.Start(c.Request.Context(), "/users/:id", trace.WithAttributes(
    attribute.String("user.id", userID),
  ))
  defer span.End()

  // Create a child span for database operation
  dbTracer := otel.GetTracer("database-service")
  ctx, dbSpan := dbTracer.Start(ctx, "database.query", trace.WithAttributes(
    attribute.String("db.operation", "get_user"),
    attribute.String("db.user_id", userID),
  ))
    
  // Simulate database query
  time.Sleep(50 * time.Millisecond)
  dbSpan.End()

  // Use type-safe JSON response
  c.JSONMap(vayu.StatusOK, map[string]string{"id": userID})
})
```

## Working with OpenTelemetry Exporters

### Jaeger

```go
config := vayuOtel.DefaultConfig()
config.ServiceName = "my-service"
config.OTLPEndpoint = "jaeger:4317" // Jaeger OTLP endpoint
config.Insecure = true

// Set up integration...
```

### Zipkin

```go
config := vayuOtel.DefaultConfig()
config.ServiceName = "my-service"
config.OTLPEndpoint = "zipkin:4317" // Zipkin OTLP endpoint
config.Insecure = true

// Set up integration...
```

### Custom OTLP Endpoint

```go
config := vayuOtel.DefaultConfig()
config.ServiceName = "my-service"
config.OTLPEndpoint = "collector.example.com:4317"
config.Insecure = false // Use TLS
config.Headers = map[string]string{
    "Authorization": "Bearer token",
}

// Set up integration...
```

## Development & Testing

### Local Development with Stdout

For quick local development, you can use the stdout exporter to see traces directly in your console:

```go
config := vayuOtel.DefaultConfig()
config.ServiceName = "my-service"
config.UseStdout = true
```

### Using Docker Compose with Jaeger

This repository includes a Docker Compose setup that provides a complete tracing environment with Jaeger UI for visualizing traces:

1. Start the Docker Compose environment:

```bash
docker-compose up -d
```

2. Configure your application to send traces to Jaeger:

```go
config := vayuOtel.DefaultConfig()
config.ServiceName = "my-service"
config.UseStdout = false // Don't use stdout for production
config.OTLPEndpoint = "localhost:4317" // Jaeger OTLP gRPC endpoint
config.Insecure = true // Don't use TLS for local development
```

3. Run your application and generate some traffic

4. Open the Jaeger UI in your browser at http://localhost:16686

5. In the Jaeger UI:
   - Select your service from the "Service" dropdown
   - Click "Find Traces" to see your traces
   - Click on any trace to see the detailed span hierarchy
   - Explore the timeline view, tags, and logs for each span

The Docker Compose setup includes:
- Jaeger All-in-One (collector, query service, and UI)
- OpenTelemetry Collector (for production-like setup)

## Advanced Usage

### Creating Complex Span Hierarchies

The integration supports creating complex span hierarchies with multiple levels and sibling relationships. This is useful for visualizing complex operations with multiple nested steps:

```go
app.GET("/span-hierarchy", func(c *vayu.Context, next vayu.NextFunc) {
  ctx := c.Request.Context()

  // Create span2 as a child of span1
  span2 := vayuOtel.Start(ctx, "/span-hierarchy/child-span2")
  span2.AddAttributes(map[string]interface{}{
    "span.type": "parent",
  })
  defer span2.End()

  // Create span3 as a child of span2
  span3 := vayuOtel.Start(span2.Context(), "/span-hierarchy/child-span3")
  span3.AddAttributes(map[string]interface{}{
    "span.type": "child1",
  })
  span3.AddEvent("Processing item 1")
  time.Sleep(10 * time.Millisecond) // Simulate work
  span3.End()

  // Create span4 as another child of span2
  span4 := vayuOtel.Start(span2.Context(), "/span-hierarchy/child-span4")
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
```

This creates a span hierarchy that can be visualized in Jaeger UI as follows:

- `/span-hierarchy` (root span)
  - `/span-hierarchy/child-span2`
    - `/span-hierarchy/child-span3`
    - `/span-hierarchy/child-span4`
  - `/span-hierarchy/child-span5`



## Key Benefits of the Fluent API

The vayu-otel library has been designed with simplicity and usability in mind. Here are the key benefits of the fluent API:

1. **Zero OpenTelemetry Imports**: End users don't need to import any OpenTelemetry packages directly in their application code. All OpenTelemetry complexity is abstracted away.

2. **Fluent, Chainable Methods**: The `Span` wrapper provides chainable methods for adding attributes, events, and recording errors, enabling a more idiomatic and expressive coding style.

3. **Context Management**: The `Span` wrapper maintains its own context, making it easier to create child spans without manually tracking context variables.

4. **Consistent Tracer Naming**: The middleware uses a consistent tracer name (`vayu-http`) for all spans, which simplifies span organization and visualization.

5. **Type-Safe Attribute Helpers**: Helper functions like `WithStringAttribute`, `WithIntAttribute`, etc. provide type-safe ways to create span attributes.

6. **Reduced Boilerplate**: Common operations like adding attributes, recording events, and handling errors are simplified with fluent methods on the `Span` wrapper.

## Code Structure

The codebase has been reorganized into focused files for better maintainability:

- **span.go**: Contains the `Span` wrapper struct and its methods for adding attributes, events, recording errors, and ending spans.
- **attributes.go**: Contains attribute helper functions and `SpanOption` implementations.
- **integration.go**: Contains the `Integration` struct and setup functions.
- **middleware.go**: Contains the HTTP middleware for automatic tracing.
- **middleware_options.go**: Contains middleware configuration options.
- **context.go**: Contains context key definitions and helper functions.
- **config.go**: Contains configuration types and provider implementation.
- **errors.go**: Contains error definitions.
- **http_helpers.go**: Contains HTTP-related helper functions.
- **otel.go**: Main entry point that documents the package structure.

This modular structure makes the codebase easier to navigate, understand, and maintain.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License
