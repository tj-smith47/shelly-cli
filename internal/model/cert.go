// Package model defines core domain types for the Shelly CLI.
package model

// CertInstallData holds certificate data for installation on a device.
type CertInstallData struct {
	CAData   []byte // CA certificate data (PEM format)
	CertData []byte // Client certificate data (PEM format)
	KeyData  []byte // Client private key data (PEM format)
}
