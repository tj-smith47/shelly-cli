package shelly

import (
	"context"
	"fmt"
	"os"

	"github.com/tj-smith47/shelly-go/integrator"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// IntegratorCredentials holds integrator API credentials.
type IntegratorCredentials struct {
	Tag   string
	Token string
}

// GetIntegratorCredentials retrieves integrator credentials from environment or config.
func GetIntegratorCredentials(ios *iostreams.IOStreams, cfg *config.Config) (*IntegratorCredentials, error) {
	tag := os.Getenv("SHELLY_INTEGRATOR_TAG")
	token := os.Getenv("SHELLY_INTEGRATOR_TOKEN")

	if cfg != nil {
		if tag == "" {
			tag = cfg.Integrator.Tag
		}
		if token == "" {
			token = cfg.Integrator.Token
		}
	}

	if tag == "" || token == "" {
		return nil, fmt.Errorf("integrator credentials required. Run 'shelly fleet connect' first or set SHELLY_INTEGRATOR_TAG and SHELLY_INTEGRATOR_TOKEN")
	}

	return &IntegratorCredentials{Tag: tag, Token: token}, nil
}

// FleetConnection wraps a fleet manager with connection state.
type FleetConnection struct {
	Client  *integrator.Client
	Manager *integrator.FleetManager
	ios     *iostreams.IOStreams
}

// ConnectFleet creates an authenticated fleet manager and connects to all hosts.
func ConnectFleet(ctx context.Context, ios *iostreams.IOStreams, creds *IntegratorCredentials) (*FleetConnection, error) {
	// Create client and authenticate
	client := integrator.New(creds.Tag, creds.Token)
	if err := client.Authenticate(ctx); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Create fleet manager and connect
	fm := integrator.NewFleetManager(client)

	ios.Info("Connecting to fleet...")
	connErrors := fm.ConnectAll(ctx, nil)
	for host, err := range connErrors {
		ios.Warning("Failed to connect to %s: %v", host, err)
	}

	return &FleetConnection{
		Client:  client,
		Manager: fm,
		ios:     ios,
	}, nil
}

// Close disconnects from all hosts.
func (fc *FleetConnection) Close() {
	if err := fc.Manager.DisconnectAll(); err != nil {
		fc.ios.DebugErr("disconnect", err)
	}
}

// ReportBatchResults reports batch command results and returns an error if any failed.
func ReportBatchResults(ios *iostreams.IOStreams, results []integrator.BatchResult, action string) error {
	successCount := 0
	failCount := 0

	for _, r := range results {
		if r.Success {
			successCount++
			ios.Printf("  %s: %s\n", r.DeviceID, action)
		} else {
			failCount++
			ios.Error("  %s: failed - %s", r.DeviceID, r.Error)
		}
	}

	if failCount > 0 {
		ios.Warning("%d device(s) failed, %d succeeded", failCount, successCount)
		return fmt.Errorf("%d device(s) failed to %s", failCount, action)
	}

	if successCount > 0 {
		ios.Success("%s %d device(s)", capitalizeFirst(action), successCount)
	} else {
		ios.Warning("No devices were affected")
	}

	return nil
}

// capitalizeFirst returns the string with the first letter capitalized.
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	if s[0] >= 'a' && s[0] <= 'z' {
		return string(s[0]-32) + s[1:]
	}
	return s
}
