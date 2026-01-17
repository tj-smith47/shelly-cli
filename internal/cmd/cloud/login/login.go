// Package login provides the cloud login subcommand.
package login

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/cloud"

	"github.com/tj-smith47/shelly-cli/internal/browser"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly/network"
)

// Options holds command options.
type Options struct {
	Factory *cmdutil.Factory

	// Auth key method
	Key    string
	Server string

	// Email/password method
	Email    string
	Password string

	// Browser OAuth method (default)
	Port      int
	NoBrowser bool
	Timeout   time.Duration
}

// NewCommand creates the cloud login command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "login",
		Aliases: []string{"auth", "signin"},
		Short:   "Authenticate with Shelly Cloud",
		Long: `Authenticate with the Shelly Cloud API.

The access token is stored locally for future use.

Three authentication methods are available:

1. Browser OAuth (default):
   Opens your web browser to the Shelly Cloud login page. After you log in,
   the authorization code is automatically captured. This is the most secure
   method as your password is never stored locally.

2. Auth Key (--key, --server):
   Use the authorization key from the Shelly mobile app. Find it in:
   User Settings → Authorization cloud key. You must also provide the
   server URL shown with the key.

3. Email/Password (--email, --password):
   Provide your Shelly Cloud email and password via flags or environment
   variables (SHELLY_CLOUD_EMAIL, SHELLY_CLOUD_PASSWORD).`,
		Example: `  # OAuth browser flow (default, most secure)
  shelly cloud login

  # Browser flow without auto-opening browser
  shelly cloud login --no-browser

  # Auth key from Shelly App
  shelly cloud login --key MTZkZGM3dWlk... --server shelly-59-eu.shelly.cloud

  # Email/password login
  shelly cloud login --email user@example.com --password mypassword`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), opts)
		},
	}

	// Auth key flags
	cmd.Flags().StringVar(&opts.Key, "key", "", "Authorization key from Shelly App")
	cmd.Flags().StringVar(&opts.Server, "server", "", "Server URL for auth key (e.g., shelly-59-eu.shelly.cloud)")

	// Email/password flags
	cmd.Flags().StringVar(&opts.Email, "email", "", "Shelly Cloud email")
	cmd.Flags().StringVar(&opts.Password, "password", "", "Shelly Cloud password")

	// Browser OAuth flags (default method, these modify behavior)
	cmd.Flags().IntVar(&opts.Port, "port", 0, "Port for OAuth callback server (default: auto-select)")
	cmd.Flags().BoolVar(&opts.NoBrowser, "no-browser", false, "Don't auto-open browser, just print the URL")
	cmd.Flags().DurationVar(&opts.Timeout, "timeout", 5*time.Minute, "Timeout waiting for OAuth callback")

	return cmd
}

//nolint:gocyclo // Login command handles three distinct auth methods in one function
func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()

	// Determine which method to use based on flags
	hasKeyFlags := opts.Key != "" || opts.Server != "" ||
		os.Getenv("SHELLY_CLOUD_AUTH_KEY") != "" || os.Getenv("SHELLY_CLOUD_SERVER") != ""
	hasPasswordFlags := opts.Email != "" || opts.Password != "" ||
		os.Getenv("SHELLY_CLOUD_EMAIL") != "" || os.Getenv("SHELLY_CLOUD_PASSWORD") != ""

	// Validate mutually exclusive options
	if hasKeyFlags && hasPasswordFlags {
		return errors.New("cannot combine --key/--server and --email/--password flags")
	}

	// Route to appropriate method based on flags provided
	switch {
	case hasKeyFlags:
		// === AUTH KEY LOGIN ===
		ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
		defer cancel()

		authKey := opts.Key
		if authKey == "" {
			authKey = os.Getenv("SHELLY_CLOUD_AUTH_KEY")
		}
		if authKey == "" {
			if !ios.CanPrompt() {
				return errors.New("auth key required (use --key flag or SHELLY_CLOUD_AUTH_KEY env var)")
			}
			var err error
			authKey, err = iostreams.Password("Authorization Key")
			if err != nil {
				return fmt.Errorf("failed to read auth key: %w", err)
			}
		}

		serverURL := opts.Server
		if serverURL == "" {
			serverURL = os.Getenv("SHELLY_CLOUD_SERVER")
		}
		if serverURL == "" {
			if !ios.CanPrompt() {
				return errors.New("server URL required (use --server flag or SHELLY_CLOUD_SERVER env var)")
			}
			var err error
			serverURL, err = iostreams.InputRequired("Server URL (e.g., shelly-59-eu.shelly.cloud)")
			if err != nil {
				return fmt.Errorf("failed to read server URL: %w", err)
			}
		}

		if authKey == "" || serverURL == "" {
			return errors.New("auth key and server URL are required")
		}

		return cmdutil.RunWithSpinner(ctx, ios, "Validating credentials...", func(ctx context.Context) error {
			client := network.NewCloudClientWithAuthKey(authKey, serverURL)
			if _, err := client.GetAllDevices(ctx); err != nil {
				return fmt.Errorf("authentication failed: %w", err)
			}

			cfg := config.Get()
			cfg.Cloud.Enabled = true
			cfg.Cloud.AuthKey = authKey
			cfg.Cloud.ServerURL = serverURL
			cfg.Cloud.AccessToken = ""
			cfg.Cloud.Email = ""

			if err := config.Save(); err != nil {
				return fmt.Errorf("failed to save credentials: %w", err)
			}

			ios.Success("Authenticated with auth key")
			ios.Info("Server: %s", serverURL)
			return nil
		})

	case hasPasswordFlags:
		// === EMAIL/PASSWORD LOGIN ===
		ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
		defer cancel()

		email := opts.Email
		if email == "" {
			email = os.Getenv("SHELLY_CLOUD_EMAIL")
		}
		if email == "" {
			if !ios.CanPrompt() {
				return errors.New("email required (use --email flag or SHELLY_CLOUD_EMAIL env var)")
			}
			var err error
			email, err = iostreams.InputRequired("Email")
			if err != nil {
				return fmt.Errorf("failed to read email: %w", err)
			}
		}

		password := opts.Password
		if password == "" {
			password = os.Getenv("SHELLY_CLOUD_PASSWORD")
		}
		if password == "" {
			if !ios.CanPrompt() {
				return errors.New("password required (use --password flag or SHELLY_CLOUD_PASSWORD env var)")
			}
			var err error
			password, err = iostreams.Password("Password")
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}
		}

		if email == "" || password == "" {
			return errors.New("email and password are required")
		}

		return cmdutil.RunWithSpinner(ctx, ios, "Authenticating with Shelly Cloud...", func(ctx context.Context) error {
			_, result, err := network.NewCloudClientWithCredentials(ctx, email, password)
			if err != nil {
				return fmt.Errorf("authentication failed: %w", err)
			}

			cfg := config.Get()
			cfg.Cloud.Enabled = true
			cfg.Cloud.Email = email
			cfg.Cloud.AccessToken = result.Token
			cfg.Cloud.ServerURL = result.UserAPIURL
			cfg.Cloud.AuthKey = ""

			if err := config.Save(); err != nil {
				return fmt.Errorf("failed to save credentials: %w", err)
			}

			ios.Success("Logged in as %s", email)
			if !result.Expiry.IsZero() {
				ios.Info("Token expires: %s", result.Expiry.Format("2006-01-02 15:04:05"))
			}
			return nil
		})

	default:
		// === BROWSER OAUTH LOGIN (default) ===
		lc := net.ListenConfig{}
		listener, err := lc.Listen(ctx, "tcp", fmt.Sprintf("127.0.0.1:%d", opts.Port))
		if err != nil {
			return fmt.Errorf("failed to start callback server: %w", err)
		}
		defer func() {
			ios.DebugErr("close listener", listener.Close())
		}()

		tcpAddr, ok := listener.Addr().(*net.TCPAddr)
		if !ok {
			return fmt.Errorf("unexpected listener address type: %T", listener.Addr())
		}
		port := tcpAddr.Port
		redirectURI := fmt.Sprintf("http://127.0.0.1:%d/callback", port)
		authorizeURL := cloud.AuthorizeURL(cloud.ClientIDDIY, redirectURI)

		codeCh := make(chan string, 1)
		errCh := make(chan error, 1)

		mux := http.NewServeMux()
		mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
			code := r.URL.Query().Get("code")
			if code == "" {
				errMsg := r.URL.Query().Get("error")
				if errMsg == "" {
					errMsg = "no authorization code received"
				}
				http.Error(w, errMsg, http.StatusBadRequest)
				errCh <- fmt.Errorf("oauth callback error: %s", errMsg)
				return
			}

			w.Header().Set("Content-Type", "text/html")
			_, writeErr := fmt.Fprint(w, `<!DOCTYPE html><html><body>
				<h1>Login Successful!</h1>
				<p>You can close this window and return to the CLI.</p>
				<script>window.close();</script>
			</body></html>`)
			ios.DebugErr("write callback response", writeErr)

			codeCh <- code
		})

		server := &http.Server{
			Handler:           mux,
			ReadHeaderTimeout: 10 * time.Second,
		}

		go func() {
			if serveErr := server.Serve(listener); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
				errCh <- serveErr
			}
		}()

		ios.Info("Opening browser for Shelly Cloud login...")
		ios.Println("")
		ios.Plain("If browser doesn't open, visit this URL:")
		ios.Plain("  %s", authorizeURL)
		ios.Println("")

		if !opts.NoBrowser {
			b := browser.New()
			if browseErr := b.Browse(ctx, authorizeURL); browseErr != nil {
				ios.Warning("Could not open browser automatically")
			}
		}

		ios.StartProgress("Waiting for authorization...")

		ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
		defer cancel()

		var code string
		select {
		case code = <-codeCh:
			ios.StopProgress()
			ios.DebugErr("shutdown server", server.Shutdown(context.Background()))
		case err := <-errCh:
			ios.StopProgress()
			ios.DebugErr("shutdown server", server.Shutdown(context.Background()))
			return err
		case <-ctx.Done():
			ios.StopProgress()
			ios.DebugErr("shutdown server", server.Shutdown(context.Background()))
			return fmt.Errorf("login timed out after %s", opts.Timeout)
		}

		ios.StartProgress("Exchanging authorization code...")
		token, err := cloud.ExchangeCode(ctx, "api.shelly.cloud", code, cloud.ClientIDDIY)
		ios.StopProgress()
		if err != nil {
			return fmt.Errorf("failed to exchange code: %w", err)
		}

		cfg := config.Get()
		cfg.Cloud.Enabled = true
		cfg.Cloud.AccessToken = token.AccessToken
		cfg.Cloud.ServerURL = token.UserAPIURL
		cfg.Cloud.AuthKey = ""
		cfg.Cloud.Email = ""

		if err := config.Save(); err != nil {
			return fmt.Errorf("failed to save credentials: %w", err)
		}

		ios.Success("Logged in via browser")
		if token.UserAPIURL != "" {
			ios.Info("Server: %s", token.UserAPIURL)
		}
		if !token.Expiry.IsZero() {
			ios.Info("Token expires: %s", token.Expiry.Format("2006-01-02 15:04:05"))
		}
		return nil
	}
}
