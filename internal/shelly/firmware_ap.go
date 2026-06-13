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

// prefetchAPFirmware downloads, before the host hops onto the device's factory AP,
// the Gen1 firmware image to flash at that AP — for a --to-ap restore that requested
// a firmware update on a Gen1 target. It resolves the image URL (an explicit
// --firmware-url override, else one derived from the backup's model) and downloads
// it now, while the host still has internet (the factory AP has none). It returns an
// empty path and no error when no at-AP update applies (not a firmware update, or not
// Gen1), so the caller can branch on the path. The caller removes the returned file.
func prefetchAPFirmware(
	ctx context.Context,
	generation int,
	bkp *backup.DeviceBackup,
	opts backup.RestoreOptions,
) (string, error) {
	if !opts.UpdateFirmware || generation != 1 {
		return "", nil
	}
	fwURL := opts.FirmwareURL
	if fwURL == "" {
		fwURL = shellybackup.OfficialGen1FirmwareURL(bkp.Device().Model)
	}
	if fwURL == "" {
		return "", fmt.Errorf(
			"cannot update firmware: the backup carries no device model to derive a firmware " +
				"URL from and none was supplied — set --firmware-url")
	}
	return fetchGen1Firmware(ctx, fwURL)
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

// updateGen1FirmwareAtAP flashes the device at its factory AP to the backup's
// firmware before the config restore, for a device whose build is older than the
// backup's — notably a corrupt build that reboot-loops the instant WiFi station
// mode is active and so can never complete an OTA once on the LAN. The device is
// stable at its AP but has no internet, so the image (already downloaded to fwPath)
// is re-served from the host's AP-subnet address and the device is pointed at it.
// It is a no-op when the device is already at or beyond the backup's firmware.
//
// Must be called while the host is hopped onto the device's AP (so bindIP is live
// and the device is reachable at discovery.DefaultAPIP), before any station config
// is written.
func (s *Service) updateGen1FirmwareAtAP(ctx context.Context, bindIP, fwPath, backupFW string) error {
	fwURL, stop, err := serveFirmwareFile(ctx, bindIP, fwPath)
	if err != nil {
		return err
	}
	defer stop()

	return s.WithGen1Connection(ctx, discovery.DefaultAPIP, func(conn *client.Gen1Client) error {
		dev := conn.Device()
		liveFW := shellybackup.Gen1LiveFirmware(ctx, dev)
		if liveFW == "" {
			return errors.New("could not read the device's firmware at its AP to decide on an update")
		}
		if !shellybackup.Gen1FirmwareDowngrade(liveFW, backupFW) {
			debug.TraceEvent("firmware-at-ap: device on %q already >= backup %q; skipping OTA", liveFW, backupFW)
			return nil
		}
		debug.TraceEvent("firmware-at-ap: serving %s, OTA from %q toward backup %q", fwURL, liveFW, backupFW)
		if updErr := shellybackup.UpdateGen1FirmwareAndWait(ctx, dev, fwURL, liveFW); updErr != nil {
			return fmt.Errorf("at-AP firmware update failed: %w", updErr)
		}
		return nil
	})
}
