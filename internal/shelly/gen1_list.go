package shelly

import (
	"context"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// Gen1 device modes as reported by /settings.
const (
	gen1ModeRoller = "roller"
	gen1ModeRelay  = "relay"
	gen1ModeColor  = "color"
)

// gen1ComponentNames returns the configured name of every component of one
// kind on a Gen1 device, indexed by component ID. Gen1 has no call that
// enumerates components, so the list is read from /settings. The device mode
// decides which kinds exist: a Shelly 2.5 in roller mode has no usable relays,
// and a bulb in color mode has a color channel in place of its white light.
func gen1ComponentNames(ctx context.Context, conn *client.Gen1Client, kind model.ComponentType) ([]string, error) {
	settings, err := conn.GetSettings(ctx)
	if err != nil {
		return nil, err
	}

	names := []string{}
	switch kind {
	case model.ComponentSwitch:
		if settings.Mode != gen1ModeRoller {
			for i := range settings.Relays {
				names = append(names, settings.Relays[i].Name)
			}
		}
	case model.ComponentCover:
		if settings.Mode != gen1ModeRelay {
			for range settings.Rollers {
				names = append(names, "")
			}
		}
	case model.ComponentLight:
		if settings.Mode != gen1ModeColor {
			for i := range settings.Lights {
				names = append(names, settings.Lights[i].Name)
			}
		}
	case model.ComponentRGB, model.ComponentRGBW:
		if settings.Mode == gen1ModeColor {
			for i := range settings.Lights {
				names = append(names, settings.Lights[i].Name)
			}
		}
	default:
	}
	return names, nil
}

// gen1InputList lists the inputs a Gen1 device reports in /status. Gen1 inputs
// have no name; the type is the button type of the relay, roller or light the
// input drives, when the device has one at the same index.
func gen1InputList(ctx context.Context, conn *client.Gen1Client) ([]InputInfo, error) {
	status, err := conn.GetStatus(ctx)
	if err != nil {
		return nil, err
	}
	settings, err := conn.GetSettings(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]InputInfo, 0, len(status.Inputs))
	for i, in := range status.Inputs {
		info := InputInfo{ID: i, State: in.Input != 0}
		switch {
		case i < len(settings.Relays):
			info.Type = settings.Relays[i].BtnType
		case i < len(settings.Rollers):
			info.Type = settings.Rollers[i].BtnType
		case i < len(settings.Lights):
			info.Type = settings.Lights[i].BtnType
		}
		result = append(result, info)
	}
	return result, nil
}
