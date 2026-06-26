// Package backup provides backup and restore operations for Shelly devices.
package backup

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	shellybackup "github.com/tj-smith47/shelly-go/backup"
)

// ErrEncryptedNeedsPassword is returned by Load when the data is an encrypted
// backup envelope but no decryption password was supplied. Callers surface it as
// the "use --decrypt" hint rather than a generic parse failure.
var ErrEncryptedNeedsPassword = errors.New("backup is encrypted: a password is required to decrypt it")

// encryptedEnvelope is the minimal projection of the shelly-go EncryptedBackup
// wire format used to recognise an encrypted backup. Detection keys on a
// non-empty encrypted_data field, which a plaintext backup never carries.
type encryptedEnvelope struct {
	EncryptedData string `json:"encrypted_data"`
}

// IsEncrypted reports whether raw backup bytes are an encrypted envelope rather
// than a plaintext backup. It is content-based, not extension-based, so an
// encrypted backup is recognised regardless of how the file was named.
func IsEncrypted(data []byte) bool {
	var env encryptedEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return false
	}
	return env.EncryptedData != ""
}

// EncryptedInfo holds the cleartext metadata an encrypted backup exposes without
// the password, used to describe encrypted files in listings and validation.
type EncryptedInfo struct {
	CreatedAt   time.Time
	DeviceID    string
	DeviceModel string
}

// ReadEncryptedInfo extracts the cleartext envelope metadata from encrypted
// backup bytes. It does not decrypt anything.
func ReadEncryptedInfo(data []byte) (EncryptedInfo, error) {
	var env shellybackup.EncryptedBackup
	if err := json.Unmarshal(data, &env); err != nil {
		return EncryptedInfo{}, fmt.Errorf("invalid encrypted backup: %w", err)
	}
	return EncryptedInfo{
		CreatedAt:   env.CreatedAt,
		DeviceID:    env.DeviceID,
		DeviceModel: env.DeviceModel,
	}, nil
}

// Encrypt wraps a plaintext backup in an AES-256-GCM encrypted envelope keyed by
// password, returning indented JSON ready to write to disk. It marks bkp as
// encrypted so a post-write summary reports the file's protected state.
func Encrypt(bkp *DeviceBackup, password string) ([]byte, error) {
	plain, err := json.Marshal(bkp.Backup)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal backup: %w", err)
	}

	enc := shellybackup.NewEncryptor(password)
	encoded, err := enc.EncryptToBase64(plain)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt backup: %w", err)
	}

	env := shellybackup.EncryptedBackup{
		Version:       shellybackup.EncryptedBackupVersion,
		CreatedAt:     time.Now().UTC(),
		EncryptedData: encoded,
	}
	if dev := bkp.Device(); dev.ID != "" || dev.Model != "" {
		env.DeviceID = dev.ID
		env.DeviceModel = dev.Model
	}

	out, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal encrypted backup: %w", err)
	}

	bkp.encrypted = true
	return out, nil
}

// Load decodes backup bytes into a DeviceBackup, transparently decrypting an
// encrypted envelope when password is non-empty. It returns
// ErrEncryptedNeedsPassword when the data is encrypted but no password is given,
// and wraps a wrong-password failure with a recognisable hint.
func Load(data []byte, password string) (*DeviceBackup, error) {
	if !IsEncrypted(data) {
		return Validate(data)
	}

	if password == "" {
		return nil, ErrEncryptedNeedsPassword
	}

	var env shellybackup.EncryptedBackup
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("invalid encrypted backup: %w", err)
	}

	enc := shellybackup.NewEncryptor(password)
	plain, err := enc.DecryptFromBase64(env.EncryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt backup (wrong password?): %w", err)
	}

	bkp, err := Validate(plain)
	if err != nil {
		return nil, err
	}
	bkp.encrypted = true
	return bkp, nil
}
