// Package install provides the cert install subcommand.
package install

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// Options holds the command options.
type Options struct {
	CAFile     string
	ClientCert string
	ClientKey  string
}

// NewCommand creates the cert install command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{}

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
			return run(cmd.Context(), f, args[0], opts)
		},
	}

	cmd.Flags().StringVar(&opts.CAFile, "ca", "", "CA certificate file (PEM format)")
	cmd.Flags().StringVar(&opts.ClientCert, "client-cert", "", "Client certificate file (PEM format)")
	cmd.Flags().StringVar(&opts.ClientKey, "client-key", "", "Client private key file (PEM format)")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, opts *Options) error {
	ios := f.IOStreams()
	svc := f.ShellyService()

	if opts.CAFile == "" && opts.ClientCert == "" {
		return fmt.Errorf("specify --ca or --client-cert")
	}

	if opts.ClientCert != "" && opts.ClientKey == "" {
		return fmt.Errorf("--client-key required with --client-cert")
	}

	ios.StartProgress("Installing certificate...")

	conn, err := svc.Connect(ctx, device)
	if err != nil {
		ios.StopProgress()
		return fmt.Errorf("connect: %w", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			ios.DebugErr("close connection", err)
		}
	}()

	// Install CA certificate if provided
	if opts.CAFile != "" {
		caData, err := os.ReadFile(opts.CAFile)
		if err != nil {
			ios.StopProgress()
			return fmt.Errorf("read CA file: %w", err)
		}

		params := map[string]any{
			"data": string(caData),
		}

		_, err = conn.Call(ctx, "Shelly.PutUserCA", params)
		if err != nil {
			ios.StopProgress()
			return fmt.Errorf("install CA: %w", err)
		}

		ios.StopProgress()
		ios.Success("Installed CA certificate on %s", device)
	}

	// Install client certificate if provided
	if opts.ClientCert != "" {
		certData, err := os.ReadFile(opts.ClientCert)
		if err != nil {
			return fmt.Errorf("read client cert: %w", err)
		}

		keyData, err := os.ReadFile(opts.ClientKey)
		if err != nil {
			return fmt.Errorf("read client key: %w", err)
		}

		params := map[string]any{
			"data": string(certData),
			"key":  string(keyData),
		}

		_, err = conn.Call(ctx, "Shelly.PutTLSClientCert", params)
		if err != nil {
			return fmt.Errorf("install client cert: %w", err)
		}

		ios.Success("Installed client certificate on %s", device)
	}

	return nil
}
