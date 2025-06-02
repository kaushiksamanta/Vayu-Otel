# Vayu OpenTelemetry Integration

This package provides a focused OpenTelemetry tracing integration for the [Vayu web framework](https://github.com/kaushiksamanta/vayu), enabling distributed tracing and enhanced observability for your applications.

## Features

- **Distributed Tracing**: Automatically trace HTTP requests and create child spans for operations
- **Automatic Middleware**: One-line setup for end-to-end request tracing
- **Dynamic Tracer Names**: Create spans with different tracer names for better organization
- **Type-Safe Integration**: Leverages Vayu's type-safe context methods for clean integration
- **Minimal Configuration**: Simple API with sensible defaults
- **Flexible Options**: Fine-grained control over what gets traced
- **Standalone Package**: Keeps the core Vayu framework lean while providing full observability
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

### Option 1: One-line Automatic Tracing (Recommended)

The simplest way to add tracing to your Vayu application is with the `TraceAllRequests` function:

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
    // Create Vayu app
    app := vayu.New()

    // One line to set up tracing with automatic middleware
    otel, err := vayuOtel.TraceAllRequests(app, "my-service")
    if err != nil {
        log.Fatalf("Failed to initialize OpenTelemetry: %v", err)
    }

    // Ensure graceful shutdown
    defer func() {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        otel.Shutdown(ctx)
    }()

    // Define your routes - they will be automatically traced!
    app.GET("/", func(c *vayu.Context, next vayu.NextFunc) {
        // The request is already being traced by the middleware
        // You can access the current context which contains the span
        ctx := c.Request.Context()
        
        // Your handler logic here
        c.JSON(http.StatusOK, map[string]string{
            "message": "Hello, traced world!"
        })
    })

    app.Listen(":8080")
}
```

### Option 2: Manual Setup with More Control

```go
package main

import (
    "context"
    "log"
    "net/http"
    "time"

    "github.com/kaushiksamanta/vayu"
    vayuOtel "github.com/kaushiksamanta/vayu-otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/trace"
)

func main() {
    // Create Vayu app
    app := vayu.New()

    // Set up OpenTelemetry integration with default options
    options := vayuOtel.DefaultSetupOptions()
    options.App = app
    options.Config.ServiceName = "my-service"

    // Initialize OpenTelemetry
    otel, err := vayuOtel.Setup(options)
    if err != nil {
        log.Fatalf("Failed to initialize OpenTelemetry: %v", err)
    }

    // Ensure graceful shutdown
    defer func() {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        otel.Shutdown(ctx)
    }()

    // Add the tracing middleware with custom options
    app.Use(otel.Middleware(vayuOtel.MiddlewareOptions{
        TracerName: "http-requests",
        SpanNameFormatter: func(c *vayu.Context) string {
            return "Custom: " + c.Request.Method + " " + c.Request.URL.Path
        },
        CustomAttributes: func(c *vayu.Context) []attribute.KeyValue {
            return []attribute.KeyValue{
                attribute.String("custom.attribute", "value"),
                attribute.String("request.id", "12345"),
            }
        },
    }))

    // Define your routes
    app.GET("/", func(c *vayu.Context, next vayu.NextFunc) {
        // Get a tracer with a specific name for this operation
        tracer := otel.GetTracer("home-service")
        
        // Create a child span using the tracer
        ctx, span := tracer.Start(c.Request.Context(), "/home", trace.WithAttributes(
            attribute.String("handler", "home"),
        ))
        defer span.End()
        
        c.JSON(http.StatusOK, map[string]string{
            "message": "Hello, traced world!"
        })
    })

    app.Listen(":8080")
}
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
  // Get a tracer with a specific name
  tracer := otel.GetTracer("hierarchy-demo")

  // Start with the root span
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
  span3.End()

  // Create span4 as another child of span2
  _, span4 := tracer.Start(ctx2, "/span-hierarchy/child-span4", trace.WithAttributes(
    attribute.String("span.type", "child2"),
  ))
  span4.AddEvent("Processing item 2")
  span4.End()

  // Create span5 as a sibling of span2 (child of span1)
  // Use the span1 context to make it a direct child of span1
  _, span5 := tracer.Start(ctx, "/span-hierarchy/child-span5", trace.WithAttributes(
    attribute.String("span.type", "sibling"),
  ))
  span5.AddEvent("Finalizing process")
  span5.End()

  c.JSONMap(vayu.StatusOK, map[string]string{"status": "completed"})
})
```

This creates a span hierarchy that can be visualized in Jaeger UI as follows:

- `/span-hierarchy` (root span)
  - `/span-hierarchy/child-span2`
    - `/span-hierarchy/child-span3`
    - `/span-hierarchy/child-span4`
  - `/span-hierarchy/child-span5`

### Using Multiple Tracer Names

You can organize your spans by using different tracer names for different components of your application:

```go
// Get tracers for different services
userTracer := otel.GetTracer("user-service")
databaseTracer := otel.GetTracer("database-service")
authTracer := otel.GetTracer("auth-service")

// Use them in your handlers
app.GET("/users/:id", func(c *vayu.Context, next vayu.NextFunc) {
  // Start a span with the user service tracer
  ctx, span := userTracer.Start(c.Request.Context(), "/users/:id")
  defer span.End()
    
  // Use the database tracer for database operations
  ctx, dbSpan := databaseTracer.Start(ctx, "query.user")
  // ... database operations
  dbSpan.End()
    
  // ... rest of handler
})
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License
