package cmdutil_test

import (
	"bytes"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

func TestNew(t *testing.T) {
	t.Parallel()

	f := cmdutil.New()
	if f == nil {
		t.Fatal("New() returned nil")
	}

	// Verify functions are set
	if f.IOStreams == nil {
		t.Error("IOStreams function is nil")
	}
	if f.Config == nil {
		t.Error("Config function is nil")
	}
	if f.ShellyService == nil {
		t.Error("ShellyService function is nil")
	}
}

func TestFactory_IOStreams_LazyInit(t *testing.T) {
	t.Parallel()

	f := cmdutil.New()

	// First call should initialize
	ios1 := f.IOStreams()
	if ios1 == nil {
		t.Fatal("IOStreams() returned nil")
	}

	// Second call should return same instance
	ios2 := f.IOStreams()
	if ios1 != ios2 {
		t.Error("IOStreams() should return cached instance")
	}
}

func TestFactory_ShellyService_LazyInit(t *testing.T) {
	t.Parallel()

	f := cmdutil.New()

	// First call should initialize
	svc1 := f.ShellyService()
	if svc1 == nil {
		t.Fatal("ShellyService() returned nil")
	}

	// Second call should return same instance
	svc2 := f.ShellyService()
	if svc1 != svc2 {
		t.Error("ShellyService() should return cached instance")
	}
}

func TestNewWithIOStreams(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewWithIOStreams(ios)
	if f == nil {
		t.Fatal("NewWithIOStreams() returned nil")
	}

	// Should return the custom IOStreams
	gotIOS := f.IOStreams()
	if gotIOS != ios {
		t.Error("NewWithIOStreams should use provided IOStreams")
	}
}

func TestFactory_SetIOStreams(t *testing.T) {
	t.Parallel()

	f := cmdutil.New()

	// Create custom IOStreams
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	// Set custom IOStreams
	result := f.SetIOStreams(ios)
	if result != f {
		t.Error("SetIOStreams should return factory for chaining")
	}

	// Should return the custom IOStreams
	gotIOS := f.IOStreams()
	if gotIOS != ios {
		t.Error("SetIOStreams should set the IOStreams")
	}
}

func TestFactory_SetConfig(t *testing.T) {
	t.Parallel()

	f := cmdutil.New()

	// Create custom config
	cfg := &config.Config{}

	// Set custom config
	result := f.SetConfig(cfg)
	if result != f {
		t.Error("SetConfig should return factory for chaining")
	}

	// Should return the custom config
	gotCfg, err := f.Config()
	if err != nil {
		t.Fatalf("Config() returned error: %v", err)
	}
	if gotCfg != cfg {
		t.Error("SetConfig should set the config")
	}
}

func TestFactory_SetShellyService(t *testing.T) {
	t.Parallel()

	f := cmdutil.New()

	// Create custom service
	svc := shelly.NewService()

	// Set custom service
	result := f.SetShellyService(svc)
	if result != f {
		t.Error("SetShellyService should return factory for chaining")
	}

	// Should return the custom service
	gotSvc := f.ShellyService()
	if gotSvc != svc {
		t.Error("SetShellyService should set the service")
	}
}

func TestFactory_Chaining(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)
	cfg := &config.Config{}
	svc := shelly.NewService()

	f := cmdutil.New().
		SetIOStreams(ios).
		SetConfig(cfg).
		SetShellyService(svc)

	if f.IOStreams() != ios {
		t.Error("Chained SetIOStreams failed")
	}
	gotCfg, err := f.Config()
	if err != nil {
		t.Fatalf("Config() error: %v", err)
	}
	if gotCfg != cfg {
		t.Error("Chained SetConfig failed")
	}
	if f.ShellyService() != svc {
		t.Error("Chained SetShellyService failed")
	}
}

func TestFactory_MustConfig(t *testing.T) {
	t.Parallel()

	f := cmdutil.New()
	cfg := &config.Config{}
	f.SetConfig(cfg)

	// MustConfig should return config when no error
	gotCfg := f.MustConfig()
	if gotCfg != cfg {
		t.Error("MustConfig should return the config")
	}
}

func TestFactory_SetIOStreams_OverridesOriginal(t *testing.T) {
	t.Parallel()

	// Create factory with initial IOStreams
	in1 := &bytes.Buffer{}
	out1 := &bytes.Buffer{}
	errOut1 := &bytes.Buffer{}
	ios1 := iostreams.Test(in1, out1, errOut1)
	f := cmdutil.NewWithIOStreams(ios1)

	// Override with new IOStreams
	in2 := &bytes.Buffer{}
	out2 := &bytes.Buffer{}
	errOut2 := &bytes.Buffer{}
	ios2 := iostreams.Test(in2, out2, errOut2)
	f.SetIOStreams(ios2)

	// Should return the new IOStreams
	gotIOS := f.IOStreams()
	if gotIOS != ios2 {
		t.Error("SetIOStreams should override previous IOStreams")
	}
}

func TestFactory_SetConfig_OverridesOriginal(t *testing.T) {
	t.Parallel()

	f := cmdutil.New()

	// Set first config
	cfg1 := &config.Config{}
	f.SetConfig(cfg1)

	// Override with second config
	cfg2 := &config.Config{}
	f.SetConfig(cfg2)

	// Should return the new config
	gotCfg, err := f.Config()
	if err != nil {
		t.Fatalf("Config() error: %v", err)
	}
	if gotCfg != cfg2 {
		t.Error("SetConfig should override previous config")
	}
}

func TestFactory_SetShellyService_OverridesOriginal(t *testing.T) {
	t.Parallel()

	f := cmdutil.New()

	// Set first service
	svc1 := shelly.NewService()
	f.SetShellyService(svc1)

	// Override with second service
	svc2 := shelly.NewService()
	f.SetShellyService(svc2)

	// Should return the new service
	gotSvc := f.ShellyService()
	if gotSvc != svc2 {
		t.Error("SetShellyService should override previous service")
	}
}
