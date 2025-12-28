package factory_test

import (
	"context"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewTestIOStreams(t *testing.T) {
	t.Parallel()

	testIO := factory.NewTestIOStreams()

	if testIO == nil {
		t.Fatal("NewTestIOStreams returned nil")
	}
	if testIO.IOStreams == nil {
		t.Error("IOStreams is nil")
	}
	if testIO.In == nil {
		t.Error("In buffer is nil")
	}
	if testIO.Out == nil {
		t.Error("Out buffer is nil")
	}
	if testIO.ErrOut == nil {
		t.Error("ErrOut buffer is nil")
	}
}

func TestTestIO_Reset(t *testing.T) {
	t.Parallel()

	testIO := factory.NewTestIOStreams()

	// Write to buffers
	testIO.Println("stdout content")
	testIO.Errorln("stderr content")

	if testIO.OutString() == "" {
		t.Error("stdout should have content")
	}
	if testIO.ErrString() == "" {
		t.Error("stderr should have content")
	}

	// Reset and verify
	testIO.Reset()

	if testIO.OutString() != "" {
		t.Error("stdout should be empty after Reset")
	}
	if testIO.ErrString() != "" {
		t.Error("stderr should be empty after Reset")
	}
}

func TestNewTestFactory(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	if tf == nil {
		t.Fatal("NewTestFactory returned nil")
	}
	if tf.Factory == nil {
		t.Error("Factory is nil")
	}
	if tf.TestIO == nil {
		t.Error("TestIO is nil")
	}
	if tf.Config == nil {
		t.Error("Config is nil")
	}
	if tf.Manager == nil {
		t.Error("Manager is nil")
	}

	// Verify factory works
	ios := tf.IOStreams()
	if ios == nil {
		t.Error("IOStreams() returned nil")
	}

	cfg, err := tf.Factory.Config()
	if err != nil {
		t.Errorf("Config() returned error: %v", err)
	}
	if cfg == nil {
		t.Error("Config() returned nil")
	}

	svc := tf.ShellyService()
	if svc == nil {
		t.Error("ShellyService() returned nil")
	}
}

func TestNewTestFactoryWithDevices(t *testing.T) {
	t.Parallel()

	devices := map[string]model.Device{
		"dev1": {Name: "dev1", Address: "192.168.1.1"},
		"dev2": {Name: "dev2", Address: "192.168.1.2"},
	}

	tf := factory.NewTestFactoryWithDevices(t, devices)

	if len(tf.Config.Devices) != 2 {
		t.Errorf("Config.Devices = %d, want 2", len(tf.Config.Devices))
	}
	if tf.Config.Devices["dev1"].Address != "192.168.1.1" {
		t.Error("dev1 not set correctly")
	}
	if tf.Config.Devices["dev2"].Address != "192.168.1.2" {
		t.Error("dev2 not set correctly")
	}
}

func TestNewTestFactoryWithGroups(t *testing.T) {
	t.Parallel()

	groups := map[string]config.Group{
		"group1": {Devices: []string{"dev1", "dev2"}},
	}

	tf := factory.NewTestFactoryWithGroups(t, groups)

	if len(tf.Config.Groups) != 1 {
		t.Errorf("Config.Groups = %d, want 1", len(tf.Config.Groups))
	}
	if len(tf.Config.Groups["group1"].Devices) != 2 {
		t.Error("group1 not set correctly")
	}
}

func TestTestFactory_Reset(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	ios := tf.IOStreams()

	ios.Println("test output")

	if tf.OutString() == "" {
		t.Error("should have output")
	}

	tf.Reset()

	if tf.OutString() != "" {
		t.Error("should be empty after Reset")
	}
}

func TestTestFactory_OutString_ErrString(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	ios := tf.IOStreams()

	ios.Println("stdout content")
	ios.Errorln("stderr content")

	if tf.OutString() != "stdout content\n" {
		t.Errorf("OutString = %q, want %q", tf.OutString(), "stdout content\n")
	}
	if tf.ErrString() != "stderr content\n" {
		t.Errorf("ErrString = %q, want %q", tf.ErrString(), "stderr content\n")
	}
}

func TestMockBrowser(t *testing.T) {
	t.Parallel()

	mb := &factory.MockBrowser{}
	ctx := context.Background()

	// Browse should record the call
	err := mb.Browse(ctx, "https://example.com")

	if err != nil {
		t.Errorf("Browse() error = %v, want nil", err)
	}
	if !mb.BrowseCalled {
		t.Error("BrowseCalled should be true")
	}
	if mb.LastURL != "https://example.com" {
		t.Errorf("LastURL = %q, want %q", mb.LastURL, "https://example.com")
	}
}

func TestMockBrowser_WithError(t *testing.T) {
	t.Parallel()

	expectedErr := factory.ErrMock
	mb := &factory.MockBrowser{Err: expectedErr}
	ctx := context.Background()

	err := mb.Browse(ctx, "https://example.com")

	if err != expectedErr { //nolint:errorlint // exact error comparison for test
		t.Errorf("Browse() error = %v, want %v", err, expectedErr)
	}
	if !mb.BrowseCalled {
		t.Error("BrowseCalled should be true even with error")
	}
}

func TestMockBrowser_OpenDeviceUI(t *testing.T) {
	t.Parallel()

	mb := &factory.MockBrowser{}
	ctx := context.Background()

	err := mb.OpenDeviceUI(ctx, "192.168.1.100")

	if err != nil {
		t.Errorf("OpenDeviceUI() error = %v, want nil", err)
	}
	if mb.LastURL != "http://192.168.1.100" {
		t.Errorf("LastURL = %q, want %q", mb.LastURL, "http://192.168.1.100")
	}
}

func TestNewTestFactoryWithMockBrowser(t *testing.T) {
	t.Parallel()

	tf, mb := factory.NewTestFactoryWithMockBrowser(t)

	if tf == nil {
		t.Fatal("TestFactory is nil")
	}
	if mb == nil {
		t.Fatal("MockBrowser is nil")
	}

	// Verify the factory uses the mock browser
	br := tf.Browser()
	if br == nil {
		t.Error("Browser() returned nil")
	}

	ctx := context.Background()

	// Using the browser should update the mock
	if err := br.Browse(ctx, "https://test.com"); err != nil {
		t.Errorf("Browse() error = %v", err)
	}
	if !mb.BrowseCalled {
		t.Error("mock browser should have been called")
	}
	if mb.LastURL != "https://test.com" {
		t.Errorf("LastURL = %q, want %q", mb.LastURL, "https://test.com")
	}
}
