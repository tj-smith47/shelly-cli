// Package output provides output formatting utilities for the CLI.
package output

import (
	"os"
	"time"

	"github.com/briandowns/spinner"

	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Spinner wraps briandowns/spinner with theme integration.
type Spinner struct {
	s *spinner.Spinner
}

// SpinnerOption configures a spinner.
type SpinnerOption func(*Spinner)

// NewSpinner creates a new spinner with the given message.
func NewSpinner(message string, opts ...SpinnerOption) *Spinner {
	// Use a nice character set that works across terminals
	charSet := spinner.CharSets[14] // ⣾⣽⣻⢿⡿⣟⣯⣷

	s := &Spinner{
		s: spinner.New(charSet, 100*time.Millisecond, spinner.WithWriter(os.Stderr)),
	}

	// Apply theme color.
	//nolint:errcheck // Spinner color is best-effort
	s.s.Color("cyan")

	// Set message
	s.s.Suffix = " " + message

	// Apply options
	for _, opt := range opts {
		opt(s)
	}

	return s
}

// WithSuffix sets the spinner suffix (message after the spinner).
func WithSuffix(suffix string) SpinnerOption {
	return func(s *Spinner) {
		s.s.Suffix = " " + suffix
	}
}

// WithPrefix sets the spinner prefix (message before the spinner).
func WithPrefix(prefix string) SpinnerOption {
	return func(s *Spinner) {
		s.s.Prefix = prefix + " "
	}
}

// WithFinalMessage sets the message shown when the spinner stops.
func WithFinalMessage(msg string) SpinnerOption {
	return func(s *Spinner) {
		s.s.FinalMSG = msg + "\n"
	}
}

// WithCharSet sets the spinner character set.
func WithCharSet(charSet []string) SpinnerOption {
	return func(s *Spinner) {
		s.s.UpdateCharSet(charSet)
	}
}

// Start starts the spinner.
func (s *Spinner) Start() {
	s.s.Start()
}

// Stop stops the spinner.
func (s *Spinner) Stop() {
	s.s.Stop()
}

// StopWithSuccess stops the spinner and shows a success message.
func (s *Spinner) StopWithSuccess(message string) {
	s.s.FinalMSG = theme.StatusOK().Render("✓") + " " + message + "\n"
	s.s.Stop()
}

// StopWithError stops the spinner and shows an error message.
func (s *Spinner) StopWithError(message string) {
	s.s.FinalMSG = theme.StatusError().Render("✗") + " " + message + "\n"
	s.s.Stop()
}

// StopWithWarning stops the spinner and shows a warning message.
func (s *Spinner) StopWithWarning(message string) {
	s.s.FinalMSG = theme.StatusWarn().Render("⚠") + " " + message + "\n"
	s.s.Stop()
}

// UpdateMessage updates the spinner message.
func (s *Spinner) UpdateMessage(message string) {
	s.s.Suffix = " " + message
}

// Active returns whether the spinner is currently active.
func (s *Spinner) Active() bool {
	return s.s.Active()
}

// Reverse reverses the spinner direction.
func (s *Spinner) Reverse() {
	s.s.Reverse()
}

// Common spinner presets

// DiscoverySpinner creates a spinner for device discovery.
func DiscoverySpinner() *Spinner {
	return NewSpinner("Discovering devices...")
}

// FirmwareSpinner creates a spinner for firmware operations.
func FirmwareSpinner(action string) *Spinner {
	return NewSpinner(action + " firmware...")
}

// BackupSpinner creates a spinner for backup operations.
func BackupSpinner(action string) *Spinner {
	return NewSpinner(action + " backup...")
}

// ConnectingSpinner creates a spinner for connection operations.
func ConnectingSpinner(target string) *Spinner {
	return NewSpinner("Connecting to " + target + "...")
}

// LoadingSpinner creates a generic loading spinner.
func LoadingSpinner(message string) *Spinner {
	return NewSpinner(message)
}

// WithSpinner runs a function with a spinner, handling start/stop and errors.
func WithSpinner(message string, fn func() error) error {
	s := NewSpinner(message)
	s.Start()
	defer s.Stop()

	err := fn()
	if err != nil {
		s.StopWithError(err.Error())
		return err
	}
	s.StopWithSuccess("Done")
	return nil
}

// WithSpinnerResult runs a function with a spinner and returns the result.
func WithSpinnerResult[T any](message string, fn func() (T, error)) (T, error) {
	s := NewSpinner(message)
	s.Start()
	defer s.Stop()

	result, err := fn()
	if err != nil {
		s.StopWithError(err.Error())
		return result, err
	}
	s.StopWithSuccess("Done")
	return result, nil
}
