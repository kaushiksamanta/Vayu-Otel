package unit

import (
	"context"
	"testing"
	"time"

	vayuOtel "github.com/kaushiksamanta/vayu-otel"
	"go.opentelemetry.io/otel/sdk/trace"
)

func TestDefaultConfig(t *testing.T) {
	cfg := vayuOtel.DefaultConfig()

	// Verify default values
	if cfg.ServiceName != "vayu-service" {
		t.Errorf("Expected ServiceName to be 'vayu-service', got '%s'", cfg.ServiceName)
	}

	if cfg.ServiceVersion != "0.1.0" {
		t.Errorf("Expected ServiceVersion to be '0.1.0', got '%s'", cfg.ServiceVersion)
	}

	if cfg.Environment != "development" {
		t.Errorf("Expected Environment to be 'development', got '%s'", cfg.Environment)
	}

	if cfg.OTLPEndpoint != "localhost:4317" {
		t.Errorf("Expected OTLPEndpoint to be 'localhost:4317', got '%s'", cfg.OTLPEndpoint)
	}

	if cfg.UseStdout != false {
		t.Errorf("Expected UseStdout to be false, got %v", cfg.UseStdout)
	}

	if cfg.Insecure != true {
		t.Errorf("Expected Insecure to be true, got %v", cfg.Insecure)
	}

	if cfg.BatchTimeout != 5*time.Second {
		t.Errorf("Expected BatchTimeout to be 5s, got %v", cfg.BatchTimeout)
	}
}

func TestProviderShutdown(t *testing.T) {
	// Test with nil provider
	p := &vayuOtel.Provider{}
	err := p.Shutdown(context.Background())
	if err != nil {
		t.Errorf("Expected nil error when shutting down nil provider, got %v", err)
	}

	// Test with valid provider
	p = &vayuOtel.Provider{
		TracerProvider: trace.NewTracerProvider(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = p.Shutdown(ctx)
	if err != nil {
		t.Errorf("Failed to shut down valid provider: %v", err)
	}
}
