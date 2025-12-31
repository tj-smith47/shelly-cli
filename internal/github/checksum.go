package github

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// VerifyChecksum verifies the checksum of a downloaded binary against a checksum asset.
func (c *Client) VerifyChecksum(ctx context.Context, ios *iostreams.IOStreams, binaryPath, assetName string, checksumAsset *Asset) error {
	fs := getFs()

	// Calculate checksum of downloaded file
	f, err := fs.Open(binaryPath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			ios.DebugErr("closing binary file", cerr)
		}
	}()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return fmt.Errorf("calculate hash: %w", err)
	}
	actualHash := hex.EncodeToString(hasher.Sum(nil))

	// Download and parse checksum file
	tmpDir, err := afero.TempDir(fs, "", "shelly-checksum-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer func() {
		if rerr := fs.RemoveAll(tmpDir); rerr != nil {
			ios.DebugErr("removing temp dir", rerr)
		}
	}()

	checksumPath := filepath.Join(tmpDir, checksumAsset.Name)
	if err := c.DownloadAsset(ctx, checksumAsset, checksumPath); err != nil {
		return fmt.Errorf("download checksum: %w", err)
	}

	content, err := afero.ReadFile(fs, checksumPath)
	if err != nil {
		return fmt.Errorf("read checksum: %w", err)
	}

	expectedHash, err := ParseChecksumFile(string(content), assetName)
	if err != nil {
		return err
	}

	if !strings.EqualFold(actualHash, expectedHash) {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedHash, actualHash)
	}

	return nil
}

// ParseChecksumFile parses a checksum file and returns the hash for the specified asset.
func ParseChecksumFile(content, assetName string) (string, error) {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Format: "hash  filename" or "hash filename"
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			hash := parts[0]
			filename := parts[len(parts)-1]

			// Handle "*filename" format (binary mode indicator)
			filename = strings.TrimPrefix(filename, "*")

			if strings.EqualFold(filepath.Base(filename), assetName) {
				return hash, nil
			}
		} else if len(parts) == 1 {
			// Single hash (assume it's for this file)
			return parts[0], nil
		}
	}

	return "", fmt.Errorf("checksum not found for %s in checksum file", assetName)
}
