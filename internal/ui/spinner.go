// Package ui provides user interaction components for the CLI.
package ui

import (
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/spf13/viper"

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
	charSet := spinner.CharSets[14] // ⣾⣽⣻⢿⡿⣟⣯⣷

	s := &Spinner{
		s: spinner.New(charSet, 100*time.Millisecond, spinner.WithWriter(os.Stderr)),
	}

	//nolint:errcheck // Spinner color is best-effort
	s.s.Color("cyan")
	s.s.Suffix = " " + message

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Start starts the spinner if not in quiet mode.
func (s *Spinner) Start() {
	if !viper.GetBool("quiet") {
		s.s.Start()
	}
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

// UpdateMessage updates the spinner message.
func (s *Spinner) UpdateMessage(message string) {
	s.s.Suffix = " " + message
}

// WithSpinner runs a function with a spinner, handling start/stop.
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
