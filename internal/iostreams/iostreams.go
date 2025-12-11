// Package iostreams provides unified I/O handling for the CLI.
// It follows the gh CLI iostreams pattern, providing a consistent abstraction
// for terminal I/O, TTY detection, color management, and progress indicators.
package iostreams

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/mattn/go-isatty"
	"github.com/spf13/viper"
)

// logVerbose logs a message to stderr only if verbose mode is enabled.
func logVerbose(format string, args ...any) {
	if viper.GetBool("verbose") {
		fmt.Fprintf(os.Stderr, "debug: "+format+"\n", args...)
	}
}

// writeQuietly writes to a writer and logs any error in verbose mode.
// Use this for best-effort terminal output where errors are not critical.
func writeQuietly(w io.Writer, format string, args ...any) {
	if _, err := fmt.Fprintf(w, format, args...); err != nil {
		logVerbose("write error: %v", err)
	}
}

// writelnQuietly writes a line to a writer and logs any error in verbose mode.
func writelnQuietly(w io.Writer, args ...any) {
	if _, err := fmt.Fprintln(w, args...); err != nil {
		logVerbose("write error: %v", err)
	}
}

// IOStreams holds I/O streams and terminal state.
type IOStreams struct {
	In     io.Reader
	Out    io.Writer
	ErrOut io.Writer

	// Terminal state (detected once)
	isStdinTTY  bool
	isStdoutTTY bool
	isStderrTTY bool

	// Color settings
	colorEnabled bool
	colorForced  bool

	// Progress indicator (protected by mutex)
	progressMu        sync.Mutex
	progressIndicator *Spinner

	// Quiet mode - suppress non-essential output
	quiet bool
}

// System creates IOStreams connected to stdin/stdout/stderr.
func System() *IOStreams {
	ios := &IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	// Detect TTY - os.Stdin/Stdout/Stderr are already *os.File
	ios.isStdinTTY = isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd())
	ios.isStdoutTTY = isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
	ios.isStderrTTY = isatty.IsTerminal(os.Stderr.Fd()) || isatty.IsCygwinTerminal(os.Stderr.Fd())

	// Determine color settings
	ios.colorEnabled = ios.isStdoutTTY && !isColorDisabled()
	ios.colorForced = isColorForced()
	if ios.colorForced {
		ios.colorEnabled = true
	}

	// Check quiet mode from viper
	ios.quiet = viper.GetBool("quiet")

	return ios
}

// Test creates IOStreams for testing with provided buffers.
func Test(in io.Reader, out, errOut io.Writer) *IOStreams {
	return &IOStreams{
		In:           in,
		Out:          out,
		ErrOut:       errOut,
		isStdinTTY:   false,
		isStdoutTTY:  false,
		isStderrTTY:  false,
		colorEnabled: false,
		quiet:        false,
	}
}

// isColorDisabled checks environment variables for color disable flags.
func isColorDisabled() bool {
	// Check NO_COLOR (https://no-color.org/)
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return true
	}
	// Check SHELLY_NO_COLOR
	if _, ok := os.LookupEnv("SHELLY_NO_COLOR"); ok {
		return true
	}
	// Check TERM=dumb
	if os.Getenv("TERM") == "dumb" {
		return true
	}
	return false
}

// isColorForced checks if color is explicitly forced.
func isColorForced() bool {
	// Check FORCE_COLOR
	if _, ok := os.LookupEnv("FORCE_COLOR"); ok {
		return true
	}
	// Check SHELLY_FORCE_COLOR
	if _, ok := os.LookupEnv("SHELLY_FORCE_COLOR"); ok {
		return true
	}
	return false
}

// IsStdinTTY returns true if stdin is a terminal.
func (s *IOStreams) IsStdinTTY() bool {
	return s.isStdinTTY
}

// IsStdoutTTY returns true if stdout is a terminal.
func (s *IOStreams) IsStdoutTTY() bool {
	return s.isStdoutTTY
}

// IsStderrTTY returns true if stderr is a terminal.
func (s *IOStreams) IsStderrTTY() bool {
	return s.isStderrTTY
}

// ColorEnabled returns true if color output is enabled.
func (s *IOStreams) ColorEnabled() bool {
	return s.colorEnabled
}

// SetColorEnabled explicitly sets the color enabled state.
func (s *IOStreams) SetColorEnabled(enabled bool) {
	s.colorEnabled = enabled
}

// IsQuiet returns true if quiet mode is enabled.
func (s *IOStreams) IsQuiet() bool {
	return s.quiet
}

// SetQuiet sets the quiet mode.
func (s *IOStreams) SetQuiet(quiet bool) {
	s.quiet = quiet
}

// SetStdinTTY sets the stdin TTY state (for testing).
func (s *IOStreams) SetStdinTTY(isTTY bool) {
	s.isStdinTTY = isTTY
}

// SetStdoutTTY sets the stdout TTY state (for testing).
func (s *IOStreams) SetStdoutTTY(isTTY bool) {
	s.isStdoutTTY = isTTY
}

// SetStderrTTY sets the stderr TTY state (for testing).
func (s *IOStreams) SetStderrTTY(isTTY bool) {
	s.isStderrTTY = isTTY
}

// StartProgress starts a spinner with the given message.
// If stdout is not a TTY, it prints the message once instead.
func (s *IOStreams) StartProgress(msg string) {
	s.progressMu.Lock()
	defer s.progressMu.Unlock()

	if s.quiet {
		return
	}

	if !s.isStderrTTY {
		// No spinner for non-TTY, just print message
		writelnQuietly(s.ErrOut, msg)
		return
	}

	s.progressIndicator = NewSpinnerWithWriter(msg, s.ErrOut)
	s.progressIndicator.Start()
}

// StopProgress stops the current spinner.
func (s *IOStreams) StopProgress() {
	s.progressMu.Lock()
	defer s.progressMu.Unlock()

	if s.progressIndicator != nil {
		s.progressIndicator.Stop()
		s.progressIndicator = nil
	}
}

// StopProgressWithSuccess stops the spinner with a success message.
func (s *IOStreams) StopProgressWithSuccess(msg string) {
	s.progressMu.Lock()
	defer s.progressMu.Unlock()

	if s.progressIndicator != nil {
		s.progressIndicator.StopWithSuccess(msg)
		s.progressIndicator = nil
	}
}

// StopProgressWithError stops the spinner with an error message.
func (s *IOStreams) StopProgressWithError(msg string) {
	s.progressMu.Lock()
	defer s.progressMu.Unlock()

	if s.progressIndicator != nil {
		s.progressIndicator.StopWithError(msg)
		s.progressIndicator = nil
	}
}

// UpdateProgress updates the current spinner message.
func (s *IOStreams) UpdateProgress(msg string) {
	s.progressMu.Lock()
	defer s.progressMu.Unlock()

	if s.progressIndicator != nil {
		s.progressIndicator.UpdateMessage(msg)
	}
}

// CanPrompt returns true if the terminal supports interactive prompts.
func (s *IOStreams) CanPrompt() bool {
	return s.isStdinTTY && s.isStdoutTTY
}

// Printf writes formatted output to stdout.
func (s *IOStreams) Printf(format string, args ...any) {
	writeQuietly(s.Out, format, args...)
}

// Println writes a line to stdout.
func (s *IOStreams) Println(args ...any) {
	writelnQuietly(s.Out, args...)
}

// Errorf writes formatted output to stderr.
func (s *IOStreams) Errorf(format string, args ...any) {
	writeQuietly(s.ErrOut, format, args...)
}

// Errorln writes a line to stderr.
func (s *IOStreams) Errorln(args ...any) {
	writelnQuietly(s.ErrOut, args...)
}
