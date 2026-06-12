//go:build !linux

package client

import "syscall"

// bindControl is a no-op on platforms without SO_BINDTODEVICE: there is no
// portable way to force a socket's egress interface, so connections fall back to
// the default route. The --to-ap AP-isolation case this guards against is
// Linux-only in practice (the host doing the AP hop is Linux).
func bindControl(string) func(network, address string, c syscall.RawConn) error {
	return nil
}
