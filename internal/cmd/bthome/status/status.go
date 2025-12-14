// Package status provides the bthome status command.
package status

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Options holds command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
	ID      int
	HasID   bool
	JSON    bool
}

// NewCommand creates the bthome status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "status <device> [id]",
		Aliases: []string{"st", "info"},
		Short:   "Show BTHome device status",
		Long: `Show detailed status of BTHome devices.

Without an ID, shows the BTHome component status including any active
discovery scan. With an ID, shows detailed status of a specific
BTHomeDevice including signal strength, battery, and known objects.`,
		Example: `  # Show BTHome component status
  shelly bthome status living-room

  # Show specific device status
  shelly bthome status living-room 200

  # Output as JSON
  shelly bthome status living-room 200 --json`,
		Args:              cobra.RangeArgs(1, 2),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			if len(args) > 1 {
				id, err := strconv.Atoi(args[1])
				if err != nil {
					return fmt.Errorf("invalid device ID: %w", err)
				}
				opts.ID = id
				opts.HasID = true
			}
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output as JSON")

	return cmd
}

// BTHomeStatus represents BTHome component status.
type BTHomeStatus struct {
	Discovery *DiscoveryStatus `json:"discovery,omitempty"`
	Errors    []string         `json:"errors,omitempty"`
}

// DiscoveryStatus represents active discovery scan status.
type DiscoveryStatus struct {
	StartedAt float64 `json:"started_at"`
	Duration  int     `json:"duration"`
}

// DeviceStatus represents a BTHome device status.
type DeviceStatus struct {
	ID           int           `json:"id"`
	Name         string        `json:"name,omitempty"`
	Addr         string        `json:"addr"`
	RSSI         *int          `json:"rssi,omitempty"`
	Battery      *int          `json:"battery,omitempty"`
	PacketID     *int          `json:"packet_id,omitempty"`
	LastUpdateTS float64       `json:"last_updated_ts"`
	KnownObjects []KnownObject `json:"known_objects,omitempty"`
	Errors       []string      `json:"errors,omitempty"`
}

// KnownObject represents a known BTHome object.
type KnownObject struct {
	ObjID     int     `json:"obj_id"`
	Idx       int     `json:"idx"`
	Component *string `json:"component,omitempty"`
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	if opts.HasID {
		return showDeviceStatus(ctx, svc, opts, ios)
	}

	return showComponentStatus(ctx, svc, opts, ios)
}

func showComponentStatus(ctx context.Context, svc *shelly.Service, opts *Options, ios *iostreams.IOStreams) error {
	var status BTHomeStatus

	err := svc.WithConnection(ctx, opts.Device, func(conn *client.Client) error {
		result, err := conn.Call(ctx, "BTHome.GetStatus", nil)
		if err != nil {
			return fmt.Errorf("failed to get BTHome status: %w", err)
		}

		jsonBytes, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}
		if err := json.Unmarshal(jsonBytes, &status); err != nil {
			return fmt.Errorf("failed to parse status: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	if opts.JSON {
		output, err := json.MarshalIndent(status, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		ios.Println(string(output))
		return nil
	}

	ios.Println(theme.Bold().Render("BTHome Status:"))
	ios.Println()

	if status.Discovery != nil {
		ios.Println("  " + theme.Highlight().Render("Discovery:"))
		startTime := time.Unix(int64(status.Discovery.StartedAt), 0)
		ios.Printf("    Started: %s\n", startTime.Format(time.RFC3339))
		ios.Printf("    Duration: %ds\n", status.Discovery.Duration)
		ios.Println()
	} else {
		ios.Info("No active discovery scan.")
	}

	if len(status.Errors) > 0 {
		ios.Println()
		ios.Error("Errors:")
		for _, e := range status.Errors {
			ios.Printf("  - %s\n", e)
		}
	}

	return nil
}

func showDeviceStatus(ctx context.Context, svc *shelly.Service, opts *Options, ios *iostreams.IOStreams) error {
	status, err := fetchDeviceStatus(ctx, svc, opts.Device, opts.ID)
	if err != nil {
		return err
	}

	if opts.JSON {
		return outputDeviceJSON(ios, status)
	}

	displayDeviceStatus(ios, status)
	return nil
}

func fetchDeviceStatus(ctx context.Context, svc *shelly.Service, device string, id int) (DeviceStatus, error) {
	var status DeviceStatus
	params := map[string]any{"id": id}

	err := svc.WithConnection(ctx, device, func(conn *client.Client) error {
		var err error
		status, err = getDeviceStatusRPC(ctx, conn, params)
		if err != nil {
			return err
		}

		name, addr := getDeviceConfigRPC(ctx, conn, params)
		status.Name = name
		status.Addr = addr

		status.KnownObjects = getKnownObjectsRPC(ctx, conn, params)
		return nil
	})

	return status, err
}

func getDeviceStatusRPC(ctx context.Context, conn *client.Client, params map[string]any) (DeviceStatus, error) {
	var status DeviceStatus

	result, err := conn.Call(ctx, "BTHomeDevice.GetStatus", params)
	if err != nil {
		return status, fmt.Errorf("failed to get device status: %w", err)
	}

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return status, fmt.Errorf("failed to marshal result: %w", err)
	}
	if err := json.Unmarshal(jsonBytes, &status); err != nil {
		return status, fmt.Errorf("failed to parse status: %w", err)
	}

	return status, nil
}

func getDeviceConfigRPC(ctx context.Context, conn *client.Client, params map[string]any) (name, addr string) {
	cfgResult, err := conn.Call(ctx, "BTHomeDevice.GetConfig", params)
	if err != nil {
		return "", ""
	}

	var cfg struct {
		Name *string `json:"name"`
		Addr string  `json:"addr"`
	}
	cfgBytes, err := json.Marshal(cfgResult)
	if err != nil {
		return "", ""
	}
	if json.Unmarshal(cfgBytes, &cfg) != nil {
		return "", ""
	}

	if cfg.Name != nil {
		name = *cfg.Name
	}
	return name, cfg.Addr
}

func getKnownObjectsRPC(ctx context.Context, conn *client.Client, params map[string]any) []KnownObject {
	objResult, err := conn.Call(ctx, "BTHomeDevice.GetKnownObjects", params)
	if err != nil {
		return nil
	}

	var objResp struct {
		Objects []KnownObject `json:"objects"`
	}
	objBytes, err := json.Marshal(objResult)
	if err != nil {
		return nil
	}
	if json.Unmarshal(objBytes, &objResp) != nil {
		return nil
	}

	return objResp.Objects
}

func outputDeviceJSON(ios *iostreams.IOStreams, status DeviceStatus) error {
	output, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format JSON: %w", err)
	}
	ios.Println(string(output))
	return nil
}

func displayDeviceStatus(ios *iostreams.IOStreams, status DeviceStatus) {
	name := status.Name
	if name == "" {
		name = fmt.Sprintf("Device %d", status.ID)
	}

	ios.Println(theme.Bold().Render(fmt.Sprintf("BTHome Device: %s", name)))
	ios.Println()

	displayDeviceBasicInfo(ios, status)
	displayKnownObjects(ios, status.KnownObjects)
	displayErrors(ios, status.Errors)
}

func displayDeviceBasicInfo(ios *iostreams.IOStreams, status DeviceStatus) {
	ios.Printf("  ID: %d\n", status.ID)
	if status.Addr != "" {
		ios.Printf("  Address: %s\n", status.Addr)
	}
	if status.RSSI != nil {
		ios.Printf("  RSSI: %d dBm\n", *status.RSSI)
	}
	if status.Battery != nil {
		ios.Printf("  Battery: %d%%\n", *status.Battery)
	}
	if status.PacketID != nil {
		ios.Printf("  Packet ID: %d\n", *status.PacketID)
	}
	if status.LastUpdateTS > 0 {
		lastUpdate := time.Unix(int64(status.LastUpdateTS), 0)
		ios.Printf("  Last Update: %s\n", lastUpdate.Format(time.RFC3339))
	}
}

func displayKnownObjects(ios *iostreams.IOStreams, objects []KnownObject) {
	if len(objects) == 0 {
		return
	}

	ios.Println()
	ios.Println("  " + theme.Highlight().Render("Known Objects:"))
	for _, obj := range objects {
		managed := ""
		if obj.Component != nil {
			managed = fmt.Sprintf(" (managed by %s)", *obj.Component)
		}
		ios.Printf("    Object ID: %d, Index: %d%s\n", obj.ObjID, obj.Idx, managed)
	}
}

func displayErrors(ios *iostreams.IOStreams, errors []string) {
	if len(errors) == 0 {
		return
	}

	ios.Println()
	ios.Error("Errors:")
	for _, e := range errors {
		ios.Printf("  - %s\n", e)
	}
}
