// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tj-smith47/shelly-go/gen2/components"

	"github.com/tj-smith47/shelly-cli/internal/client"
)

// ScheduleJob contains schedule information.
type ScheduleJob struct {
	ID       int
	Enable   bool
	Timespec string
	Calls    []ScheduleCall
}

// ScheduleCall represents an RPC call in a schedule.
type ScheduleCall struct {
	Method string
	Params map[string]any
}

// ListSchedules lists all schedules on a device.
func (s *Service) ListSchedules(ctx context.Context, identifier string) ([]ScheduleJob, error) {
	var result []ScheduleJob
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		schedule := components.NewSchedule(conn.RPCClient())
		resp, err := schedule.List(ctx)
		if err != nil {
			return err
		}

		result = make([]ScheduleJob, len(resp.Jobs))
		for i, job := range resp.Jobs {
			calls := make([]ScheduleCall, len(job.Calls))
			for j, call := range job.Calls {
				params := make(map[string]any)
				if call.Params != nil {
					// Try to convert params to map
					if paramsMap, ok := call.Params.(map[string]any); ok {
						params = paramsMap
					}
				}
				calls[j] = ScheduleCall{
					Method: call.Method,
					Params: params,
				}
			}
			result[i] = ScheduleJob{
				ID:       job.ID,
				Enable:   job.Enable,
				Timespec: job.Timespec,
				Calls:    calls,
			}
		}
		return nil
	})
	return result, err
}

// CreateSchedule creates a new schedule on a device.
func (s *Service) CreateSchedule(ctx context.Context, identifier string, enable bool, timespec string, calls []ScheduleCall) (int, error) {
	var result int
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		schedule := components.NewSchedule(conn.RPCClient())

		// Convert calls to component format
		componentCalls := make([]components.ScheduleCall, len(calls))
		for i, call := range calls {
			componentCalls[i] = components.ScheduleCall{
				Method: call.Method,
				Params: call.Params,
			}
		}

		req := &components.ScheduleCreateRequest{
			Enable:   enable,
			Timespec: timespec,
			Calls:    componentCalls,
		}

		resp, err := schedule.Create(ctx, req)
		if err != nil {
			return err
		}
		result = resp.ID
		return nil
	})
	return result, err
}

// UpdateSchedule updates an existing schedule on a device.
func (s *Service) UpdateSchedule(ctx context.Context, identifier string, id int, enable *bool, timespec *string, calls []ScheduleCall) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		schedule := components.NewSchedule(conn.RPCClient())

		req := &components.ScheduleUpdateRequest{
			ID:       id,
			Enable:   enable,
			Timespec: timespec,
		}

		if calls != nil {
			componentCalls := make([]components.ScheduleCall, len(calls))
			for i, call := range calls {
				componentCalls[i] = components.ScheduleCall{
					Method: call.Method,
					Params: call.Params,
				}
			}
			req.Calls = componentCalls
		}

		_, err := schedule.Update(ctx, req)
		return err
	})
}

// DeleteSchedule deletes a schedule from a device.
func (s *Service) DeleteSchedule(ctx context.Context, identifier string, id int) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		schedule := components.NewSchedule(conn.RPCClient())
		_, err := schedule.Delete(ctx, id)
		return err
	})
}

// DeleteAllSchedules deletes all schedules from a device.
func (s *Service) DeleteAllSchedules(ctx context.Context, identifier string) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		schedule := components.NewSchedule(conn.RPCClient())
		_, err := schedule.DeleteAll(ctx)
		return err
	})
}

// EnableSchedule enables a schedule on a device.
func (s *Service) EnableSchedule(ctx context.Context, identifier string, id int) error {
	enable := true
	return s.UpdateSchedule(ctx, identifier, id, &enable, nil, nil)
}

// DisableSchedule disables a schedule on a device.
func (s *Service) DisableSchedule(ctx context.Context, identifier string, id int) error {
	enable := false
	return s.UpdateSchedule(ctx, identifier, id, &enable, nil, nil)
}

// ParseScheduleCalls parses a JSON string into schedule calls.
func ParseScheduleCalls(callsJSON string) ([]ScheduleCall, error) {
	var raw []struct {
		Method string         `json:"method"`
		Params map[string]any `json:"params"`
	}
	if err := json.Unmarshal([]byte(callsJSON), &raw); err != nil {
		return nil, fmt.Errorf("invalid calls JSON: %w", err)
	}

	calls := make([]ScheduleCall, len(raw))
	for i, r := range raw {
		calls[i] = ScheduleCall{
			Method: r.Method,
			Params: r.Params,
		}
	}
	return calls, nil
}
