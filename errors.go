package vayuotel

import "errors"

// Common errors returned by the vayuotel package.
var (
	// ErrNoApp is returned when trying to set up the integration without providing a Vayu app
	ErrNoApp = errors.New("no Vayu app instance provided")

	// ErrInvalidConfig is returned when the provided configuration is invalid
	ErrInvalidConfig = errors.New("invalid OpenTelemetry configuration")

	// ErrProviderNotInitialized is returned when trying to use the provider before initialization
	ErrProviderNotInitialized = errors.New("OpenTelemetry provider not initialized")
)
