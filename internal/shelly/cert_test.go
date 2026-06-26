package shelly

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/config"
)

const testCertDir = "/test/certs"

func TestCertInstallSpec_Validate_NoFlags(t *testing.T) {
	t.Parallel()

	err := CertInstallSpec{}.Validate()
	if err == nil {
		t.Fatal("expected error when no files provided")
	}
	if !strings.Contains(err.Error(), "--ca") || !strings.Contains(err.Error(), "--client-cert") {
		t.Errorf("error should mention --ca or --client-cert, got: %v", err)
	}
}

func TestCertInstallSpec_Validate_CAOnly(t *testing.T) {
	t.Parallel()

	if err := (CertInstallSpec{CAFile: "/path/to/ca.pem"}).Validate(); err != nil {
		t.Errorf("should not error with only --ca, got: %v", err)
	}
}

func TestCertInstallSpec_Validate_ClientCertWithoutKey(t *testing.T) {
	t.Parallel()

	err := CertInstallSpec{ClientCert: "/path/to/cert.pem"}.Validate()
	if err == nil {
		t.Fatal("expected error when client cert provided without key")
	}
	if !strings.Contains(err.Error(), "--client-key") {
		t.Errorf("error should mention --client-key, got: %v", err)
	}
}

func TestCertInstallSpec_Validate_ClientCertWithKey(t *testing.T) {
	t.Parallel()

	spec := CertInstallSpec{ClientCert: "/path/to/cert.pem", ClientKey: "/path/to/key.pem"}
	if err := spec.Validate(); err != nil {
		t.Errorf("should not error with cert and key, got: %v", err)
	}
}

func TestCertInstallSpec_Validate_AllFiles(t *testing.T) {
	t.Parallel()

	spec := CertInstallSpec{
		CAFile:     "/path/to/ca.pem",
		ClientCert: "/path/to/cert.pem",
		ClientKey:  "/path/to/key.pem",
	}
	if err := spec.Validate(); err != nil {
		t.Errorf("should not error with all files, got: %v", err)
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestLoadCertData_NonexistentCAFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	config.SetFs(fs)
	t.Cleanup(func() { config.SetFs(nil) })

	_, err := loadCertData(fs, CertInstallSpec{CAFile: "/nonexistent/ca.pem"})
	if err == nil {
		t.Fatal("expected error when CA file doesn't exist")
	}
	if !strings.Contains(err.Error(), "read CA file") {
		t.Errorf("error should mention 'read CA file', got: %v", err)
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestLoadCertData_NonexistentClientCert(t *testing.T) {
	fs := afero.NewMemMapFs()
	config.SetFs(fs)
	t.Cleanup(func() { config.SetFs(nil) })

	spec := CertInstallSpec{ClientCert: "/nonexistent/cert.pem", ClientKey: "/nonexistent/key.pem"}
	_, err := loadCertData(fs, spec)
	if err == nil {
		t.Fatal("expected error when client cert file doesn't exist")
	}
	if !strings.Contains(err.Error(), "read client cert") {
		t.Errorf("error should mention 'read client cert', got: %v", err)
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestLoadCertData_ValidCAFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	config.SetFs(fs)
	t.Cleanup(func() { config.SetFs(nil) })

	caFile := testCertDir + "/ca.pem"
	caContent := []byte("-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----")
	if err := afero.WriteFile(fs, caFile, caContent, 0o600); err != nil {
		t.Fatalf("failed to create temp CA file: %v", err)
	}

	data, err := loadCertData(fs, CertInstallSpec{CAFile: caFile})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Equal(data.CAData, caContent) {
		t.Errorf("CA data mismatch: got %q, want %q", data.CAData, caContent)
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestLoadCertData_ValidClientCertAndKey(t *testing.T) {
	fs := afero.NewMemMapFs()
	config.SetFs(fs)
	t.Cleanup(func() { config.SetFs(nil) })

	certFile := testCertDir + "/cert.pem"
	keyFile := testCertDir + "/key.pem"
	certContent := []byte("-----BEGIN CERTIFICATE-----\ncert\n-----END CERTIFICATE-----")
	keyContent := []byte("-----BEGIN PRIVATE KEY-----\nkey\n-----END PRIVATE KEY-----")

	if err := afero.WriteFile(fs, certFile, certContent, 0o600); err != nil {
		t.Fatalf("failed to create temp cert file: %v", err)
	}
	if err := afero.WriteFile(fs, keyFile, keyContent, 0o600); err != nil {
		t.Fatalf("failed to create temp key file: %v", err)
	}

	data, err := loadCertData(fs, CertInstallSpec{ClientCert: certFile, ClientKey: keyFile})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Equal(data.CertData, certContent) {
		t.Errorf("Cert data mismatch: got %q, want %q", data.CertData, certContent)
	}
	if !bytes.Equal(data.KeyData, keyContent) {
		t.Errorf("Key data mismatch: got %q, want %q", data.KeyData, keyContent)
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestLoadCertData_NonexistentClientKey(t *testing.T) {
	fs := afero.NewMemMapFs()
	config.SetFs(fs)
	t.Cleanup(func() { config.SetFs(nil) })

	certFile := testCertDir + "/cert.pem"
	certContent := []byte("-----BEGIN CERTIFICATE-----\ncert\n-----END CERTIFICATE-----")
	if err := afero.WriteFile(fs, certFile, certContent, 0o600); err != nil {
		t.Fatalf("failed to create temp cert file: %v", err)
	}

	spec := CertInstallSpec{ClientCert: certFile, ClientKey: "/nonexistent/key.pem"}
	_, err := loadCertData(fs, spec)
	if err == nil {
		t.Fatal("expected error when client key file doesn't exist")
	}
	if !strings.Contains(err.Error(), "read client key") {
		t.Errorf("error should mention 'read client key', got: %v", err)
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestLoadCertData_NoFiles(t *testing.T) {
	fs := afero.NewMemMapFs()
	config.SetFs(fs)
	t.Cleanup(func() { config.SetFs(nil) })

	data, err := loadCertData(fs, CertInstallSpec{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data.CAData) != 0 {
		t.Errorf("expected empty CA data, got %d bytes", len(data.CAData))
	}
	if len(data.CertData) != 0 {
		t.Errorf("expected empty cert data, got %d bytes", len(data.CertData))
	}
	if len(data.KeyData) != 0 {
		t.Errorf("expected empty key data, got %d bytes", len(data.KeyData))
	}
}
