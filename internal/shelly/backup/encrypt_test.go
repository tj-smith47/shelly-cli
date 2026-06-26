package backup

import (
	"encoding/json"
	"errors"
	"testing"

	shellybackup "github.com/tj-smith47/shelly-go/backup"
)

func sampleBackup() *DeviceBackup {
	return &DeviceBackup{
		Backup: &shellybackup.Backup{
			Version: shellybackup.BackupVersion,
			DeviceInfo: &shellybackup.DeviceInfo{
				ID:         "shellyplus1pm-abc123",
				Model:      "SNSW-001P16EU",
				Generation: 2,
			},
			Config: json.RawMessage(`{"switch:0":{"id":0,"name":"main"}}`),
		},
	}
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	t.Parallel()

	const password = "correct horse battery staple"
	bkp := sampleBackup()

	data, err := Encrypt(bkp, password)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if !bkp.Encrypted() {
		t.Error("Encrypt should mark the backup as encrypted")
	}
	if !IsEncrypted(data) {
		t.Fatal("Encrypt output should be detected as an encrypted envelope")
	}

	// The ciphertext must not leak the plaintext config.
	if json.Valid(data) {
		var env map[string]any
		if unErr := json.Unmarshal(data, &env); unErr == nil {
			if _, leaked := env["config"]; leaked {
				t.Error("encrypted envelope must not carry a cleartext config field")
			}
		}
	}

	got, err := Load(data, password)
	if err != nil {
		t.Fatalf("Load with correct password: %v", err)
	}
	if !got.Encrypted() {
		t.Error("Load should mark a decrypted backup as encrypted")
	}
	if got.Device().ID != bkp.Device().ID {
		t.Errorf("round-trip device ID = %q, want %q", got.Device().ID, bkp.Device().ID)
	}
	if string(got.Config) != string(bkp.Config) {
		t.Errorf("round-trip config = %q, want %q", got.Config, bkp.Config)
	}
}

func TestLoadEncryptedWithoutPassword(t *testing.T) {
	t.Parallel()

	data, err := Encrypt(sampleBackup(), "secret")
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	_, err = Load(data, "")
	if !errors.Is(err, ErrEncryptedNeedsPassword) {
		t.Errorf("Load without password = %v, want ErrEncryptedNeedsPassword", err)
	}
}

func TestLoadEncryptedWrongPassword(t *testing.T) {
	t.Parallel()

	data, err := Encrypt(sampleBackup(), "secret")
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	if _, err = Load(data, "wrong"); err == nil {
		t.Error("Load with wrong password should fail")
	}
}

func TestLoadPlaintext(t *testing.T) {
	t.Parallel()

	bkp := sampleBackup()
	data, err := json.Marshal(bkp.Backup)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if IsEncrypted(data) {
		t.Error("plaintext backup must not be detected as encrypted")
	}

	got, err := Load(data, "")
	if err != nil {
		t.Fatalf("Load plaintext: %v", err)
	}
	if got.Encrypted() {
		t.Error("a plaintext backup must not be reported as encrypted")
	}
	if got.Device().ID != bkp.Device().ID {
		t.Errorf("device ID = %q, want %q", got.Device().ID, bkp.Device().ID)
	}
}

func TestReadEncryptedInfo(t *testing.T) {
	t.Parallel()

	bkp := sampleBackup()
	data, err := Encrypt(bkp, "secret")
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	info, err := ReadEncryptedInfo(data)
	if err != nil {
		t.Fatalf("ReadEncryptedInfo: %v", err)
	}
	if info.DeviceID != bkp.Device().ID {
		t.Errorf("envelope device ID = %q, want %q", info.DeviceID, bkp.Device().ID)
	}
	if info.DeviceModel != bkp.Device().Model {
		t.Errorf("envelope device model = %q, want %q", info.DeviceModel, bkp.Device().Model)
	}
	if info.CreatedAt.IsZero() {
		t.Error("envelope CreatedAt should be set")
	}
}

func TestIsEncryptedGarbage(t *testing.T) {
	t.Parallel()

	if IsEncrypted([]byte("not json at all")) {
		t.Error("non-JSON data must not be detected as encrypted")
	}
}
