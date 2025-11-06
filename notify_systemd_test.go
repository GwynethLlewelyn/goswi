//go:build systemd

// testing messages to systemd

package main

import (
	"os"
	"testing"

	"github.com/coreos/go-systemd/daemon"
)

func TestSystemdNotify(t *testing.T) {
	t.Log("Testing appReloading...")
	notify(appReloading)
	t.Log("Testing appReady...")
	notify(appReady)
	t.Log("Testing appStopping...")
	notify(appStopping)
	// probably not a good idea (gwyneth 20251106)
	// t.Log("Testing appStopping & exiting with error code 126")
	// notify(appStoppingError)
}

// Result from
// (false, nil) - notification not supported (i.e. NOTIFY_SOCKET is unset)
// (false, err) - notification supported, but failure happened (e.g. error connecting to NOTIFY_SOCKET or while sending data)
// (true, nil) - notification supported, data has been sentif
func TestSystemdNotify_SocketUnset(t *testing.T) {
	t.Logf("NOTIFY_SOCKET before test battery: %q", os.Getenv("NOTIFY_SOCKET"))

	t.Log("Testing Reloading...")
	if err := systemdNotifyAux(t, false, daemon.SdNotifyReloading); err != nil {
		t.Error(err)
	}
	t.Log("Testing Ready...")
	if err := systemdNotifyAux(t, false, daemon.SdNotifyReady); err != nil {
		t.Error(err)
	}
	t.Log("Testing Stopping...")
	if err := systemdNotifyAux(t, false, daemon.SdNotifyStopping); err != nil {
		t.Error(err)
	}
	t.Logf("NOTIFY_SOCKET after test battery: %q", os.Getenv("NOTIFY_SOCKET"))
}

// Same as bfore, but with the NOTIFY_SOCKET forced to be unset.
func TestSystemdNotify_SocketSet(t *testing.T) {
	t.Logf("NOTIFY_SOCKET before test battery: %q", os.Getenv("NOTIFY_SOCKET"))

	t.Log("Testing Reloading...")
	if err := systemdNotifyAux(t, true, daemon.SdNotifyReloading); err != nil {
		t.Error(err)
	}
	t.Log("Testing Ready...")
	if err := systemdNotifyAux(t, true, daemon.SdNotifyReady); err != nil {
		t.Error(err)
	}
	t.Log("Testing Stopping...")
	if err := systemdNotifyAux(t, true, daemon.SdNotifyStopping); err != nil {
		t.Error(err)
	}
	t.Logf("NOTIFY_SOCKET after test battery: %q", os.Getenv("NOTIFY_SOCKET"))
}

// Auxiliary function to communicate with the systemd daemon and spew out the appropriate messages.
func systemdNotifyAux(t *testing.T, socket bool, state string) error {
	b, err := daemon.SdNotify(socket, state)
	// (false, nil) - notification not supported (i.e. NOTIFY_SOCKET is unset)
	// (false, err) - notification supported, but failure happened (e.g. error connecting to NOTIFY_SOCKET or while sending data)
	// (true, nil) - notification supported, data has been sentif

	t.Logf("ℹ️ systemdNotifyAux() is sending message %q to systemd", state)

	switch {
	case !b && err == nil:
		t.Log("❌ systemd not available")
	case !b && err != nil:
		t.Log("💣 systemd answered with error:", err)
	case b && err == nil:
		t.Log("✅ systemd was succesfully notified that we're starting")
	default:
		// very likely this message is invalid, doesn't exist, or is not appropriate here...
		t.Log("🤷‍♀️ unknown/confused systemd status, ignoring")
	}
	return err
}
