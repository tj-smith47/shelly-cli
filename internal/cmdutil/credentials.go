package cmdutil

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// DefaultDeviceUser is the user name Shelly devices use for authentication.
const DefaultDeviceUser = "admin"

// ErrNoCredentials is returned by DeviceCredentials.Auth when no credential
// flag was given.
var ErrNoCredentials = errors.New("no credentials given")

// AddDeviceCredentialFlags registers --user, --password and --password-stdin
// for c, with passwordUsage and stdinUsage as the help of the last two.
func AddDeviceCredentialFlags(cmd *cobra.Command, c *DeviceCredentials, passwordUsage, stdinUsage string) {
	cmd.Flags().StringVar(&c.User, "user", "", "Username for --password or --password-stdin (default admin)")
	AddSecretFlags(cmd, &c.Password, &c.PasswordStdin, "password", passwordUsage, stdinUsage)
}

// DeviceCredentials holds the credential flags of a command that stores
// device credentials: either Pair ("user:pass", from --auth) or User with
// Password or PasswordStdin.
type DeviceCredentials struct {
	Pair          string
	User          string
	Password      string
	PasswordStdin bool
}

// Auth returns the credentials the flags describe, or ErrNoCredentials when
// none were given. The user defaults to DefaultDeviceUser when only a password is
// given. With PasswordStdin the password is read from ios.In, or prompted for
// without echo when stdin is a terminal.
func (c DeviceCredentials) Auth(ios *iostreams.IOStreams) (*model.Auth, error) {
	if c.Pair != "" {
		user, pass, ok := strings.Cut(c.Pair, ":")
		if !ok || user == "" {
			return nil, errors.New("invalid --auth format, expected user:pass")
		}
		return &model.Auth{Username: user, Password: pass}, nil
	}
	if c.Password == "" && !c.PasswordStdin {
		if c.User != "" {
			return nil, errors.New("--user needs --password or --password-stdin")
		}
		return nil, ErrNoCredentials
	}

	password := c.Password
	if err := ResolveSecret(ios, &password, c.PasswordStdin, "password", "Password"); err != nil {
		return nil, err
	}

	user := c.User
	if user == "" {
		user = DefaultDeviceUser
	}
	return &model.Auth{Username: user, Password: password}, nil
}

// StdinFlagSuffix ends the name of the flag that reads a secret from stdin
// in place of the flag that takes it as a value: --password and
// --password-stdin.
const StdinFlagSuffix = "-stdin"

// AddSecretFlags registers --<name>, which takes a secret as its value, and
// --<name>-stdin, which reads it from stdin so it stays out of shell history
// and the process list. The two are mutually exclusive, and so is the new
// -stdin flag with every other -stdin flag on cmd, since stdin holds only one
// secret.
func AddSecretFlags(cmd *cobra.Command, value *string, fromStdin *bool, name, usage, stdinUsage string) {
	AddSecretFlagsP(cmd, value, fromStdin, name, "", usage, stdinUsage)
}

// AddSecretFlagsP is AddSecretFlags with a one-letter shorthand for --<name>.
func AddSecretFlagsP(cmd *cobra.Command, value *string, fromStdin *bool, name, shorthand, usage, stdinUsage string) {
	stdinName := name + StdinFlagSuffix
	cmd.Flags().StringVarP(value, name, shorthand, "", usage)
	cmd.Flags().BoolVar(fromStdin, stdinName, false, stdinUsage)
	cmd.MarkFlagsMutuallyExclusive(name, stdinName)
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if f.Name != stdinName && strings.HasSuffix(f.Name, StdinFlagSuffix) && f.Value.Type() == "bool" {
			cmd.MarkFlagsMutuallyExclusive(f.Name, stdinName)
		}
	})
}

// ResolveSecret sets *value to the secret read from stdin when fromStdin is
// set (the --<name>-stdin flag AddSecretFlags registers) and leaves it alone
// otherwise. One trailing line ending is dropped. When stdin is a terminal it
// prompts for the secret without echo, showing prompt.
func ResolveSecret(ios *iostreams.IOStreams, value *string, fromStdin bool, name, prompt string) error {
	if !fromStdin {
		return nil
	}
	if ios.IsStdinTTY() {
		secret, err := iostreams.Password(prompt)
		if err != nil {
			return fmt.Errorf("read --%s: %w", name, err)
		}
		*value = secret
		return nil
	}
	data, err := io.ReadAll(ios.In)
	if err != nil {
		return fmt.Errorf("read --%s from stdin: %w", name, err)
	}
	secret := strings.TrimSuffix(strings.TrimSuffix(string(data), "\n"), "\r")
	if secret == "" {
		return fmt.Errorf("--%s%s: stdin is empty", name, StdinFlagSuffix)
	}
	*value = secret
	return nil
}
