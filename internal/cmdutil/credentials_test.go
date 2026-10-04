package cmdutil

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

func TestAddSecretFlags_Exclusive(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		args    []string
		wantErr bool
	}{
		{args: []string{"--password", "x"}},
		{args: []string{"--password-stdin"}},
		{args: []string{"--password", "x", "--password-stdin"}, wantErr: true},
		// stdin holds one secret, so two -stdin flags cannot share it.
		{args: []string{"--password-stdin", "--decrypt-stdin"}, wantErr: true},
		{args: []string{"--password", "x", "--decrypt-stdin"}},
	} {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			t.Parallel()
			var password, decrypt string
			var passwordStdin, decryptStdin bool
			cmd := &cobra.Command{Use: "restore", RunE: func(*cobra.Command, []string) error { return nil }}
			AddSecretFlags(cmd, &password, &passwordStdin, "password", "Password", "Read the password from stdin")
			AddSecretFlagsP(cmd, &decrypt, &decryptStdin, "decrypt", "d", "Decrypt", "Read the decrypt password from stdin")
			cmd.SetArgs(tc.args)
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			if err := cmd.Execute(); (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestResolveSecret(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name, stdin, value, want, wantErr string
		fromStdin                         bool
	}{
		{name: "flag value kept without stdin", value: "flag", want: "flag"},
		{name: "one CRLF dropped", stdin: "s3cret\r\n", fromStdin: true, want: "s3cret"},
		{name: "inner newlines kept", stdin: "a\nb\n", fromStdin: true, want: "a\nb"},
		{name: "empty stdin names the flag", fromStdin: true, wantErr: "--encrypt-stdin: stdin is empty"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ios := iostreams.Test(strings.NewReader(tc.stdin), &bytes.Buffer{}, &bytes.Buffer{})
			value := tc.value
			err := ResolveSecret(ios, &value, tc.fromStdin, "encrypt", "Backup password")
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("err = %v, want %q", err, tc.wantErr)
				}
				return
			}
			if err != nil || value != tc.want {
				t.Fatalf("value = %q, err = %v; want %q", value, err, tc.want)
			}
		})
	}
}
