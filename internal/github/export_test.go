package github

import (
	"context"
	"io"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// CopyFile exports copyFile for testing.
func CopyFile(ios *iostreams.IOStreams, src, dst string) error {
	return copyFile(ios, src, dst)
}

// ExtractTarGz exports extractTarGz for testing.
func (c *Client) ExtractTarGz(archivePath, destDir, binaryName string) (string, error) {
	return c.extractTarGz(archivePath, destDir, binaryName)
}

// ExtractZip exports extractZip for testing.
func (c *Client) ExtractZip(archivePath, destDir, binaryName string) (string, error) {
	return c.extractZip(archivePath, destDir, binaryName)
}

// MatchesBinaryName exports matchesBinaryName for testing.
func (c *Client) MatchesBinaryName(filename, binaryName string) bool {
	return c.matchesBinaryName(filename, binaryName)
}

// SetAPIBaseURL allows tests to override the GitHub API base URL.
func SetAPIBaseURL(url string) func() {
	old := GitHubAPIBaseURL
	GitHubAPIBaseURL = url
	return func() { GitHubAPIBaseURL = old }
}

// CreateBackup exports createBackup for testing.
func CreateBackup(ios *iostreams.IOStreams, targetPath, backupPath string) error {
	return createBackup(ios, targetPath, backupPath)
}

// RestoreFromBackup exports restoreFromBackup for testing.
func RestoreFromBackup(backupPath, targetPath string, writeErr error) error {
	return restoreFromBackup(backupPath, targetPath, writeErr)
}

// SetOsExecutable sets the osExecutable function for testing.
func SetOsExecutable(fn func() (string, error)) func() {
	old := osExecutable
	osExecutable = fn
	return func() { osExecutable = old }
}

// SetEvalSymlinks sets the evalSymlinks function for testing.
func SetEvalSymlinks(fn func(string) (string, error)) func() {
	old := evalSymlinks
	evalSymlinks = fn
	return func() { evalSymlinks = old }
}

// SetExecCommandStart sets the execCommandStart function for testing.
func SetExecCommandStart(fn func(ctx context.Context, path string, args []string) error) func() {
	old := execCommandStart
	execCommandStart = fn
	return func() { execCommandStart = old }
}

// SetRuntimeGOOS sets the runtimeGOOS variable for testing.
func SetRuntimeGOOS(goos string) func() {
	old := runtimeGOOS
	runtimeGOOS = goos
	return func() { runtimeGOOS = old }
}

// ExtractToFile exports extractToFile for testing.
func (c *Client) ExtractToFile(destPath string, r io.Reader) error {
	return c.extractToFile(destPath, r)
}
