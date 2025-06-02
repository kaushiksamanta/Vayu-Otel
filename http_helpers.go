package vayuotel

import (
	"net/http"
)

// Helper function to get the scheme from the request
func getScheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}

	// Check for X-Forwarded-Proto header
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		return proto
	}

	// Default to http
	return "http"
}
