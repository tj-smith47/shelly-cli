package testutil

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/tj-smith47/shelly-go/discovery"
)

// ErrScannerCalled is returned by every CountingWiFiScanner call that has no
// answer set.
var ErrScannerCalled = errors.New("WiFi scanner called")

// CountingWiFiScanner is a WiFi backend that counts every call, keeping apart the
// calls that only read the host's WiFi state (CurrentNetwork and
// HostNetworkPassword) and the ones that would move it (Scan, Connect,
// Disconnect). Moves are always refused with ErrScannerCalled, so a test can
// prove a path never scans, joins or leaves a network. Reads answer from
// CurrentSSID and Passwords, and are refused with ErrScannerCalled when unset.
// Set the fields before the scanner is used.
type CountingWiFiScanner struct {
	// CurrentSSID is the network the host is on; empty means none.
	CurrentSSID string
	// Passwords holds the passphrases the host has stored, by SSID.
	Passwords map[string]string

	reads atomic.Int32
	moves atomic.Int32
}

// Calls returns how many scanner methods were called, reads and moves alike.
func (s *CountingWiFiScanner) Calls() int { return s.Reads() + s.Moves() }

// Reads returns how many CurrentNetwork and HostNetworkPassword calls were made.
func (s *CountingWiFiScanner) Reads() int { return int(s.reads.Load()) }

// Moves returns how many Scan, Connect and Disconnect calls were made.
func (s *CountingWiFiScanner) Moves() int { return int(s.moves.Load()) }

// Scan counts the call and refuses it.
func (s *CountingWiFiScanner) Scan(context.Context) ([]discovery.WiFiNetwork, error) {
	s.moves.Add(1)
	return nil, ErrScannerCalled
}

// Connect counts the call and refuses it.
func (s *CountingWiFiScanner) Connect(context.Context, string, string) error {
	s.moves.Add(1)
	return ErrScannerCalled
}

// Disconnect counts the call and refuses it.
func (s *CountingWiFiScanner) Disconnect(context.Context) error {
	s.moves.Add(1)
	return ErrScannerCalled
}

// CurrentNetwork counts the call and reports CurrentSSID, or refuses when it
// is empty.
func (s *CountingWiFiScanner) CurrentNetwork(context.Context) (*discovery.WiFiNetwork, error) {
	s.reads.Add(1)
	if s.CurrentSSID == "" {
		return nil, ErrScannerCalled
	}
	return &discovery.WiFiNetwork{SSID: s.CurrentSSID}, nil
}

// HostNetworkPassword counts the call and returns the passphrase stored for
// ssid in Passwords, or refuses when there is none.
func (s *CountingWiFiScanner) HostNetworkPassword(_ context.Context, ssid string) (string, error) {
	s.reads.Add(1)
	if pass, ok := s.Passwords[ssid]; ok {
		return pass, nil
	}
	return "", ErrScannerCalled
}
