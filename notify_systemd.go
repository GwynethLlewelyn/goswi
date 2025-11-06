//go:build systemd

package main

import (
	"os"

	"github.com/coreos/go-systemd/daemon"
)

// talk to systemd, inform that we're reloading
func notify(notification notificationType) {
	// TODO(gwyneth): check if activeSystemd == true

	switch notification {
	case appReloading:
		// Silently ignoring errors. We already write to the log inside systemdNotify().
		systemdNotify(daemon.SdNotifyReloading)
	case appReady:
		systemdNotify(daemon.SdNotifyReady)
	case appStopping:
		// Normal exit, no need to panic :)
		systemdNotify(daemon.SdNotifyStopping)
	case appStoppingError:
		// Notify systemd that we're stopping on fatal error.
		systemdNotify(daemon.SdNotifyStopping + "\nEXIT_STATUS=126")
		os.Exit(126) // error code 126 usually means "could not execute"
	}
}

// Auxiliary function to communicate with the systemd daemon and spew out the appropriate messages.
func systemdNotify(state string) error {
	b, err := daemon.SdNotify(false, state)
	// (false, nil) - notification not supported (i.e. NOTIFY_SOCKET is unset)
	// (false, err) - notification supported, but failure happened (e.g. error connecting to NOTIFY_SOCKET or while sending data)
	// (true, nil) - notification supported, data has been sentif

	config.LogTracef("systemdNotify() is sending message %q to systemd", state)

	switch {
	case !b && err == nil:
		config.LogWarn("❌ systemd not available")
		activeSystemd = false
	case !b && err != nil:
		config.LogWarn("💣 systemd answered with error:", err)
	case b && err == nil:
		config.LogInfo("✅ systemd was succesfully notified that we're starting")
	default:
		// very likely this message is invalid, doesn't exist, or is not appropriate here...
		config.LogWarn("🤷‍♀️ unknown/confused systemd status, ignoring")
	}
	return err
}
