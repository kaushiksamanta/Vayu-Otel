package vayuotel

import (
	"context"

	"github.com/kaushiksamanta/vayu"
)

// Integration provides an easy-to-use integration with the Vayu framework
type Integration struct {
	provider *Provider
	app      *vayu.App
}

// SetupOptions contains the options for setting up the integration
type SetupOptions struct {
	// App is the Vayu application instance
	App *vayu.App

	// Config is the OpenTelemetry configuration
	Config Config
}

// DefaultSetupOptions returns default setup options
func DefaultSetupOptions() SetupOptions {
	return SetupOptions{
		Config: DefaultConfig(),
	}
}

// Setup sets up OpenTelemetry integration with Vayu
func Setup(options SetupOptions) (*Integration, error) {
	if options.App == nil {
		return nil, ErrNoApp
	}

	integration := &Integration{
		app: options.App,
	}

	// Initialize provider
	provider, err := NewProvider(options.Config)
	if err != nil {
		return nil, err
	}
	integration.provider = provider

	return integration, nil
}

// Shutdown gracefully shuts down the OpenTelemetry integration
func (i *Integration) Shutdown(ctx context.Context) error {
	if i.provider != nil {
		return i.provider.Shutdown(ctx)
	}
	return nil
}
