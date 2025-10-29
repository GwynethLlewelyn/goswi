//go:build !systemd

package main

import (
	"github.com/coreos/go-systemd/daemon"
)

// talk to systemd, inform that we're reloading
func notify(notification notificationType) {
	switch notification {
	case appReloading:
		config.LogTrace("loading or reloading app...")
	}
}
