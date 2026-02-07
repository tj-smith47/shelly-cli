// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// ErrNoLink is returned when a device has no link configured.
var ErrNoLink = fmt.Errorf("no link configured")

// ResolveLinkStatus resolves the status of a linked child device.
// Returns ErrNoLink if the device has no link configured.
// When the parent is reachable, returns the parent switch state for deriving child state.
func (s *Service) ResolveLinkStatus(ctx context.Context, childDevice string) (*model.LinkStatus, error) {
	link, ok := config.GetLink(childDevice)
	if !ok {
		return nil, ErrNoLink
	}

	ls := &model.LinkStatus{
		ChildDevice:  childDevice,
		ParentDevice: link.ParentDevice,
		SwitchID:     link.SwitchID,
	}

	switchStatus, err := s.SwitchStatus(ctx, link.ParentDevice, link.SwitchID)
	if err != nil {
		ls.ParentOnline = false
		ls.State = "Unknown"
		return ls, nil //nolint:nilerr // parent unreachable is not an error, we return Unknown state
	}

	ls.ParentOnline = true
	ls.SwitchOutput = switchStatus.Output
	if switchStatus.Output {
		ls.State = "On"
	} else {
		ls.State = fmt.Sprintf("Off (via %s:%d)", link.ParentDevice, link.SwitchID)
	}

	return ls, nil
}
