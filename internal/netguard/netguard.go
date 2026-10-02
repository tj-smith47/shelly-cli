// Package netguard keeps test binaries off the network. It does nothing
// outside `go test`. Under `go test` it guards:
//
//   - every connection made through http.DefaultTransport, which covers each
//     http.Client with no Transport of its own (the cli's GitHub, telemetry,
//     alert, download and doctor clients, and the SDK's cloud, integrator,
//     firmware and WiFi-discovery clients) and each transport cloned from it
//     (generation detection, the SDK's device identification);
//   - the device HTTP clients internal/client builds, whose dial is
//     NetDialContext;
//   - the cli's own websocket dial (cloud events), whose dial is
//     NetDialContext;
//   - the SDK device websockets (debug websocket, automation events,
//     monitoring snapshots), all built by client.NewDeviceWebSocket, which
//     calls RefuseAddr with the device address first;
//   - every hostname lookup through net.DefaultResolver;
//   - the cli's mDNS and CoIoT constructors (shelly.NewMDNSDiscoverer,
//     shelly.NewCoIoTDiscoverer and each gen1.NewCoIoTListener call) and the
//     service's default factory access point flows (reprovision.Restore,
//     Onboard and Inspect), which call Refuse first.
//
// A guarded connection may reach only a loopback address or "localhost".
//
// Not guarded, because the SDK builds the socket and offers the cli no client
// or dialer option: the SDK's own mDNS/CoIoT/BLE sweepers when constructed
// outside the cli constructors above, and any reprovision function called
// other than through the service's flows. scripts/audit-conventions.sh refuses
// a transport.NewWebSocket call outside client.NewDeviceWebSocket, a raw
// mDNS/CoIoT constructor without Refuse, and a cli test that calls
// reprovision.Restore, Onboard, Inspect or ScanAPs directly.
package netguard

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"testing"
)

// ErrBlocked is returned for a connection a test tried to open to a
// non-loopback address or an unresolved hostname.
var ErrBlocked = errors.New("netguard: test connection to a non-loopback address refused")

func init() {
	if !testing.Testing() {
		return
	}
	if t, ok := http.DefaultTransport.(*http.Transport); ok {
		t.DialContext = DialContext
	}
	// A refusing DNS dial leaves only the hosts file, so "localhost" still
	// resolves and every other name fails before any query leaves the host.
	net.DefaultResolver = &net.Resolver{
		PreferGo: true,
		Dial: func(_ context.Context, _, address string) (net.Conn, error) {
			return nil, fmt.Errorf("%w: DNS query to %s", ErrBlocked, address)
		},
	}
}

// DialContext dials addr only when it is a loopback address or "localhost".
func DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	if !Allowed(addr) {
		return nil, fmt.Errorf("%w: %s", ErrBlocked, addr)
	}
	var d net.Dialer
	return d.DialContext(ctx, network, addr)
}

// Allowed reports whether a test may connect to addr (host or host:port).
func Allowed(addr string) bool {
	host := addr
	if h, _, err := net.SplitHostPort(addr); err == nil {
		host = h
	}
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// Refuse returns ErrBlocked under `go test` and nil otherwise, for code about
// to reach the network where no dialer guard sees it: a LAN multicast query,
// or an SDK operation that builds its own sockets.
func Refuse(what string) error {
	if testing.Testing() {
		return fmt.Errorf("%w: %s", ErrBlocked, what)
	}
	return nil
}

// RefuseAddr returns ErrBlocked under `go test` when addr (host or host:port)
// is not one Allowed accepts, and nil otherwise, for a connection the cli
// cannot give a dialer to.
func RefuseAddr(addr string) error {
	if testing.Testing() && !Allowed(addr) {
		return fmt.Errorf("%w: %s", ErrBlocked, addr)
	}
	return nil
}

// NetDialContext returns DialContext under `go test` and nil otherwise, for a
// dialer field (http.Transport.DialContext, websocket.Dialer.NetDialContext)
// where nil selects the library's default dial.
func NetDialContext() func(ctx context.Context, network, addr string) (net.Conn, error) {
	if !testing.Testing() {
		return nil
	}
	return DialContext
}
