package monitoring

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/tj-smith47/shelly-go/gen1"
	"github.com/tj-smith47/shelly-go/gen2/components"
	"golang.org/x/sync/errgroup"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// Component types reported in model.PowerReading.Type. Gen2+ readings carry
// the device's own component name (see ReadPowerReadings), so these list only
// the types the CLI treats specially.
const (
	EnergyTypeEM  = "em"
	EnergyTypeEM1 = "em1"
	MeterTypePM   = "pm"
	MeterTypePM1  = "pm1"
	// MeterTypeGen1Meter is a Gen1 /status meters[] entry (relays, plugs, bulbs).
	MeterTypeGen1Meter = "meter"
	// MeterTypeGen1EMeter is a Gen1 /status emeters[] entry (Shelly EM, 3EM).
	MeterTypeGen1EMeter = "emeter"
)

// ErrNoPowerMeter is returned when a device answers but none of its
// components reports power.
var ErrNoPowerMeter = errors.New("no component on this device reports power")

// meterRank orders a device's readings: dedicated meters first, then the
// switches, covers and lights that meter their own load, by name.
func meterRank(typ string) int {
	switch typ {
	case EnergyTypeEM:
		return 0
	case EnergyTypeEM1:
		return 1
	case MeterTypePM:
		return 2
	case MeterTypePM1:
		return 3
	default:
		return 4
	}
}

// ReadPowerReadings returns every power reading the device reports, one per
// component that meters power, in a stable order (see meterRank).
//
// A Gen2+ device is read with one Shelly.GetStatus call: em and em1 components
// are energy monitors, and any other component whose status has "apower"
// (pm, pm1, switch, cover, light, rgb, rgbw, cct, ...) is a power meter. A
// Gen1 device is read from /status meters[] and emeters[].
//
// An unreachable device returns the connection error. A device that answers
// but meters nothing returns an empty list and no error.
func (s *Service) ReadPowerReadings(ctx context.Context, device string) ([]model.PowerReading, error) {
	var readings []model.PowerReading
	var err error
	if resolved, rerr := s.connector.ResolveWithGeneration(ctx, device); rerr == nil && resolved.Generation == 1 {
		readings, err = s.readGen1PowerReadings(ctx, device)
	} else {
		readings, err = s.readGen2PowerReadings(ctx, device)
	}
	if err != nil {
		return nil, err
	}
	slices.SortStableFunc(readings, func(a, b model.PowerReading) int {
		return cmp.Or(
			cmp.Compare(meterRank(a.Type), meterRank(b.Type)),
			cmp.Compare(a.Type, b.Type),
			cmp.Compare(a.ID, b.ID),
		)
	})
	for i := range readings {
		readings[i].Name = device
	}
	return readings, nil
}

// ReadPowerReading returns one reading of the device. typ selects the
// component type ("" or "auto" for any) and id its index (negative for the
// first matching reading). The error names the readings the device does have
// when none matches.
func (s *Service) ReadPowerReading(ctx context.Context, device, typ string, id int) (model.PowerReading, error) {
	readings, err := s.ReadPowerReadings(ctx, device)
	if err != nil {
		return model.PowerReading{}, fmt.Errorf("failed to read power on %s: %w", device, err)
	}
	if len(readings) == 0 {
		return model.PowerReading{}, fmt.Errorf("%s: %w", device, ErrNoPowerMeter)
	}
	anyType := typ == "" || typ == "auto"
	for _, r := range readings {
		if (anyType || r.Type == typ) && (id < 0 || r.ID == id) {
			return r, nil
		}
	}
	want := typ
	if anyType {
		want = "component"
	}
	if id >= 0 {
		want = fmt.Sprintf("%s %d", want, id)
	}
	return model.PowerReading{}, fmt.Errorf("%s has no %s that reports power; its power readings are: %s",
		device, want, strings.Join(ReadingKeys(readings), ", "))
}

// ReadingKeys returns the "type:id" key of each reading.
func ReadingKeys(readings []model.PowerReading) []string {
	keys := make([]string, len(readings))
	for i, r := range readings {
		keys[i] = fmt.Sprintf("%s:%d", r.Type, r.ID)
	}
	return keys
}

func (s *Service) readGen2PowerReadings(ctx context.Context, device string) ([]model.PowerReading, error) {
	var status map[string]json.RawMessage
	err := s.connector.WithConnection(ctx, device, func(conn *client.Client) error {
		return conn.RPCClient().CallResult(ctx, "Shelly.GetStatus", map[string]any{}, &status)
	})
	if err != nil {
		return nil, err
	}
	readings := make([]model.PowerReading, 0, len(status))
	for key, raw := range status {
		typ, idStr, ok := strings.Cut(key, ":")
		if !ok {
			continue
		}
		id, convErr := strconv.Atoi(idStr)
		if convErr != nil {
			continue
		}
		if r, ok := gen2PowerReading(typ, id, raw); ok {
			readings = append(readings, r)
		}
	}
	return readings, nil
}

// gen2PowerReading decodes one Shelly.GetStatus entry. Every Gen2+ component
// that meters its load reports the same keys as PM1 (apower, voltage,
// current, freq, aenergy, ret_aenergy), so one decoder serves them all.
func gen2PowerReading(typ string, id int, raw json.RawMessage) (model.PowerReading, bool) {
	switch typ {
	case EnergyTypeEM:
		var st components.EMStatus
		if json.Unmarshal(raw, &st) != nil {
			return model.PowerReading{}, false
		}
		em := emStatusFromComponent(&st)
		return model.PowerReading{Type: typ, ID: id, Power: em.TotalActivePower, EM: em}, true
	case EnergyTypeEM1:
		var st components.EM1Status
		if json.Unmarshal(raw, &st) != nil {
			return model.PowerReading{}, false
		}
		em1 := em1StatusFromComponent(&st)
		return model.PowerReading{Type: typ, ID: id, Power: em1.ActPower, EM1: em1}, true
	}
	var probe struct {
		APower *float64 `json:"apower"`
	}
	if json.Unmarshal(raw, &probe) != nil || probe.APower == nil {
		return model.PowerReading{}, false
	}
	var st components.PM1Status
	if json.Unmarshal(raw, &st) != nil {
		return model.PowerReading{}, false
	}
	meter := pm1StatusFromComponent(&st)
	meter.ID = id
	return model.PowerReading{Type: typ, ID: id, Power: meter.APower, Meter: meter}, true
}

func (s *Service) readGen1PowerReadings(ctx context.Context, device string) ([]model.PowerReading, error) {
	var status *gen1.Status
	err := s.connector.WithGen1Connection(ctx, device, func(conn *client.Gen1Client) error {
		var err error
		status, err = conn.GetStatus(ctx)
		return err
	})
	if err != nil {
		return nil, err
	}
	readings := make([]model.PowerReading, 0, len(status.Meters)+len(status.EMeters))
	for i, m := range convertGen1Meters(status.Meters) {
		readings = append(readings, model.PowerReading{Type: MeterTypeGen1Meter, ID: i, Power: m.APower, Meter: &m})
	}
	for i, m := range convertGen1EMeters(status.EMeters, 0) {
		readings = append(readings, model.PowerReading{Type: MeterTypeGen1EMeter, ID: i, Power: m.APower, Meter: &m})
	}
	return readings, nil
}

// CollectPowerReadings reads every power reading on each device,
// concurrently. Readings follow the order of devices. A device that cannot be
// reached, or on which nothing meters power, is reported in skipped; the other
// devices are unaffected.
func (s *Service) CollectPowerReadings(ctx context.Context, devices []string) (readings []model.PowerReading, skipped []model.EnergyStatusSkip) {
	perDevice := make([][]model.PowerReading, len(devices))
	perSkip := make([]*model.EnergyStatusSkip, len(devices))

	var g errgroup.Group
	g.SetLimit(config.GetGlobalMaxConcurrent())
	for i, device := range devices {
		g.Go(func() error {
			r, err := s.ReadPowerReadings(ctx, device)
			switch {
			case err != nil:
				perSkip[i] = &model.EnergyStatusSkip{Device: device, Reason: fmt.Sprintf("unreachable: %v", err)}
			case len(r) == 0:
				perSkip[i] = &model.EnergyStatusSkip{Device: device, Reason: ErrNoPowerMeter.Error()}
			default:
				perDevice[i] = r
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, nil
	}

	readings = make([]model.PowerReading, 0, len(devices))
	for i := range devices {
		readings = append(readings, perDevice[i]...)
		if perSkip[i] != nil {
			skipped = append(skipped, *perSkip[i])
		}
	}
	return readings, skipped
}

// counterResetNamespace maps a component type to the RPC namespace whose
// ResetCounters method clears that component's energy counters.
var counterResetNamespace = map[string]string{
	EnergyTypeEM: "EM",
	MeterTypePM:  "PM",
	MeterTypePM1: "PM1",
	"switch":     "Switch",
	"cover":      "Cover",
	"light":      "Light",
	"rgb":        "RGB",
	"rgbw":       "RGBW",
}

// ResetPowerCounters resets the energy counters of one component that meters
// power, chosen as ReadPowerReading chooses it, and returns the reading taken
// before the reset. counterTypes limits the reset to those counters; empty
// resets all. A component without a counter reset (EM1, Gen1 meters) is an
// error that names the types that have one.
func (s *Service) ResetPowerCounters(ctx context.Context, device, typ string, id int, counterTypes []string) (model.PowerReading, error) {
	r, err := s.ReadPowerReading(ctx, device, typ, id)
	if err != nil {
		return r, err
	}
	ns, ok := counterResetNamespace[r.Type]
	if !ok {
		return r, fmt.Errorf("%s:%d on %s has no counter reset; energy reset works on %s components",
			r.Type, r.ID, device, strings.Join(slices.Sorted(maps.Keys(counterResetNamespace)), ", "))
	}
	params := map[string]any{"id": r.ID}
	if len(counterTypes) > 0 {
		params["type"] = counterTypes
	}
	err = s.connector.WithConnection(ctx, device, func(conn *client.Client) error {
		_, err := conn.RPCClient().Call(ctx, ns+".ResetCounters", params)
		return err
	})
	return r, err
}
