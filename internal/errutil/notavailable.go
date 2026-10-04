package errutil

import (
	"errors"
	"strings"

	"github.com/tj-smith47/shelly-go/rpc"
)

// RPC error codes with which a Gen2+ device says it has no such method:
// 404 "No handler for <Method>" from the firmware's HTTP and WebSocket RPC,
// -106 from its RPC layer, and -32601 from JSON-RPC.
const (
	rpcCodeNoHandler      = 404
	rpcCodeMethodNotFound = -106
	rpcCodeJSONRPCMissing = -32601
)

const noHandlerPrefix = "No handler for "

// NotAvailableError reports that a device has no component or RPC method for
// what a command asked, for example a thermostat command sent to a relay.
type NotAvailableError struct {
	// Feature is the component the device lacks, as the device names it
	// ("Thermostat", "LoRa").
	Feature string
	Err     error
}

func (e *NotAvailableError) Error() string {
	return e.Feature + " not available on this device: " + e.Err.Error()
}

func (e *NotAvailableError) Unwrap() error { return e.Err }

// IsNotAvailable reports whether err is a device's answer that it has no
// handler for the RPC method called.
func IsNotAvailable(err error) bool {
	var rpcErr *rpc.ErrorObject
	if !errors.As(err, &rpcErr) {
		return false
	}
	switch rpcErr.Code {
	case rpcCodeNoHandler, rpcCodeMethodNotFound, rpcCodeJSONRPCMissing:
		return true
	}
	return false
}

// NotAvailable returns err as a *NotAvailableError naming feature when the
// device answered that it has no handler for the method called; any other
// error, and an error that already is a *NotAvailableError, is returned
// unchanged. With an empty feature the name is read from the device's answer
// ("No handler for Thermostat.GetStatus" names "Thermostat"), falling back to
// "This feature".
func NotAvailable(feature string, err error) error {
	if !IsNotAvailable(err) {
		return err
	}
	var already *NotAvailableError
	if errors.As(err, &already) {
		return err
	}
	if feature == "" {
		feature = featureFromAnswer(err)
	}
	return &NotAvailableError{Feature: feature, Err: err}
}

// featureFromAnswer reads the component name from a "No handler for X.Y"
// answer, or returns "This feature".
func featureFromAnswer(err error) string {
	var rpcErr *rpc.ErrorObject
	if !errors.As(err, &rpcErr) {
		return "This feature"
	}
	method, ok := strings.CutPrefix(rpcErr.Message, noHandlerPrefix)
	if component, _, _ := strings.Cut(method, "."); ok && component != "" {
		return component
	}
	return "This feature"
}
