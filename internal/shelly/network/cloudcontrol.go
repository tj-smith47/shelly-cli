// Package network provides cloud control operations.
package network

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// CloudControlResult holds the result of a cloud control action.
type CloudControlResult struct {
	Success bool
	Message string
}

// ExecuteAction executes a cloud control action on a device.
func (c *CloudClient) ExecuteAction(ctx context.Context, deviceID, action string, channel int) (CloudControlResult, error) {
	actionLower := strings.ToLower(action)

	// Try simple action
	if result, err := c.executeSimpleAction(ctx, deviceID, actionLower, channel); err == nil {
		return result, nil
	} else if !errors.Is(err, errNotSimpleAction) {
		return CloudControlResult{}, err
	}

	// Try parameterized action
	return c.executeParameterizedAction(ctx, deviceID, actionLower, channel)
}

var errNotSimpleAction = fmt.Errorf("not a simple action")

type actionHandler struct {
	fn      func(ctx context.Context, c *CloudClient, deviceID string, channel int) error
	success string
	errMsg  string
}

func buildCloudActionHandlers() map[string]actionHandler {
	return map[string]actionHandler{
		"on": {
			fn: func(ctx context.Context, c *CloudClient, d string, ch int) error {
				return c.SetSwitch(ctx, d, ch, true)
			},
			success: "Switch turned on",
			errMsg:  "failed to turn on switch",
		},
		"off": {
			fn: func(ctx context.Context, c *CloudClient, d string, ch int) error {
				return c.SetSwitch(ctx, d, ch, false)
			},
			success: "Switch turned off",
			errMsg:  "failed to turn off switch",
		},
		"toggle": {
			fn: func(ctx context.Context, c *CloudClient, d string, ch int) error {
				return c.ToggleSwitch(ctx, d, ch)
			},
			success: "Switch toggled",
			errMsg:  "failed to toggle switch",
		},
		"open": {
			fn: func(ctx context.Context, c *CloudClient, d string, ch int) error {
				return c.OpenCover(ctx, d, ch)
			},
			success: "Cover opening",
			errMsg:  "failed to open cover",
		},
		"close": {
			fn: func(ctx context.Context, c *CloudClient, d string, ch int) error {
				return c.CloseCover(ctx, d, ch)
			},
			success: "Cover closing",
			errMsg:  "failed to close cover",
		},
		"stop": {
			fn: func(ctx context.Context, c *CloudClient, d string, ch int) error {
				return c.StopCover(ctx, d, ch)
			},
			success: "Cover stopped",
			errMsg:  "failed to stop cover",
		},
		"light-on": {
			fn: func(ctx context.Context, c *CloudClient, d string, ch int) error {
				return c.SetLight(ctx, d, ch, true)
			},
			success: "Light turned on",
			errMsg:  "failed to turn on light",
		},
		"light-off": {
			fn: func(ctx context.Context, c *CloudClient, d string, ch int) error {
				return c.SetLight(ctx, d, ch, false)
			},
			success: "Light turned off",
			errMsg:  "failed to turn off light",
		},
		"light-toggle": {
			fn: func(ctx context.Context, c *CloudClient, d string, ch int) error {
				return c.ToggleLight(ctx, d, ch)
			},
			success: "Light toggled",
			errMsg:  "failed to toggle light",
		},
	}
}

func (c *CloudClient) executeSimpleAction(ctx context.Context, deviceID, action string, channel int) (CloudControlResult, error) {
	handlers := buildCloudActionHandlers()

	handler, ok := handlers[action]
	if !ok {
		return CloudControlResult{}, errNotSimpleAction
	}

	if err := handler.fn(ctx, c, deviceID, channel); err != nil {
		return CloudControlResult{}, fmt.Errorf("%s: %w", handler.errMsg, err)
	}

	return CloudControlResult{Success: true, Message: handler.success}, nil
}

func (c *CloudClient) executeParameterizedAction(ctx context.Context, deviceID, action string, channel int) (CloudControlResult, error) {
	switch {
	case strings.HasPrefix(action, "position="):
		return c.handlePositionAction(ctx, deviceID, action, channel)

	case strings.HasPrefix(action, "brightness="):
		return c.handleBrightnessAction(ctx, deviceID, action, channel)

	default:
		return CloudControlResult{}, fmt.Errorf("unknown action: %s", action)
	}
}

func (c *CloudClient) handlePositionAction(ctx context.Context, deviceID, action string, channel int) (CloudControlResult, error) {
	var position int
	if _, err := fmt.Sscanf(action, "position=%d", &position); err != nil {
		return CloudControlResult{}, fmt.Errorf("invalid position format: %s", action)
	}
	if position < 0 || position > 100 {
		return CloudControlResult{}, fmt.Errorf("position must be 0-100, got %d", position)
	}
	if err := c.SetCoverPosition(ctx, deviceID, channel, position); err != nil {
		return CloudControlResult{}, fmt.Errorf("failed to set cover position: %w", err)
	}
	return CloudControlResult{Success: true, Message: fmt.Sprintf("Cover position set to %d%%", position)}, nil
}

func (c *CloudClient) handleBrightnessAction(ctx context.Context, deviceID, action string, channel int) (CloudControlResult, error) {
	var brightness int
	if _, err := fmt.Sscanf(action, "brightness=%d", &brightness); err != nil {
		return CloudControlResult{}, fmt.Errorf("invalid brightness format: %s", action)
	}
	if brightness < 0 || brightness > 100 {
		return CloudControlResult{}, fmt.Errorf("brightness must be 0-100, got %d", brightness)
	}
	if err := c.SetLightBrightness(ctx, deviceID, channel, brightness); err != nil {
		return CloudControlResult{}, fmt.Errorf("failed to set light brightness: %w", err)
	}
	return CloudControlResult{Success: true, Message: fmt.Sprintf("Light brightness set to %d%%", brightness)}, nil
}
