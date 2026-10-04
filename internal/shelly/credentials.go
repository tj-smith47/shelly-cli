package shelly

import (
	"context"
	"errors"
	"fmt"

	"github.com/tj-smith47/shelly-go/types"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// ErrCredentialsRejected indicates a device refused the credentials it was
// given. It is model.ErrCredentialsRejected, defined there so that
// internal/client can return it too.
var ErrCredentialsRejected = model.ErrCredentialsRejected

// CredentialsRejectedError is model.CredentialsRejectedError.
type CredentialsRejectedError = model.CredentialsRejectedError

// VerifyCredentials checks that the device at dev.Address accepts dev.Auth, by
// making a read-only request the device answers only to an authenticated
// caller: Sys.GetStatus on Gen2+ and GET /settings on Gen1 (dev.Generation
// selects which; anything other than 1 is treated as Gen2+). The device info
// endpoints (Shelly.GetDeviceInfo, /shelly) answer without credentials, so a
// successful probe says nothing about them.
//
// It returns nil when the request is answered, which is also the case for a
// device with authentication disabled. A refusal is a
// *CredentialsRejectedError (errors.Is ErrCredentialsRejected) when dev.Auth
// is set and model.ErrAuthRequired when it is not. Any other error, such as a
// device that cannot be reached, is returned as it is.
func (s *Service) VerifyCredentials(ctx context.Context, dev model.Device) error {
	var err error
	if dev.Generation == 1 {
		err = s.connManager.WithGen1DeviceConnection(ctx, dev, func(conn *client.Gen1Client) error {
			_, callErr := conn.GetSettings(ctx)
			return callErr
		})
	} else {
		err = s.connManager.WithDeviceConnection(ctx, dev, func(conn *client.Client) error {
			_, callErr := conn.Call(ctx, "Sys.GetStatus", nil)
			return callErr
		})
	}
	return credentialsError(dev, err)
}

// credentialsError turns a device's "unauthorized" answer to a request made
// for dev into a *CredentialsRejectedError, or into model.ErrAuthRequired when
// dev carries no credentials. Any other error, and nil, is returned unchanged.
func credentialsError(dev model.Device, err error) error {
	if err == nil || !errors.Is(err, types.ErrAuth) {
		return err
	}
	if !dev.HasAuth() {
		return fmt.Errorf("%w: no credentials were given", model.ErrAuthRequired)
	}
	return &CredentialsRejectedError{User: dev.Auth.Username}
}
