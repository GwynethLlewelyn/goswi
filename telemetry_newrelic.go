//go:build opentelemetry

// Note: this requires some external tool to capture the telemetry data.

package main

import (
	"log"
	"os"

	nrgin "github.com/newrelic/go-agent/v3/integrations/nrgin"
	"github.com/newrelic/go-agent/v3/newrelic"
)

// If we have a valid New Relic configuration, add it to the middleware list first (gwyneth 20210422)
// @see https://github.com/newrelic/go-agent/blob/v3.11.0/_integrations/nrgin/v1/example/main.go
func initTelemetry() {
	// environment overrides configuration
	appName := os.Getenv("NEW_RELIC_APP_NAME")
	licenseKey := os.Getenv("NEW_RELIC_LICENSE_KEY")

	if appName == "" {
		appName = *config["NewRelicAppName"]
	}
	if licenseKey == "" {
		licenseKey = *config["NewRelicLicenseKey"]
	}

	// check again!
	if appName != "" && licenseKey != "" {
		app, err := newrelic.NewApplication(
			newrelic.ConfigAppName(appName),
			newrelic.ConfigLicense(licenseKey),
			// Experiment to send New Relic logging to our 'default' logger. (gwyneth 20251030)
			newrelic.ConfigDebugLogger(log.Writer()),
			newrelic.ConfigInfoLogger(log.Writer()),
		)

		if err != nil {
			config.LogError("Failed to init New Relic, did you provide your App Name and License Key?\nError was:", err)
			return
		}
		// router is a *gin.Engine defined as a global singleton for this application
		router.Use(nrgin.Middleware(app))
	}
}
