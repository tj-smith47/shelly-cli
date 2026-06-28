package cmdutil

import (
	"context"
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Group control actions, re-exported so command files can select a quick action
// without importing the service package directly.
const (
	GroupActionOn     = shelly.ActionOn
	GroupActionOff    = shelly.ActionOff
	GroupActionToggle = shelly.ActionToggle
)

// ResolveGroupDevices resolves a group name to its member device identifiers,
// mirroring the resolution used by `group members`. It returns an error if the
// group does not exist or has no members.
func ResolveGroupDevices(f *Factory, groupName string) ([]string, error) {
	group := f.GetGroup(groupName)
	if group == nil {
		return nil, fmt.Errorf("group %q not found", groupName)
	}
	if len(group.Devices) == 0 {
		return nil, fmt.Errorf("group %q has no devices", groupName)
	}

	targets := make([]string, len(group.Devices))
	copy(targets, group.Devices)
	return targets, nil
}

// RunGroupLightSet resolves a group's members and applies the given light
// parameters to every member concurrently, printing a per-member result
// summary. Nil pointers leave the corresponding parameter unchanged.
func RunGroupLightSet(
	ctx context.Context,
	f *Factory,
	groupName string,
	concurrent, lightID int,
	brightness, temp *int,
	on *bool,
) error {
	targets, err := ResolveGroupDevices(f, groupName)
	if err != nil {
		return err
	}

	ios := f.IOStreams()
	svc := f.ShellyService()
	ios.Info("Setting light parameters on %d device(s) in group %q", len(targets), groupName)

	return RunBatch(ctx, ios, svc, targets, concurrent, func(ctx context.Context, svc *shelly.Service, device string) error {
		return svc.LightSet(ctx, device, lightID, brightness, temp, on)
	})
}

// RunGroupQuick resolves a group's members and applies a quick on/off/toggle
// action to every member concurrently, printing a per-member result summary.
// The action must be one of GroupActionOn, GroupActionOff, or GroupActionToggle.
// A nil componentID targets all controllable components on each member.
func RunGroupQuick(
	ctx context.Context,
	f *Factory,
	groupName, action string,
	componentID *int,
	concurrent int,
) error {
	targets, err := ResolveGroupDevices(f, groupName)
	if err != nil {
		return err
	}

	ios := f.IOStreams()
	svc := f.ShellyService()
	ios.Info("Sending %q to %d device(s) in group %q", action, len(targets), groupName)

	return RunBatch(ctx, ios, svc, targets, concurrent, func(ctx context.Context, svc *shelly.Service, device string) error {
		switch action {
		case GroupActionOn:
			_, err := svc.QuickOn(ctx, device, componentID)
			return err
		case GroupActionOff:
			_, err := svc.QuickOff(ctx, device, componentID)
			return err
		case GroupActionToggle:
			_, err := svc.QuickToggle(ctx, device, componentID)
			return err
		default:
			return fmt.Errorf("unknown group action %q", action)
		}
	})
}
