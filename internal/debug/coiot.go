// Package debug provides debug command helpers.
package debug

import (
	"context"
	"encoding/json"
	"time"

	"github.com/tj-smith47/shelly-go/gen1"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/term"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// CoIoTListenerOptions configures the CoIoT listener.
type CoIoTListenerOptions struct {
	Stream   bool
	Duration time.Duration
	Raw      bool
}

// RunCoIoTListener starts a CoIoT multicast listener and displays events.
func RunCoIoTListener(ctx context.Context, ios *iostreams.IOStreams, opts CoIoTListenerOptions) error {
	ios.Println(theme.Bold().Render("CoIoT Multicast Listener:"))
	ios.Printf("  Joining multicast group %s:%d\n", gen1.CoIoTMulticastAddr, gen1.CoIoTPort)
	ios.Println()

	listener := gen1.NewCoIoTListener()

	eventCount := 0
	listener.OnStatus(func(deviceID string, status *gen1.CoIoTStatus) {
		eventCount++
		timestamp := time.Now().Format("15:04:05")

		if opts.Raw {
			jsonOutput, err := json.Marshal(status)
			if err != nil {
				ios.Printf("[%s] %s: (marshal error: %v)\n", timestamp, deviceID, err)
				return
			}
			ios.Printf("[%s] %s: %s\n", timestamp, deviceID, string(jsonOutput))
		} else {
			term.DisplayCoIoTEvent(ios, timestamp, deviceID, status)
		}
	})

	if err := listener.Start(); err != nil {
		return err
	}
	defer func() {
		if err := listener.Stop(); err != nil {
			ios.DebugErr("stop coiot listener", err)
		}
	}()

	ios.Success("Listening for CoIoT multicast updates...")
	ios.Println()

	// Wait for completion
	if opts.Stream {
		ios.Info("Streaming indefinitely (press Ctrl+C to stop)...")
		ios.Println()
		<-ctx.Done()
		ios.Println()
		ios.Info("Stopped by user")
	} else {
		ios.Info("Listening for %s (press Ctrl+C to stop)...", opts.Duration)
		ios.Println()
		select {
		case <-ctx.Done():
			ios.Println()
			ios.Info("Stopped by user")
		case <-time.After(opts.Duration):
			ios.Println()
			ios.Info("Duration reached")
		}
	}

	ios.Println()
	ios.Success("Received %d CoIoT messages", eventCount)
	return nil
}
