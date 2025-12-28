// Package factory provides test factory utilities for creating cmdutil.Factory with test dependencies.
// This package is separate from the main testutil package to avoid import cycles.
package factory

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/browser"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// ErrMock is a standard mock error for testing.
var ErrMock = errors.New("mock error")

// TestIO holds test IOStreams and buffers for assertions.
type TestIO struct {
	*iostreams.IOStreams
	In     *bytes.Buffer
	Out    *bytes.Buffer
	ErrOut *bytes.Buffer
}

// NewTestIOStreams creates IOStreams for testing with associated buffers.
// Returns the IOStreams and the underlying buffers for assertions.
func NewTestIOStreams() *TestIO {
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)
	return &TestIO{
		IOStreams: ios,
		In:        in,
		Out:       out,
		ErrOut:    errOut,
	}
}

// Reset clears all buffers.
func (t *TestIO) Reset() {
	t.In.Reset()
	t.Out.Reset()
	t.ErrOut.Reset()
}

// OutString returns stdout content as string.
func (t *TestIO) OutString() string {
	return t.Out.String()
}

// ErrString returns stderr content as string.
func (t *TestIO) ErrString() string {
	return t.ErrOut.String()
}

// TestFactory holds a test Factory with associated test IO and config.
type TestFactory struct {
	*cmdutil.Factory
	TestIO  *TestIO
	Config  *config.Config
	Manager *config.Manager
}

// NewTestFactory creates a Factory with test dependencies.
// The factory is pre-configured with test IOStreams and an empty config.
// The ShellyService is lazily initialized by the factory when first accessed.
func NewTestFactory(t *testing.T) *TestFactory {
	t.Helper()

	testIO := NewTestIOStreams()
	cfg := &config.Config{
		Devices:   make(map[string]model.Device),
		Groups:    make(map[string]config.Group),
		Aliases:   make(map[string]config.Alias),
		Scenes:    make(map[string]config.Scene),
		Templates: config.TemplatesConfig{},
	}
	mgr := config.NewTestManager(cfg)

	// ShellyService is lazily initialized by factory when first accessed
	f := cmdutil.NewFactory().
		SetIOStreams(testIO.IOStreams).
		SetConfigManager(mgr)

	return &TestFactory{
		Factory: f,
		TestIO:  testIO,
		Config:  cfg,
		Manager: mgr,
	}
}

// NewTestFactoryWithDevices creates a test factory with pre-registered devices.
func NewTestFactoryWithDevices(t *testing.T, devices map[string]model.Device) *TestFactory {
	t.Helper()

	tf := NewTestFactory(t)
	for k, v := range devices {
		tf.Config.Devices[k] = v
	}
	return tf
}

// NewTestFactoryWithGroups creates a test factory with pre-registered groups.
func NewTestFactoryWithGroups(t *testing.T, groups map[string]config.Group) *TestFactory {
	t.Helper()

	tf := NewTestFactory(t)
	for k, v := range groups {
		tf.Config.Groups[k] = v
	}
	return tf
}

// Reset clears all IO buffers.
func (tf *TestFactory) Reset() {
	tf.TestIO.Reset()
}

// OutString returns stdout content as string.
func (tf *TestFactory) OutString() string {
	return tf.TestIO.OutString()
}

// ErrString returns stderr content as string.
func (tf *TestFactory) ErrString() string {
	return tf.TestIO.ErrString()
}

// MockBrowser implements browser.Browser for testing.
type MockBrowser struct {
	BrowseCalled bool
	LastURL      string
	Err          error
}

// Browse records the call and returns the mock error.
func (m *MockBrowser) Browse(_ context.Context, url string) error {
	m.BrowseCalled = true
	m.LastURL = url
	return m.Err
}

// OpenDeviceUI opens a Shelly device's web interface by IP address.
func (m *MockBrowser) OpenDeviceUI(ctx context.Context, deviceIP string) error {
	return m.Browse(ctx, "http://"+deviceIP)
}

// NewTestFactoryWithMockBrowser creates a test factory with a mock browser.
func NewTestFactoryWithMockBrowser(t *testing.T) (*TestFactory, *MockBrowser) {
	t.Helper()

	tf := NewTestFactory(t)
	mb := &MockBrowser{}
	tf.SetBrowser(mb)
	return tf, mb
}

// Ensure MockBrowser implements browser.Browser.
var _ browser.Browser = (*MockBrowser)(nil)
