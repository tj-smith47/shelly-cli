// Package iostreams provides unified I/O handling for the CLI.
package iostreams

import (
	"io"
	"os"
	"time"

	"github.com/briandowns/spinner"

	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// setSpinnerColorQuietly sets the spinner color and logs any error in verbose mode.
// Spinner color is a best-effort operation - failure is acceptable.
func setSpinnerColorQuietly(s *spinner.Spinner, color string) {
	if err := s.Color(color); err != nil {
		logVerbose("failed to set spinner color: %v", err)
	}
}

// Spinner wraps briandowns/spinner with theme integration.
type Spinner struct {
	s      *spinner.Spinner
	writer io.Writer
}

// SpinnerOption configures a spinner.
type SpinnerOption func(*Spinner)

// NewSpinner creates a new spinner with the given message.
func NewSpinner(message string, opts ...SpinnerOption) *Spinner {
	return NewSpinnerWithWriter(message, os.Stderr, opts...)
}

// NewSpinnerWithWriter creates a new spinner with a custom writer.
func NewSpinnerWithWriter(message string, w io.Writer, opts ...SpinnerOption) *Spinner {
	// Use braille dot pattern spinner (CharSets[11])
	s := &Spinner{
		s:      spinner.New(spinner.CharSets[11], 100*time.Millisecond, spinner.WithWriter(w)),
		writer: w,
	}

	// Apply shelly blue (bright cyan) for spinners
	setSpinnerColorQuietly(s.s, "fgHiCyan")

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

// WithColor sets the spinner color.
func WithColor(color string) SpinnerOption {
	return func(s *Spinner) {
		setSpinnerColorQuietly(s.s, color)
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
