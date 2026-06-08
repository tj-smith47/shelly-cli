package cmdutil

import "net"

// DeviceDisplayName resolves the stored device name to apply during a restore or
// migrate. An explicit name always wins. Otherwise the identifier is used, but
// only when it is a friendly alias — never a raw address — so a device is never
// named after its IP, MAC, or host:port. Returns "" to leave the existing name
// unchanged.
func DeviceDisplayName(explicit, identifier string) string {
	if explicit != "" {
		return explicit
	}
	if isAddressLike(identifier) {
		return ""
	}
	return identifier
}

// isAddressLike reports whether an identifier is a network address (IP, MAC, or
// host:port) rather than a friendly alias.
func isAddressLike(id string) bool {
	if id == "" {
		return true
	}
	if net.ParseIP(id) != nil {
		return true
	}
	if _, err := net.ParseMAC(id); err == nil {
		return true
	}
	if host, _, err := net.SplitHostPort(id); err == nil && net.ParseIP(host) != nil {
		return true
	}
	return false
}
