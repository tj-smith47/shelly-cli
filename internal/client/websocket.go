package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	neturl "net/url"

	"github.com/tj-smith47/shelly-go/transport"
	"github.com/tj-smith47/shelly-go/types"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/netguard"
)

// gen2User is the only user a Gen2+ device has.
const gen2User = "admin"

// DeviceWebSocket is a websocket to a Gen2+ device together with the user
// whose credentials it answers a digest challenge with.
type DeviceWebSocket struct {
	*transport.WebSocket

	// user is empty when the websocket was opened without a password.
	user string
}

// NewDeviceWebSocket opens transport.NewWebSocket(url, opts...). When auth
// carries a password the websocket answers a device's digest challenge with
// it; a device with authentication disabled is never sent credentials.
//
// The SDK websocket takes no dialer, so under `go test` a url whose host is
// not a loopback address is refused here with netguard.ErrBlocked before
// anything dials.
func NewDeviceWebSocket(url string, auth *model.Auth, opts ...transport.Option) (*DeviceWebSocket, error) {
	u, err := neturl.Parse(url)
	if err != nil {
		return nil, fmt.Errorf("device websocket url %q: %w", url, err)
	}
	if err := netguard.RefuseAddr(u.Host); err != nil {
		return nil, err
	}
	ws := &DeviceWebSocket{}
	if auth != nil && auth.Password != "" {
		ws.user = auth.Username
		if ws.user == "" {
			ws.user = gen2User
		}
		opts = append(opts, transport.WithDigestAuth(ws.user, auth.Password))
	}
	ws.WebSocket = transport.NewWebSocket(url, opts...)
	return ws, nil
}

// wsRequest is a request frame without an id or auth object; the transport
// adds both.
type wsRequest struct {
	method string
	params json.RawMessage
}

func (r *wsRequest) GetID() any                 { return nil }
func (r *wsRequest) GetMethod() string          { return r.method }
func (r *wsRequest) GetParams() json.RawMessage { return r.params }
func (r *wsRequest) GetJSONRPC() string         { return "2.0" }
func (r *wsRequest) GetAuth() any               { return nil }
func (r *wsRequest) IsREST() bool               { return false }

// CallDeviceWebSocket makes one RPC call over a device websocket.
//
// A device that refuses the websocket's credentials returns a
// *model.CredentialsRejectedError (errors.Is model.ErrCredentialsRejected).
// A device that asks for credentials when the websocket has none returns an
// error that errors.Is model.ErrAuthRequired.
func CallDeviceWebSocket(ctx context.Context, ws *DeviceWebSocket, method string, params any) (json.RawMessage, error) {
	req := &wsRequest{method: method}
	if params != nil {
		data, err := json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("marshal %s params: %w", method, err)
		}
		req.params = data
	}

	result, err := ws.Call(ctx, req)
	if !errors.Is(err, types.ErrAuth) {
		return result, err
	}
	if ws.user == "" {
		return nil, fmt.Errorf("%w: no credentials were given", model.ErrAuthRequired)
	}
	return nil, &model.CredentialsRejectedError{User: ws.user}
}

// StartDeviceNotifications makes the request frame a Gen2+ device needs
// before it sends notifications over ws (Shelly.GetStatus, answered with the
// device's full status, which is returned). Call it after ws.Connect. A
// device forgets the subscription when the connection drops, so the frame is
// sent again after every reconnect until ctx ends; a failure there is only
// logged, since nothing is waiting for it.
func StartDeviceNotifications(ctx context.Context, ws *DeviceWebSocket) (json.RawMessage, error) {
	status, err := CallDeviceWebSocket(ctx, ws, "Shelly.GetStatus", nil)
	if err != nil {
		return nil, err
	}
	ws.OnStateChange(func(state transport.ConnectionState) {
		if state != transport.StateConnected || ctx.Err() != nil {
			return
		}
		// State listeners run on the goroutine that reconnected; a call
		// made there would hold up every following state change until answered.
		go func() {
			if _, err := CallDeviceWebSocket(ctx, ws, "Shelly.GetStatus", nil); err != nil && ctx.Err() == nil {
				iostreams.DebugErrCat(iostreams.CategoryNetwork, "resubscribe to device notifications", err)
			}
		}()
	})
	return status, nil
}
