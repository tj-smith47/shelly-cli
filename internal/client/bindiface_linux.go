//go:build linux

package client

import (
	"syscall"

	"golang.org/x/sys/unix"
)

// bindControl returns a net.Dialer Control hook that forces the socket to egress
// the named interface via SO_BINDTODEVICE. This overrides the kernel's route-by-
// metric choice, which on a host with two interfaces on the same subnet would
// otherwise send to the wireless egress regardless of source address.
//
// Binding requires CAP_NET_RAW (or root); without it SetsockoptString returns
// EPERM, the dial fails, and the confirm loop falls through to its unbound
// default-route candidate — never worse than not binding at all.
func bindControl(iface string) func(network, address string, c syscall.RawConn) error {
	return func(_, _ string, c syscall.RawConn) error {
		var sockErr error
		if err := c.Control(func(fd uintptr) {
			sockErr = unix.SetsockoptString(int(fd), unix.SOL_SOCKET, unix.SO_BINDTODEVICE, iface)
		}); err != nil {
			return err
		}
		return sockErr
	}
}
