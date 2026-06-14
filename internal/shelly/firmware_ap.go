package shelly

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	shellybackup "github.com/tj-smith47/shelly-go/backup"
	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/shelly/backup"
	"github.com/tj-smith47/shelly-cli/internal/tui/debug"
)

// firmwareServePort is the port the host serves a Gen1 firmware image on during an
// at-AP recovery; the device fetches http://<apHostIP>:<port>/firmware.zip from it.
const firmwareServePort = 8512

// firmwareFetchTimeout bounds downloading the image from the public CDN.
const firmwareFetchTimeout = 90 * time.Second

// fetchGen1Firmware downloads the Gen1 firmware image at url to a temp file and
// returns its path. It runs BEFORE the host hops onto the device's AP, while the
// host still has internet — the factory AP itself has none. The caller removes the
// returned file when done.
func fetchGen1Firmware(ctx context.Context, url string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, firmwareFetchTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("build firmware request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("download firmware from %s: %w", url, err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			debug.TraceEvent("firmware-at-ap: close download body: %v", cerr)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download firmware from %s: unexpected status %s", url, resp.Status)
	}

	f, err := os.CreateTemp("", "shelly-gen1-fw-*.zip")
	if err != nil {
		return "", fmt.Errorf("create firmware temp file: %w", err)
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		if cerr := f.Close(); cerr != nil {
			debug.TraceEvent("firmware-at-ap: close temp after copy error: %v", cerr)
		}
		removeFirmwareTemp(f.Name())
		return "", fmt.Errorf("write firmware temp file: %w", err)
	}
	if err := f.Close(); err != nil {
		removeFirmwareTemp(f.Name())
		return "", fmt.Errorf("close firmware temp file: %w", err)
	}
	return f.Name(), nil
}

// apFirmwareBindIP resolves the host's actual AP-subnet address for serving the
// firmware image. withAPHop keeps discovery.DefaultAPHostIP when --ap-ip is unset
// (an empty apHostIP means "use the default"), so the same resolution must happen
// here — otherwise the served URL would carry no host (http://:8512/...) and the
// device at its AP could never fetch the image.
func apFirmwareBindIP(apHostIP string) string {
	if apHostIP == "" {
		return discovery.DefaultAPHostIP
	}
	return apHostIP
}

// prefetchAPFirmware downloads, before the host hops onto the device's factory AP,
// the Gen1 firmware image to flash at that AP — for a --to-ap restore that requested
// a possible firmware update on a Gen1 target. It resolves the image URL (an explicit
// --firmware-url override, else one derived from the backup's model) and downloads it
// now, while the host still has internet (the factory AP has none). It is best-effort:
// it returns an empty path whenever no image can or should be staged (a Gen2 target, a
// forced downgrade, an underivable URL, or a failed download), leaving the at-AP check
// to fail loudly only if an update turns out to be required. The caller removes the
// returned file.
func prefetchAPFirmware(
	ctx context.Context,
	generation int,
	bkp *backup.DeviceBackup,
	opts backup.RestoreOptions,
) string {
	// A firmware update is automatic: whether the device needs one is only knowable
	// at the AP (its live build cannot be read until the host has hopped onto it),
	// and the factory AP has no internet — so the image must be fetched NOW, before
	// the hop, while the host still has connectivity. The at-AP check then decides
	// whether to actually flash. AllowFirmwareDowngrade opts out (force the older
	// config write, no update), so skip the prefetch then.
	if generation != 1 || opts.AllowFirmwareDowngrade {
		return ""
	}
	fwURL := opts.FirmwareURL
	if fwURL == "" {
		fwURL = shellybackup.OfficialGen1FirmwareURL(bkp.Device().Model)
	}
	if fwURL == "" {
		// No derivable URL: cannot pre-stage an image. Not fatal here — if the device
		// turns out to need an update, ensureGen1FirmwareAtAP fails loudly at the AP.
		debug.TraceEvent("firmware-at-ap: no firmware URL derivable from model %q; skipping prefetch", bkp.Device().Model)
		return ""
	}
	// Best-effort: a download hiccup must not block a restore that may not even need
	// an update. If an update IS needed and the image is missing, the at-AP check
	// surfaces it loudly.
	fwPath, err := fetchGen1Firmware(ctx, fwURL)
	if err != nil {
		debug.TraceEvent("firmware-at-ap: prefetch of %s failed (continuing; at-AP check decides if it mattered): %v", fwURL, err)
		return ""
	}
	return fwPath
}

// removeFirmwareTemp deletes a downloaded firmware temp file, logging a failure
// rather than propagating it (cleanup is best-effort).
func removeFirmwareTemp(path string) {
	if path == "" {
		return
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		debug.TraceEvent("firmware-at-ap: remove temp image %s: %v", path, err)
	}
}

// serveFirmwareFile starts an HTTP server bound to bindIP that serves the image at
// path as /firmware.zip, returning the URL the device should fetch and a stop
// function. A device at its factory AP has no internet, so the image is served from
// the host's own address on the AP subnet (bindIP, the --ap-ip the host took when
// it hopped onto the AP).
func serveFirmwareFile(ctx context.Context, bindIP, path string) (url string, stop func(), err error) {
	addr := net.JoinHostPort(bindIP, strconv.Itoa(firmwareServePort))
	var lc net.ListenConfig
	ln, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return "", nil, fmt.Errorf("listen on %s to serve firmware: %w", addr, err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/firmware.zip", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path)
	})
	srv := &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	go func() {
		if serveErr := srv.Serve(ln); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			debug.TraceEvent("firmware-at-ap: serve firmware: %v", serveErr)
		}
	}()

	stop = func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if shutErr := srv.Shutdown(shutdownCtx); shutErr != nil {
			debug.TraceEvent("firmware-at-ap: shut down firmware serve: %v", shutErr)
		}
	}
	return "http://" + addr + "/firmware.zip", stop, nil
}

// ensureGen1FirmwareAtAP brings the device at its factory AP up to the backup's
// firmware before the config restore when its build is older than the backup's —
// notably a corrupt build that reboot-loops the instant WiFi station mode is active
// and so can never complete an OTA once on the LAN. Whether an update is needed is
// decided here, at the AP, because the device's live build cannot be read until the
// host has hopped onto it. It is a no-op when the device is already at or beyond the
// backup's firmware, and skips the update entirely when allowDowngrade forces the
// older config write.
//
// The device is stable at its AP but has no internet, so the image (prefetched to
// fwPath before the hop) is re-served from the host's AP-subnet address and the
// device is pointed at it. If an update is needed but no image was prefetched (an
// underivable URL or a failed download), this fails loudly rather than writing a
// station config that would reboot-loop the device onto a dead LAN.
//
// Must be called while the host is hopped onto the device's AP (so bindIP is live
// and the device is reachable at discovery.DefaultAPIP), before any station config
// is written.
func (s *Service) ensureGen1FirmwareAtAP(ctx context.Context, bindIP, fwPath, backupFW string, allowDowngrade bool) error {
	return s.WithGen1Connection(ctx, discovery.DefaultAPIP, func(conn *client.Gen1Client) error {
		dev := conn.Device()
		liveFW := shellybackup.Gen1LiveFirmware(ctx, dev)
		if liveFW == "" {
			return errors.New("could not read the device's firmware at its AP to decide on an update")
		}
		if allowDowngrade || !shellybackup.Gen1FirmwareDowngrade(liveFW, backupFW) {
			debug.TraceEvent("firmware-at-ap: device on %q vs backup %q; no update (allowDowngrade=%v)", liveFW, backupFW, allowDowngrade)
			return nil
		}
		if fwPath == "" {
			return fmt.Errorf(
				"device on firmware %q needs an update to the backup's %q before restore, but no "+
					"firmware image is available (the factory AP has no internet, so the image is "+
					"prefetched before the hop — its URL was underivable or the download failed); "+
					"retry with connectivity, pass --firmware-url, or --allow-firmware-downgrade to "+
					"force the downgrade and accept the reboot-loop risk",
				liveFW, backupFW)
		}
		fwURL, stop, err := serveFirmwareFile(ctx, bindIP, fwPath)
		if err != nil {
			return err
		}
		defer stop()
		debug.TraceEvent("firmware-at-ap: serving %s, OTA from %q toward backup %q", fwURL, liveFW, backupFW)
		if updErr := shellybackup.UpdateGen1FirmwareAndWait(ctx, dev, fwURL, liveFW); updErr != nil {
			return fmt.Errorf("at-AP firmware update failed: %w", updErr)
		}
		return nil
	})
}

// confirmGen1StableAtAP verifies the Gen1 device is booted and holding at its
// factory AP — not caught in a reboot loop — immediately before the station config
// write. That write reboots the device, and on firmware that cannot survive station
// mode it would strand it off the LAN, so gating it on confirmed stability means the
// one write that can brick is never issued to a device that is not provably healthy.
// The device is still on its recoverable AP here, so a failure aborts the restore
// with the device intact rather than bricked. ensureGen1FirmwareAtAP runs first, so
// an unstable device at this point is one that did not come up clean even on matched
// firmware — precisely the device a station write would lose.
//
// Must be called while the host is hopped onto the device's AP, after
// ensureGen1FirmwareAtAP and before any station config is written.
func (s *Service) confirmGen1StableAtAP(ctx context.Context) error {
	return s.WithGen1Connection(ctx, discovery.DefaultAPIP, func(conn *client.Gen1Client) error {
		uptime, required, stable := shellybackup.Gen1ConfirmStable(ctx, conn.Device())
		if !stable {
			return fmt.Errorf(
				"refusing to write the station config: the device is not stable at its factory AP "+
					"(highest uptime %ds, need %ds held — the signature of a reboot loop); writing it now "+
					"would reboot the device onto firmware it cannot hold and strand it off the LAN. The "+
					"device remains on its recoverable factory AP",
				uptime, required)
		}
		debug.TraceEvent("firmware-at-ap: device stable at AP (uptime %ds >= %ds); safe to write station config", uptime, required)
		return nil
	})
}
