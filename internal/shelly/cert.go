package shelly

import (
	"context"
	"fmt"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// tlsKeyParam is the RPC parameter name carrying the client private key in a
// Shelly.PutTLSClientCert call.
const tlsKeyParam = "key"

// CertInstallSpec describes a certificate installation request by file path.
type CertInstallSpec struct {
	CAFile     string
	ClientCert string
	ClientKey  string
}

// CertInstallResult reports which certificates were installed.
type CertInstallResult struct {
	InstalledCA     bool
	InstalledClient bool
}

// Validate checks that the spec names a coherent set of certificate files.
func (s CertInstallSpec) Validate() error {
	if s.CAFile == "" && s.ClientCert == "" {
		return fmt.Errorf("specify --ca or --client-cert")
	}
	if s.ClientCert != "" && s.ClientKey == "" {
		return fmt.Errorf("--client-key required with --client-cert")
	}
	return nil
}

// loadCertData reads the certificate files named by the spec.
func loadCertData(fs afero.Fs, spec CertInstallSpec) (*model.CertInstallData, error) {
	data := &model.CertInstallData{}
	var err error

	if spec.CAFile != "" {
		data.CAData, err = afero.ReadFile(fs, spec.CAFile)
		if err != nil {
			return nil, fmt.Errorf("read CA file: %w", err)
		}
	}

	if spec.ClientCert != "" {
		data.CertData, err = afero.ReadFile(fs, spec.ClientCert)
		if err != nil {
			return nil, fmt.Errorf("read client cert: %w", err)
		}
		data.KeyData, err = afero.ReadFile(fs, spec.ClientKey)
		if err != nil {
			return nil, fmt.Errorf("read client key: %w", err)
		}
	}

	return data, nil
}

// InstallCert validates the spec, reads the named certificate files, and installs
// them on a Gen2+ device. Certificate installation is unsupported on Gen1.
func (s *Service) InstallCert(ctx context.Context, identifier string, spec CertInstallSpec) (CertInstallResult, error) {
	if err := spec.Validate(); err != nil {
		return CertInstallResult{}, err
	}

	data, err := loadCertData(config.Fs(), spec)
	if err != nil {
		return CertInstallResult{}, err
	}

	var result CertInstallResult
	err = s.WithDevice(ctx, identifier, func(dev *DeviceClient) error {
		if dev.IsGen1() {
			return fmt.Errorf("certificate installation is only supported on Gen2+ devices")
		}

		conn := dev.Gen2()

		if len(data.CAData) > 0 {
			if _, callErr := conn.Call(ctx, "Shelly.PutUserCA", map[string]any{"data": string(data.CAData)}); callErr != nil {
				return fmt.Errorf("install CA: %w", callErr)
			}
			result.InstalledCA = true
		}

		if len(data.CertData) > 0 {
			params := map[string]any{"data": string(data.CertData), tlsKeyParam: string(data.KeyData)}
			if _, callErr := conn.Call(ctx, "Shelly.PutTLSClientCert", params); callErr != nil {
				return fmt.Errorf("install client cert: %w", callErr)
			}
			result.InstalledClient = true
		}

		return nil
	})

	return result, err
}
