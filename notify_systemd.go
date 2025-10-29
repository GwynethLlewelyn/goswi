//go:build systemd

package main

import (
	"github.com/coreos/go-systemd/daemon"
)

// talk to systemd, inform that we're reloading
func notify(notification notificationType) {
	switch notification {
	case appReloading:
		b, err := daemon.SdNotify(false, daemon.SdNotifyReloading)
		// (false, nil) - notification not supported (i.e. NOTIFY_SOCKET is unset)
		// (false, err) - notification supported, but failure happened (e.g. error connecting to NOTIFY_SOCKET or while sending data)
		// (true, nil) - notification supported, data has been sentif
		switch {
		case !b && err == nil:
			// the logging system is not available, either, so we just print out
			config.LogWarn("❌ systemd not available")
			activeSystemd = false
		case !b && err != nil:
			config.LogWarn("💣 systemd answered with error:", err)
		case b && err == nil:
			config.LogInfo("✅ systemd was succesfully notified that we're starting")
		default:
			config.LogWarn("🤷‍♀️ unknown/confused systemd status, ignoring")
		}
	}
}
