// Package client provides device communication abstraction over shelly-go SDK.
package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/tj-smith47/shelly-go/gen1"
	gen1comp "github.com/tj-smith47/shelly-go/gen1/components"
	"github.com/tj-smith47/shelly-go/transport"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// Gen1Client represents a connection to a Gen1 Shelly device.
type Gen1Client struct {
	device    *gen1.Device
	transport transport.Transport
	info      *DeviceInfo
}

// ConnectGen1 establishes a connection to a Gen1 Shelly device.
func ConnectGen1(ctx context.Context, device model.Device) (*Gen1Client, error) {
	url := device.Address
	if url != "" && url[0] != 'h' {
		url = "http://" + url
	}

	var opts []transport.Option
	if device.HasAuth() {
		opts = append(opts, transport.WithAuth(device.Auth.Username, device.Auth.Password))
	}
	if strings.HasPrefix(url, "https") {
		opts = append(opts, transport.WithInsecureSkipVerify())
	}

	httpTransport := transport.NewHTTP(url, opts...)
	gen1Device := gen1.NewDevice(httpTransport)

	info, err := gen1Device.GetDeviceInfo(ctx)
	if err != nil {
		iostreams.CloseWithDebug("closing gen1 device after connection failure", gen1Device)
		return nil, fmt.Errorf("%w: %w", model.ErrConnectionFailed, err)
	}

	return &Gen1Client{
		device:    gen1Device,
		transport: httpTransport,
		info: &DeviceInfo{
			ID:         info.ID,
			MAC:        info.MAC,
			Model:      info.Model,
			Generation: int(info.Generation),
			Firmware:   info.Version,
			App:        info.App,
			AuthEn:     info.AuthEnabled,
		},
	}, nil
}

// Close closes the device connection.
func (c *Gen1Client) Close() error {
	if c.device != nil {
		return c.device.Close()
	}
	return nil
}

// Info returns the device information.
func (c *Gen1Client) Info() *DeviceInfo {
	return c.info
}

// Device returns the underlying gen1.Device for advanced operations.
func (c *Gen1Client) Device() *gen1.Device {
	return c.device
}

// Call makes a raw REST API call to the device.
func (c *Gen1Client) Call(ctx context.Context, path string) ([]byte, error) {
	return c.device.Call(ctx, path)
}

// GetStatus returns the full device status.
func (c *Gen1Client) GetStatus(ctx context.Context) (*gen1.Status, error) {
	return c.device.GetFullStatus(ctx)
}

// GetSettings returns the full device settings.
func (c *Gen1Client) GetSettings(ctx context.Context) (*gen1.Settings, error) {
	return c.device.GetSettings(ctx)
}

// GetDebugLog returns the device debug log.
func (c *Gen1Client) GetDebugLog(ctx context.Context) (string, error) {
	return c.device.GetDebugLog(ctx)
}

// GetActions returns all configured action URLs.
func (c *Gen1Client) GetActions(ctx context.Context) (*gen1.ActionSettings, error) {
	return c.device.GetActions(ctx)
}

// SetAction configures an action URL for a specific event.
func (c *Gen1Client) SetAction(ctx context.Context, index int, event gen1.ActionEvent, urls []string, enabled bool) error {
	return c.device.SetAction(ctx, index, event, urls, enabled)
}

// SetActionURL is a convenience method to set a single action URL.
func (c *Gen1Client) SetActionURL(ctx context.Context, index int, event gen1.ActionEvent, url string, enabled bool) error {
	return c.device.SetActionURL(ctx, index, event, url, enabled)
}

// ClearAction disables and clears an action.
func (c *Gen1Client) ClearAction(ctx context.Context, index int, event gen1.ActionEvent) error {
	return c.device.ClearAction(ctx, index, event)
}

// Reboot reboots the device.
func (c *Gen1Client) Reboot(ctx context.Context) error {
	return c.device.Reboot(ctx)
}

// FactoryReset performs a factory reset on the device.
func (c *Gen1Client) FactoryReset(ctx context.Context) error {
	return c.device.FactoryReset(ctx)
}

// CheckForUpdate checks if a firmware update is available.
func (c *Gen1Client) CheckForUpdate(ctx context.Context) (*gen1.UpdateInfo, error) {
	return c.device.CheckForUpdate(ctx)
}

// Update starts a firmware update.
func (c *Gen1Client) Update(ctx context.Context, url string) error {
	return c.device.Update(ctx, url)
}

// ErrInvalidComponentID is returned when a component ID is negative.
var ErrInvalidComponentID = fmt.Errorf("invalid component ID: must be >= 0")

// Relay returns a relay component accessor.
// Returns an error if the ID is negative.
func (c *Gen1Client) Relay(id int) (*Gen1RelayComponent, error) {
	if id < 0 {
		return nil, fmt.Errorf("%w: relay ID %d", ErrInvalidComponentID, id)
	}
	return &Gen1RelayComponent{
		relay: c.device.Relay(id),
		id:    id,
	}, nil
}

// Roller returns a roller component accessor.
// Returns an error if the ID is negative.
func (c *Gen1Client) Roller(id int) (*Gen1RollerComponent, error) {
	if id < 0 {
		return nil, fmt.Errorf("%w: roller ID %d", ErrInvalidComponentID, id)
	}
	return &Gen1RollerComponent{
		roller: c.device.Roller(id),
		id:     id,
	}, nil
}

// Light returns a light component accessor.
// Returns an error if the ID is negative.
func (c *Gen1Client) Light(id int) (*Gen1LightComponent, error) {
	if id < 0 {
		return nil, fmt.Errorf("%w: light ID %d", ErrInvalidComponentID, id)
	}
	return &Gen1LightComponent{
		light: c.device.Light(id),
		id:    id,
	}, nil
}

// Color returns a color (RGBW) component accessor.
// Returns an error if the ID is negative.
func (c *Gen1Client) Color(id int) (*Gen1ColorComponent, error) {
	if id < 0 {
		return nil, fmt.Errorf("%w: color ID %d", ErrInvalidComponentID, id)
	}
	return &Gen1ColorComponent{
		color: c.device.Color(id),
		id:    id,
	}, nil
}

// White returns a white channel component accessor.
// Returns an error if the ID is negative.
func (c *Gen1Client) White(id int) (*Gen1WhiteComponent, error) {
	if id < 0 {
		return nil, fmt.Errorf("%w: white ID %d", ErrInvalidComponentID, id)
	}
	return &Gen1WhiteComponent{
		white: c.device.White(id),
		id:    id,
	}, nil
}

// Gen1RelayComponent wraps a Gen1 relay for switch-like operations.
type Gen1RelayComponent struct {
	relay *gen1comp.Relay
	id    int
}

// ID returns the relay ID.
func (r *Gen1RelayComponent) ID() int {
	return r.id
}

// TurnOn turns the relay on.
func (r *Gen1RelayComponent) TurnOn(ctx context.Context) error {
	return r.relay.TurnOn(ctx)
}

// TurnOff turns the relay off.
func (r *Gen1RelayComponent) TurnOff(ctx context.Context) error {
	return r.relay.TurnOff(ctx)
}

// Toggle toggles the relay state.
func (r *Gen1RelayComponent) Toggle(ctx context.Context) error {
	return r.relay.Toggle(ctx)
}

// Set sets the relay to a specific state.
func (r *Gen1RelayComponent) Set(ctx context.Context, on bool) error {
	return r.relay.Set(ctx, on)
}

// GetStatus returns the relay status.
func (r *Gen1RelayComponent) GetStatus(ctx context.Context) (*gen1comp.RelayStatus, error) {
	return r.relay.GetStatus(ctx)
}

// GetConfig returns the relay configuration.
func (r *Gen1RelayComponent) GetConfig(ctx context.Context) (*gen1comp.RelayConfig, error) {
	return r.relay.GetConfig(ctx)
}

// TurnOnForDuration turns the relay on for a specified duration.
func (r *Gen1RelayComponent) TurnOnForDuration(ctx context.Context, duration int) error {
	return r.relay.TurnOnForDuration(ctx, duration)
}

// TurnOffForDuration turns the relay off for a specified duration.
func (r *Gen1RelayComponent) TurnOffForDuration(ctx context.Context, duration int) error {
	return r.relay.TurnOffForDuration(ctx, duration)
}

// Gen1RollerComponent wraps a Gen1 roller for cover-like operations.
type Gen1RollerComponent struct {
	roller *gen1comp.Roller
	id     int
}

// ID returns the roller ID.
func (r *Gen1RollerComponent) ID() int {
	return r.id
}

// Open opens the roller (cover).
func (r *Gen1RollerComponent) Open(ctx context.Context) error {
	return r.roller.Open(ctx)
}

// Close closes the roller (cover).
func (r *Gen1RollerComponent) Close(ctx context.Context) error {
	return r.roller.Close(ctx)
}

// Stop stops the roller.
func (r *Gen1RollerComponent) Stop(ctx context.Context) error {
	return r.roller.Stop(ctx)
}

// GoToPosition moves the roller to a specific position.
func (r *Gen1RollerComponent) GoToPosition(ctx context.Context, position int) error {
	return r.roller.GoToPosition(ctx, position)
}

// GetStatus returns the roller status.
func (r *Gen1RollerComponent) GetStatus(ctx context.Context) (*gen1comp.RollerStatus, error) {
	return r.roller.GetStatus(ctx)
}

// OpenForDuration opens the roller for a specified duration.
func (r *Gen1RollerComponent) OpenForDuration(ctx context.Context, seconds float64) error {
	return r.roller.OpenForDuration(ctx, seconds)
}

// CloseForDuration closes the roller for a specified duration.
func (r *Gen1RollerComponent) CloseForDuration(ctx context.Context, seconds float64) error {
	return r.roller.CloseForDuration(ctx, seconds)
}

// Calibrate starts the roller calibration procedure.
func (r *Gen1RollerComponent) Calibrate(ctx context.Context) error {
	return r.roller.Calibrate(ctx)
}

// Gen1LightComponent wraps a Gen1 light.
type Gen1LightComponent struct {
	light *gen1comp.Light
	id    int
}

// ID returns the light ID.
func (l *Gen1LightComponent) ID() int {
	return l.id
}

// TurnOn turns the light on.
func (l *Gen1LightComponent) TurnOn(ctx context.Context) error {
	return l.light.TurnOn(ctx)
}

// TurnOff turns the light off.
func (l *Gen1LightComponent) TurnOff(ctx context.Context) error {
	return l.light.TurnOff(ctx)
}

// Toggle toggles the light state.
func (l *Gen1LightComponent) Toggle(ctx context.Context) error {
	return l.light.Toggle(ctx)
}

// SetBrightness sets the light brightness (0-100).
func (l *Gen1LightComponent) SetBrightness(ctx context.Context, brightness int) error {
	return l.light.SetBrightness(ctx, brightness)
}

// GetStatus returns the light status.
func (l *Gen1LightComponent) GetStatus(ctx context.Context) (*gen1comp.LightStatus, error) {
	return l.light.GetStatus(ctx)
}

// TurnOnWithBrightness turns on with a specific brightness.
func (l *Gen1LightComponent) TurnOnWithBrightness(ctx context.Context, brightness int) error {
	return l.light.TurnOnWithBrightness(ctx, brightness)
}

// TurnOnForDuration turns the light on for a specified duration.
func (l *Gen1LightComponent) TurnOnForDuration(ctx context.Context, duration int) error {
	return l.light.TurnOnForDuration(ctx, duration)
}

// SetBrightnessWithTransition sets brightness with a transition time.
func (l *Gen1LightComponent) SetBrightnessWithTransition(ctx context.Context, brightness, transitionMs int) error {
	return l.light.SetBrightnessWithTransition(ctx, brightness, transitionMs)
}

// Gen1ColorComponent wraps a Gen1 color (RGBW) light.
type Gen1ColorComponent struct {
	color *gen1comp.Color
	id    int
}

// ID returns the color ID.
func (c *Gen1ColorComponent) ID() int {
	return c.id
}

// TurnOn turns the color light on.
func (c *Gen1ColorComponent) TurnOn(ctx context.Context) error {
	return c.color.TurnOn(ctx)
}

// TurnOff turns the color light off.
func (c *Gen1ColorComponent) TurnOff(ctx context.Context) error {
	return c.color.TurnOff(ctx)
}

// Toggle toggles the color light state.
func (c *Gen1ColorComponent) Toggle(ctx context.Context) error {
	return c.color.Toggle(ctx)
}

// SetRGB sets the RGB color.
func (c *Gen1ColorComponent) SetRGB(ctx context.Context, red, green, blue int) error {
	return c.color.SetRGB(ctx, red, green, blue)
}

// SetRGBW sets the RGBW color.
func (c *Gen1ColorComponent) SetRGBW(ctx context.Context, red, green, blue, white int) error {
	return c.color.SetRGBW(ctx, red, green, blue, white)
}

// SetGain sets the gain (brightness) for the color light.
func (c *Gen1ColorComponent) SetGain(ctx context.Context, gain int) error {
	return c.color.SetGain(ctx, gain)
}

// GetStatus returns the color light status.
func (c *Gen1ColorComponent) GetStatus(ctx context.Context) (*gen1comp.ColorStatus, error) {
	return c.color.GetStatus(ctx)
}

// TurnOnForDuration turns the color light on for a specified duration.
func (c *Gen1ColorComponent) TurnOnForDuration(ctx context.Context, duration int) error {
	return c.color.TurnOnForDuration(ctx, duration)
}

// TurnOnWithRGB turns on with specific RGB values and gain.
func (c *Gen1ColorComponent) TurnOnWithRGB(ctx context.Context, red, green, blue, gain int) error {
	return c.color.TurnOnWithRGB(ctx, red, green, blue, gain)
}

// Gen1WhiteComponent wraps a Gen1 white channel.
type Gen1WhiteComponent struct {
	white *gen1comp.White
	id    int
}

// ID returns the white channel ID.
func (w *Gen1WhiteComponent) ID() int {
	return w.id
}

// TurnOn turns the white channel on.
func (w *Gen1WhiteComponent) TurnOn(ctx context.Context) error {
	return w.white.TurnOn(ctx)
}

// TurnOff turns the white channel off.
func (w *Gen1WhiteComponent) TurnOff(ctx context.Context) error {
	return w.white.TurnOff(ctx)
}

// SetBrightness sets the white channel brightness.
func (w *Gen1WhiteComponent) SetBrightness(ctx context.Context, brightness int) error {
	return w.white.SetBrightness(ctx, brightness)
}

// GetStatus returns the white channel status.
func (w *Gen1WhiteComponent) GetStatus(ctx context.Context) (*gen1comp.WhiteStatus, error) {
	return w.white.GetStatus(ctx)
}
