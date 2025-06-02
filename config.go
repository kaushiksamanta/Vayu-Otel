package vayuotel

import (
	"context"
	"time"

	"maps"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Config holds configuration for OpenTelemetry integration
type Config struct {
	// ServiceName is the name of the service (required)
	ServiceName string

	// ServiceVersion is the version of the service
	ServiceVersion string

	// Environment is the deployment environment (e.g., "production", "staging")
	Environment string

	// OTLPEndpoint is the endpoint for the OpenTelemetry collector (e.g., "localhost:4317")
	OTLPEndpoint string

	// UseStdout enables printing traces to stdout (useful for development)
	UseStdout bool

	// Insecure disables transport security for gRPC connections to the collector
	Insecure bool

	// Headers to add to the gRPC connection
	Headers map[string]string

	// BatchTimeout is the maximum time to wait for a batch to be exported
	BatchTimeout time.Duration

	// BatchSize is the maximum number of spans to batch before exporting
	BatchSize int

	// AdditionalAttributes are custom attributes to add to every span
	AdditionalAttributes []ResourceAttribute
}

// ResourceAttribute is a key-value pair to add to resource attributes
type ResourceAttribute struct {
	Key   string
	Value string
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
	return Config{
		ServiceName:    "vayu-service",
		ServiceVersion: "0.1.0",
		Environment:    "development",
		OTLPEndpoint:   "localhost:4317",
		UseStdout:      false,
		Insecure:       true,
		BatchTimeout:   5 * time.Second,
		BatchSize:      512,
	}
}

// Provider is the OpenTelemetry provider that holds resources needed for telemetry
type Provider struct {
	TracerProvider *sdktrace.TracerProvider
	Config         Config
}

// NewProvider creates and initializes a new OpenTelemetry provider
func NewProvider(cfg Config) (*Provider, error) {
	ctx := context.Background()

	// Create resource attributes
	resourceAttrs := []ResourceAttribute{
		{Key: string(semconv.ServiceNameKey), Value: cfg.ServiceName},
	}

	if cfg.ServiceVersion != "" {
		resourceAttrs = append(resourceAttrs, ResourceAttribute{
			Key:   string(semconv.ServiceVersionKey),
			Value: cfg.ServiceVersion,
		})
	}

	if cfg.Environment != "" {
		resourceAttrs = append(resourceAttrs, ResourceAttribute{
			Key:   string(semconv.DeploymentEnvironmentKey),
			Value: cfg.Environment,
		})
	}

	// Add user-provided attributes
	resourceAttrs = append(resourceAttrs, cfg.AdditionalAttributes...)

	// Convert to OTel attribute format
	attrs := make([]attribute.KeyValue, 0, len(resourceAttrs))
	for _, attr := range resourceAttrs {
		attrs = append(attrs, attribute.String(attr.Key, attr.Value))
	}

	// Create resource
	res, err := resource.New(ctx, resource.WithAttributes(attrs...))
	if err != nil {
		return nil, err
	}

	// Create appropriate exporter based on configuration
	var exporter sdktrace.SpanExporter
	if cfg.UseStdout {
		exporter, err = stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
		)
	} else {
		// Set up OTLP exporter
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
		}

		// Configure security options
		if cfg.Insecure {
			opts = append(opts, otlptracegrpc.WithInsecure())
			opts = append(opts, otlptracegrpc.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())))
		}

		// Add headers if provided
		if len(cfg.Headers) > 0 {
			headers := make(map[string]string)
			maps.Copy(headers, cfg.Headers)
			opts = append(opts, otlptracegrpc.WithHeaders(headers))
		}

		// Create OTLP client
		client := otlptracegrpc.NewClient(opts...)
		exporter, err = otlptrace.New(ctx, client)
	}
	if err != nil {
		return nil, err
	}

	// Create batch span processor
	bsp := sdktrace.NewBatchSpanProcessor(
		exporter,
		sdktrace.WithBatchTimeout(cfg.BatchTimeout),
		sdktrace.WithMaxExportBatchSize(cfg.BatchSize),
	)

	// Create trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	// Set global provider and propagator
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return &Provider{
		TracerProvider: tp,
		Config:         cfg,
	}, nil
}

// Shutdown gracefully shuts down the provider
func (p *Provider) Shutdown(ctx context.Context) error {
	if p.TracerProvider != nil {
		return p.TracerProvider.Shutdown(ctx)
	}
	return nil
}
