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

	// Read files outside the spinner (I/O can be done before spinner)
	var caData, certData, keyData []byte
	var err error

	if opts.CAFile != "" {
		caData, err = os.ReadFile(opts.CAFile)
		if err != nil {
			return fmt.Errorf("read CA file: %w", err)
		}
	}

	if opts.ClientCert != "" {
		certData, err = os.ReadFile(opts.ClientCert)
		if err != nil {
			return fmt.Errorf("read client cert: %w", err)
		}
		keyData, err = os.ReadFile(opts.ClientKey)
		if err != nil {
			return fmt.Errorf("read client key: %w", err)
		}
	}

	installedCA := false
	installedClient := false

	err = cmdutil.RunWithSpinner(ctx, ios, "Installing certificate...", func(ctx context.Context) error {
		conn, connErr := svc.Connect(ctx, device)
		if connErr != nil {
			return fmt.Errorf("connect: %w", connErr)
		}
		defer func() {
			if closeErr := conn.Close(); closeErr != nil {
				ios.DebugErr("close connection", closeErr)
			}
		}()

		// Install CA certificate if provided
		if opts.CAFile != "" {
			params := map[string]any{
				"data": string(caData),
			}
			_, callErr := conn.Call(ctx, "Shelly.PutUserCA", params)
			if callErr != nil {
				return fmt.Errorf("install CA: %w", callErr)
			}
			installedCA = true
		}

		// Install client certificate if provided
		if opts.ClientCert != "" {
			params := map[string]any{
				"data": string(certData),
				"key":  string(keyData),
			}
			_, callErr := conn.Call(ctx, "Shelly.PutTLSClientCert", params)
			if callErr != nil {
				return fmt.Errorf("install client cert: %w", callErr)
			}
			installedClient = true
		}

		return nil
	})
	if err != nil {
		return err
	}

	if installedCA {
		ios.Success("Installed CA certificate on %s", device)
	}
	if installedClient {
		ios.Success("Installed client certificate on %s", device)
	}

	return nil
}
