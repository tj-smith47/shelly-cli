// Package install provides the cert install subcommand.
package install

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
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

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	spec := shelly.CertInstallSpec{
		CAFile:     opts.CAFile,
		ClientCert: opts.ClientCert,
		ClientKey:  opts.ClientKey,
	}
	// Validate eagerly so a bad flag combination errors without starting a spinner.
	if err := spec.Validate(); err != nil {
		return err
	}

	var result shelly.CertInstallResult
	err := cmdutil.RunWithSpinner(ctx, ios, "Installing certificate...", func(ctx context.Context) error {
		var installErr error
		result, installErr = svc.InstallCert(ctx, opts.Device, spec)
		return installErr
	})
	if err != nil {
		return err
	}

	if result.InstalledCA {
		ios.Success("Installed CA certificate on %s", opts.Device)
	}
	if result.InstalledClient {
		ios.Success("Installed client certificate on %s", opts.Device)
	}

	return nil
}
