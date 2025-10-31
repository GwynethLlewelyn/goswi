//go:build !newrelic && !opentelemetry

package main

// gOSWI was compiled without any telemetry middleware, thus this is a no-op.
func initTelemetry() {
	// ... nothing to see here ...
}
