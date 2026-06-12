package client

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

// bindInterfaceContextKey is the context key carrying the network interface a
// connection must egress through.
type bindInterfaceContextKey struct{}

// WithBindInterface returns a context that pins every connection opened under it
// to egress the named host interface (e.g. "eth0", "vmbr0"). An empty name is a
// no-op, returning the context unchanged so callers can pass a computed name
// without branching.
//
// This exists for the --to-ap confirm path: after the host hops back from a
// device's factory AP, a freshly-rejoined device sitting on the same wireless AP
// as the host is reached station-to-station and may be dropped by AP client
// isolation. On a multi-homed host the wired interface reaches it, but the kernel
// route table prefers the wireless egress for a shared subnet — so the source
// interface must be forced, not merely preferred. The binding is honoured by
// Connect and ConnectGen1, threaded purely through the context so no connection
// signature changes (mirroring ratelimit.MarkAsPolling).
//
// Enforcement is via SO_BINDTODEVICE on Linux and is a no-op on other platforms
// (see bindControl); a host lacking the privilege to bind falls back to the
// default route, never worse than the unbound behaviour.
func WithBindInterface(ctx context.Context, iface string) context.Context {
	if iface == "" {
		return ctx
	}
	return context.WithValue(ctx, bindInterfaceContextKey{}, iface)
}

// bindInterfaceFromContext returns the interface a connection must egress, or ""
// when none was set.
func bindInterfaceFromContext(ctx context.Context) string {
	iface, ok := ctx.Value(bindInterfaceContextKey{}).(string)
	if !ok {
		return ""
	}
	return iface
}

// boundHTTPClient builds an *http.Client whose dialer egresses the named
// interface, mirroring the transport defaults shelly-go's transport.NewHTTP
// applies to its own client (30s timeout, bounded idle pool). insecureTLS skips
// verification for https Shelly endpoints, matching the WithInsecureSkipVerify
// path taken for the default client.
func boundHTTPClient(iface string, insecureTLS bool) *http.Client {
	tr := &http.Transport{
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			Control:   bindControl(iface),
		}).DialContext,
	}
	if insecureTLS {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // Shelly devices use self-signed TLS certs; skipping verification matches the default client path
	}
	return &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
}
