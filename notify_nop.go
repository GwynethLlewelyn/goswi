//go:build !systemd

package main

import (
	"os"
)

// talk to systemd, inform that we're reloading
func notify(notification notificationType) {
	switch notification {
	case appReloading:
		config.LogTrace("(re)loading app configuration...")
	case appReady:
		config.LogTrace("app configuration loaded, system is now ready!")
	case appStopping:
		config.LogTrace("stopping app...")
	case appStoppingError:
		config.LogFatal("fatal system error, code 126")
		os.Exit(126)
	}
}
