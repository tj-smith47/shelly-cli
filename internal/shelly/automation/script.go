// Package automation provides script, schedule, and event automation for Shelly devices.
package automation

import (
	"context"

	"github.com/tj-smith47/shelly-go/gen2/components"

	"github.com/tj-smith47/shelly-cli/internal/client"
)

// ScriptInfo contains information about a script.
type ScriptInfo struct {
	ID      int
	Name    string
	Enable  bool
	Running bool
}

// ScriptStatus contains detailed script status.
type ScriptStatus struct {
	ID       int
	Running  bool
	MemUsage int
	MemPeak  int
	MemFree  int
	Errors   []string
}

// ScriptConfig contains script configuration.
type ScriptConfig struct {
	ID     int
	Name   string
	Enable bool
}

// ListScripts lists all scripts on a device.
func (s *Service) ListScripts(ctx context.Context, identifier string) ([]ScriptInfo, error) {
	var result []ScriptInfo
	err := s.parent.WithConnection(ctx, identifier, func(conn *client.Client) error {
		script := components.NewScript(conn.RPCClient())
		resp, err := script.List(ctx)
		if err != nil {
			return err
		}

		result = make([]ScriptInfo, len(resp.Scripts))
		for i, item := range resp.Scripts {
			name := ""
			if item.Name != nil {
				name = *item.Name
			}
			result[i] = ScriptInfo{
				ID:      item.ID,
				Name:    name,
				Enable:  item.Enable,
				Running: item.Running,
			}
		}
		return nil
	})
	return result, err
}

// GetScriptStatus gets the status of a specific script.
func (s *Service) GetScriptStatus(ctx context.Context, identifier string, id int) (*ScriptStatus, error) {
	var result *ScriptStatus
	err := s.parent.WithConnection(ctx, identifier, func(conn *client.Client) error {
		script := components.NewScript(conn.RPCClient())
		status, err := script.GetStatus(ctx, id)
		if err != nil {
			return err
		}

		result = &ScriptStatus{
			ID:      status.ID,
			Running: status.Running,
			Errors:  status.Errors,
		}
		if status.MemUsage != nil {
			result.MemUsage = *status.MemUsage
		}
		if status.MemPeak != nil {
			result.MemPeak = *status.MemPeak
		}
		if status.MemFree != nil {
			result.MemFree = *status.MemFree
		}
		return nil
	})
	return result, err
}

// GetScriptConfig gets the configuration of a specific script.
func (s *Service) GetScriptConfig(ctx context.Context, identifier string, id int) (*ScriptConfig, error) {
	var result *ScriptConfig
	err := s.parent.WithConnection(ctx, identifier, func(conn *client.Client) error {
		script := components.NewScript(conn.RPCClient())
		config, err := script.GetConfig(ctx, id)
		if err != nil {
			return err
		}

		result = &ScriptConfig{
			ID: config.ID,
		}
		if config.Name != nil {
			result.Name = *config.Name
		}
		if config.Enable != nil {
			result.Enable = *config.Enable
		}
		return nil
	})
	return result, err
}

// GetScriptCode retrieves the source code of a script.
func (s *Service) GetScriptCode(ctx context.Context, identifier string, id int) (string, error) {
	var result string
	err := s.parent.WithConnection(ctx, identifier, func(conn *client.Client) error {
		script := components.NewScript(conn.RPCClient())
		resp, err := script.GetCode(ctx, id)
		if err != nil {
			return err
		}
		result = resp.Data
		return nil
	})
	return result, err
}

// CreateScript creates a new script on a device.
func (s *Service) CreateScript(ctx context.Context, identifier, name string) (int, error) {
	var result int
	err := s.parent.WithConnection(ctx, identifier, func(conn *client.Client) error {
		script := components.NewScript(conn.RPCClient())
		var namePtr *string
		if name != "" {
			namePtr = &name
		}
		resp, err := script.Create(ctx, namePtr)
		if err != nil {
			return err
		}
		result = resp.ID
		return nil
	})
	return result, err
}

// UpdateScriptCode updates the code of an existing script.
func (s *Service) UpdateScriptCode(ctx context.Context, identifier string, id int, code string, appendCode bool) error {
	return s.parent.WithConnection(ctx, identifier, func(conn *client.Client) error {
		script := components.NewScript(conn.RPCClient())
		return script.PutCode(ctx, id, code, appendCode)
	})
}

// UpdateScriptConfig updates the configuration of a script.
func (s *Service) UpdateScriptConfig(ctx context.Context, identifier string, id int, name *string, enable *bool) error {
	return s.parent.WithConnection(ctx, identifier, func(conn *client.Client) error {
		script := components.NewScript(conn.RPCClient())
		config := &components.ScriptConfig{
			Name:   name,
			Enable: enable,
		}
		return script.SetConfig(ctx, id, config)
	})
}

// DeleteScript deletes a script from a device.
func (s *Service) DeleteScript(ctx context.Context, identifier string, id int) error {
	return s.parent.WithConnection(ctx, identifier, func(conn *client.Client) error {
		script := components.NewScript(conn.RPCClient())
		return script.Delete(ctx, id)
	})
}

// StartScript starts a script on a device.
func (s *Service) StartScript(ctx context.Context, identifier string, id int) error {
	return s.parent.WithConnection(ctx, identifier, func(conn *client.Client) error {
		script := components.NewScript(conn.RPCClient())
		return script.Start(ctx, id)
	})
}

// StopScript stops a running script on a device.
func (s *Service) StopScript(ctx context.Context, identifier string, id int) error {
	return s.parent.WithConnection(ctx, identifier, func(conn *client.Client) error {
		script := components.NewScript(conn.RPCClient())
		return script.Stop(ctx, id)
	})
}

// EvalScript evaluates a JavaScript expression in the context of a running script.
func (s *Service) EvalScript(ctx context.Context, identifier string, id int, code string) (any, error) {
	var result any
	err := s.parent.WithConnection(ctx, identifier, func(conn *client.Client) error {
		script := components.NewScript(conn.RPCClient())
		resp, err := script.Eval(ctx, id, code)
		if err != nil {
			return err
		}
		result = resp.Result
		return nil
	})
	return result, err
}

// InstallScriptResult contains the result of installing a script.
type InstallScriptResult struct {
	ID      int
	Name    string
	Enabled bool
}

// InstallScript creates a new script, uploads code, and optionally enables it.
func (s *Service) InstallScript(ctx context.Context, identifier, name, code string, enable bool) (*InstallScriptResult, error) {
	var result InstallScriptResult
	result.Name = name
	result.Enabled = enable

	err := s.parent.WithConnection(ctx, identifier, func(conn *client.Client) error {
		script := components.NewScript(conn.RPCClient())

		// Create script
		var namePtr *string
		if name != "" {
			namePtr = &name
		}
		createResp, err := script.Create(ctx, namePtr)
		if err != nil {
			return err
		}
		result.ID = createResp.ID

		// Upload code
		if err := script.PutCode(ctx, result.ID, code, false); err != nil {
			return err
		}

		// Enable if requested
		if enable {
			cfg := &components.ScriptConfig{Enable: &enable}
			if err := script.SetConfig(ctx, result.ID, cfg); err != nil {
				return err
			}
		}

		return nil
	})

	return &result, err
}
