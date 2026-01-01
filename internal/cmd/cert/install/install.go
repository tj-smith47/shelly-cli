// Package install provides the cert install subcommand.
package install

import (
	"context"
	"fmt"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Options holds the command options.
type Options struct {
	Factory    *cmdutil.Factory
	CAFile     string
	ClientCert string
	ClientKey  string
	Device     string
}

// NewCommand creates the cert install command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "install <device>",
		Aliases: []string{"add", "set", "upload"},
		Short:   "Install a certificate on a device",
		Long: `Install a TLS certificate on a Gen2+ Shelly device.

Supports installing CA certificates for MQTT/cloud TLS verification,
as well as client certificates for mutual TLS authentication.`,
		Example: `  # Install CA certificate
  shelly cert install kitchen --ca /path/to/ca.pem

  # Install client certificate and key
  shelly cert install kitchen --client-cert cert.pem --client-key key.pem`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.CAFile, "ca", "", "CA certificate file (PEM format)")
	cmd.Flags().StringVar(&opts.ClientCert, "client-cert", "", "Client certificate file (PEM format)")
	cmd.Flags().StringVar(&opts.ClientKey, "client-key", "", "Client private key file (PEM format)")

	return cmd
}

func (opts *Options) validate() error {
	if opts.CAFile == "" && opts.ClientCert == "" {
		return fmt.Errorf("specify --ca or --client-cert")
	}
	if opts.ClientCert != "" && opts.ClientKey == "" {
		return fmt.Errorf("--client-key required with --client-cert")
	}
	return nil
}

func (opts *Options) loadCertData() (*model.CertInstallData, error) {
	data := &model.CertInstallData{}
	var err error
	fs := config.Fs()

	if opts.CAFile != "" {
		data.CAData, err = afero.ReadFile(fs, opts.CAFile)
		if err != nil {
			return nil, fmt.Errorf("read CA file: %w", err)
		}
	}

	if opts.ClientCert != "" {
		data.CertData, err = afero.ReadFile(fs, opts.ClientCert)
		if err != nil {
			return nil, fmt.Errorf("read client cert: %w", err)
		}
		data.KeyData, err = afero.ReadFile(fs, opts.ClientKey)
		if err != nil {
			return nil, fmt.Errorf("read client key: %w", err)
		}
	}

	return data, nil
}

func run(ctx context.Context, opts *Options) error {
	if err := opts.validate(); err != nil {
		return err
	}

	data, err := opts.loadCertData()
	if err != nil {
		return err
	}

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	installedCA := false
	installedClient := false

	err = cmdutil.RunWithSpinner(ctx, ios, "Installing certificate...", func(ctx context.Context) error {
		return svc.WithDevice(ctx, opts.Device, func(dev *shelly.DeviceClient) error {
			if dev.IsGen1() {
				return fmt.Errorf("certificate installation is only supported on Gen2+ devices")
			}

			conn := dev.Gen2()

			if len(data.CAData) > 0 {
				if _, callErr := conn.Call(ctx, "Shelly.PutUserCA", map[string]any{"data": string(data.CAData)}); callErr != nil {
					return fmt.Errorf("install CA: %w", callErr)
				}
				installedCA = true
			}

			if len(data.CertData) > 0 {
				params := map[string]any{"data": string(data.CertData), "key": string(data.KeyData)}
				if _, callErr := conn.Call(ctx, "Shelly.PutTLSClientCert", params); callErr != nil {
					return fmt.Errorf("install client cert: %w", callErr)
				}
				installedClient = true
			}

			return nil
		})
	})
	if err != nil {
		return err
	}

	if installedCA {
		ios.Success("Installed CA certificate on %s", opts.Device)
	}
	if installedClient {
		ios.Success("Installed client certificate on %s", opts.Device)
	}

	return nil
}
