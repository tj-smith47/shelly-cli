package shelly

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	shellybackup "github.com/tj-smith47/shelly-go/backup"
	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/shelly/backup"
)

const fakeFirmwareBody = "PK\x03\x04 fake gen1 firmware image"

// fakeFirmwareServer serves fakeFirmwareBody at any path, like the Shelly CDN does
// for a Gen1 image, so the fetch path can be exercised without the network.
func fakeFirmwareServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if _, err := io.WriteString(w, fakeFirmwareBody); err != nil {
			t.Errorf("serve fake firmware: %v", err)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestFetchGen1Firmware_DownloadsToTemp(t *testing.T) {
	t.Parallel()
	srv := fakeFirmwareServer(t)

	path, err := fetchGen1Firmware(context.Background(), srv.URL+"/gen1/SHBDUO-1.zip")
	if err != nil {
		t.Fatalf("fetchGen1Firmware: %v", err)
	}
	defer removeFirmwareTemp(path)

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read temp image: %v", err)
	}
	if string(got) != fakeFirmwareBody {
		t.Errorf("temp image = %q, want %q", got, fakeFirmwareBody)
	}
}

func TestFetchGen1Firmware_Non200Errors(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "nope", http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	if _, err := fetchGen1Firmware(context.Background(), srv.URL); err == nil {
		t.Fatal("expected an error for a non-200 firmware response")
	}
}

func TestRemoveFirmwareTemp_IdempotentOnMissing(t *testing.T) {
	t.Parallel()
	// A blank path and an already-removed file must both be no-ops, not panics —
	// cleanup runs on defer even when the fetch never created a file.
	removeFirmwareTemp("")
	removeFirmwareTemp("/no/such/shelly-gen1-fw-deadbeef.zip")
}

//nolint:paralleltest // serially bound: shares the fixed production port 8512
func TestServeFirmwareFile_ServesImage(t *testing.T) {
	// Not parallel: serveFirmwareFile binds the fixed production port firmwareServePort
	// (8512), which the other at-AP serve tests also bind; running them sequentially
	// avoids a port clash on that shared address.
	// Stage a real on-disk image via the production fetch path; serveFirmwareFile
	// hands it to http.ServeFile, which reads the real filesystem, so an in-memory
	// FS would not exercise the serve.
	src := fakeFirmwareServer(t)
	imgPath, err := fetchGen1Firmware(context.Background(), src.URL+"/fw.zip")
	if err != nil {
		t.Fatalf("stage firmware image: %v", err)
	}
	defer removeFirmwareTemp(imgPath)

	url, stop, err := serveFirmwareFile(context.Background(), "127.0.0.1", imgPath)
	if err != nil {
		t.Fatalf("serveFirmwareFile: %v", err)
	}
	defer stop()

	if !strings.HasSuffix(url, "/firmware.zip") {
		t.Errorf("served URL %q does not end in /firmware.zip", url)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, http.NoBody)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			t.Errorf("close body: %v", cerr)
		}
	}()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read served body: %v", err)
	}
	if string(body) != fakeFirmwareBody {
		t.Errorf("served body = %q, want %q", body, fakeFirmwareBody)
	}
}

func TestAPFirmwareBindIP(t *testing.T) {
	t.Parallel()
	// Regression: an empty --ap-ip must resolve to the same default the AP hop uses,
	// not stay "" — a "" bind produced http://:8512/firmware.zip, which the device
	// could not fetch, so the at-AP OTA timed out with the device still on its build.
	if got := apFirmwareBindIP(""); got != discovery.DefaultAPHostIP {
		t.Errorf("apFirmwareBindIP(\"\") = %q, want default %q", got, discovery.DefaultAPHostIP)
	}
	if got := apFirmwareBindIP("192.168.33.150"); got != "192.168.33.150" {
		t.Errorf("apFirmwareBindIP(explicit) = %q, want it unchanged", got)
	}
}

func TestPrefetchAPFirmware_NoOpBranches(t *testing.T) {
	t.Parallel()
	bkp := &backup.DeviceBackup{Backup: &shellybackup.Backup{
		DeviceInfo: &shellybackup.DeviceInfo{Model: "SHBDUO-1", Generation: 1},
	}}

	// A Gen2 target -> the at-AP Gen1 OTA does not apply, no download attempted.
	if path := prefetchAPFirmware(context.Background(), 2, bkp, backup.RestoreOptions{}); path != "" {
		t.Errorf("gen2: path=%q, want empty", path)
	}

	// AllowFirmwareDowngrade forces the older config write -> no firmware update is
	// intended, so nothing is prefetched even for a Gen1 target.
	if path := prefetchAPFirmware(context.Background(), 1, bkp,
		backup.RestoreOptions{AllowFirmwareDowngrade: true}); path != "" {
		t.Errorf("allow-downgrade: path=%q, want empty", path)
	}
}

func TestPrefetchAPFirmware_NoModelSkips(t *testing.T) {
	t.Parallel()
	// Gen1, but no model to derive a URL from and no override: nothing can be
	// prefetched. This is best-effort, not fatal — if the device turns out to need an
	// update, ensureGen1FirmwareAtAP fails loudly at the AP. So return empty/nil here.
	bkp := &backup.DeviceBackup{Backup: &shellybackup.Backup{}}
	if path := prefetchAPFirmware(context.Background(), 1, bkp, backup.RestoreOptions{}); path != "" {
		t.Fatalf("no-model: path=%q, want empty", path)
	}
}

func TestPrefetchAPFirmware_DownloadFailureIsBestEffort(t *testing.T) {
	t.Parallel()
	// A download failure must NOT block a restore that may not even need an update:
	// prefetch swallows it and returns empty/nil, leaving the at-AP check to fail
	// loudly only if an update is actually required.
	bkp := &backup.DeviceBackup{Backup: &shellybackup.Backup{}}
	if path := prefetchAPFirmware(context.Background(), 1, bkp,
		backup.RestoreOptions{FirmwareURL: "http://127.0.0.1:1/does-not-exist.zip"}); path != "" {
		t.Fatalf("download failure: path=%q, want empty (best-effort)", path)
	}
}

func TestPrefetchAPFirmware_ExplicitURLDownloads(t *testing.T) {
	t.Parallel()
	srv := fakeFirmwareServer(t)
	bkp := &backup.DeviceBackup{Backup: &shellybackup.Backup{}}

	path := prefetchAPFirmware(context.Background(), 1, bkp,
		backup.RestoreOptions{FirmwareURL: srv.URL + "/fw.zip"})
	defer removeFirmwareTemp(path)
	if path == "" {
		t.Fatal("expected a downloaded firmware path, got empty")
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read downloaded image: %v", err)
	}
	if string(got) != fakeFirmwareBody {
		t.Errorf("downloaded image = %q, want %q", got, fakeFirmwareBody)
	}
}

// stageFirmwareImage downloads a fake image to a temp file the at-AP serve path can
// read, registering cleanup. ensureGen1FirmwareAtAP serves it via http.ServeFile,
// which reads the real filesystem, so the image must exist on disk.
func stageFirmwareImage(t *testing.T) string {
	t.Helper()
	src := fakeFirmwareServer(t)
	path, err := fetchGen1Firmware(context.Background(), src.URL+"/fw.zip")
	if err != nil {
		t.Fatalf("stage firmware image: %v", err)
	}
	t.Cleanup(func() { removeFirmwareTemp(path) })
	return path
}

// TestEnsureGen1FirmwareAtAP_NoUpdateWhenCurrent covers the no-op: a device already
// at or beyond the backup's firmware needs no flash, so the function returns nil
// without ever reaching for an image. The fake reports a newer build than the backup.
func TestEnsureGen1FirmwareAtAP_NoUpdateWhenCurrent(t *testing.T) {
	t.Parallel()
	d := newAPDevServer(t, 1)
	d.gen1.setFW("20230101-000000/v2.0") // device newer than the backup below
	svc := apdevService(d, 1)

	err := svc.ensureGen1FirmwareAtAP(context.Background(), d.addr(), "", "20210101-000000/v1.0", false)
	if err != nil {
		t.Fatalf("ensureGen1FirmwareAtAP (no update needed): %v", err)
	}
	if atomic.LoadInt32(&d.gen1.otaHits) != 0 {
		t.Error("an OTA was triggered even though the device was already current")
	}
}

// TestEnsureGen1FirmwareAtAP_AllowDowngradeSkips covers the opt-out: allowDowngrade
// forces the older config write, so no firmware update is attempted even when the
// device's build predates the backup's.
func TestEnsureGen1FirmwareAtAP_AllowDowngradeSkips(t *testing.T) {
	t.Parallel()
	d := newAPDevServer(t, 1)
	d.gen1.setFW("20200101-000000/v0.9") // older than the backup
	svc := apdevService(d, 1)

	err := svc.ensureGen1FirmwareAtAP(context.Background(), d.addr(), "", "20210601-000000/v1.5", true)
	if err != nil {
		t.Fatalf("ensureGen1FirmwareAtAP (allow downgrade): %v", err)
	}
	if atomic.LoadInt32(&d.gen1.otaHits) != 0 {
		t.Error("an OTA was triggered despite allowDowngrade forcing the older write")
	}
}

// TestEnsureGen1FirmwareAtAP_UpdateNeededNoImage covers the loud failure: an update is
// required (device older than the backup) but no image was prefetched, so the function
// must refuse rather than write a station config that would reboot-loop the device.
func TestEnsureGen1FirmwareAtAP_UpdateNeededNoImage(t *testing.T) {
	t.Parallel()
	d := newAPDevServer(t, 1)
	d.gen1.setFW("20200101-000000/v0.9")
	svc := apdevService(d, 1)

	err := svc.ensureGen1FirmwareAtAP(context.Background(), d.addr(), "", "20210601-000000/v1.5", false)
	if err == nil {
		t.Fatal("expected a loud refusal when an update is needed but no image is staged")
	}
	if !strings.Contains(err.Error(), "no") || !strings.Contains(err.Error(), "firmware image") {
		t.Errorf("err = %q, want it to name the missing firmware image", err)
	}
}

// TestEnsureGen1FirmwareAtAP_UnreadableFirmware covers the guard: if the device's live
// build cannot be read at the AP, the update decision cannot be made safely, so the
// function errors rather than guessing.
func TestEnsureGen1FirmwareAtAP_UnreadableFirmware(t *testing.T) {
	t.Parallel()
	d := newAPDevServer(t, 1)
	svc := apdevService(d, 1)
	// /shelly answers (so the connection establishes) but /settings 500s, so the live
	// firmware read returns "" and the update decision must abort. A short deadline
	// caps the SDK's retry-with-backoff on the 500 — the read returns "" either way.
	d.gen1.settingsErr = true
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	err := svc.ensureGen1FirmwareAtAP(ctx, d.addr(), "", "20210601-000000/v1.5", false)
	if err == nil {
		t.Fatal("expected an error when the device's firmware cannot be read at its AP")
	}
	if !strings.Contains(err.Error(), "firmware") {
		t.Errorf("err = %q, want it to mention the unreadable firmware", err)
	}
}

// TestEnsureGen1FirmwareAtAP_FlashesAndWaits covers the happy OTA path: the device is
// older than the backup and an image is staged, so the function serves the image,
// triggers the OTA, and the post-flash wait observes the new build. The fake flips its
// reported firmware to the backup's the instant /ota is hit, modelling the flash.
//
//nolint:paralleltest // serially bound: shares the fixed production port 8512
func TestEnsureGen1FirmwareAtAP_FlashesAndWaits(t *testing.T) {
	// Not parallel: this serves the image on the fixed production port 8512 (see
	// TestServeFirmwareFile_ServesImage).
	const backupFW = "20210601-000000/v1.5"
	d := newAPDevServer(t, 1)
	d.gen1.setFW("20200101-000000/v0.9") // older -> downgrade -> flash needed
	d.gen1.otaFlipFW = backupFW          // the flash lands the backup's build
	d.gen1.uptime = 99                   // already stable, so the wait returns at once
	svc := apdevService(d, 1)
	imgPath := stageFirmwareImage(t)

	err := svc.ensureGen1FirmwareAtAP(context.Background(), "127.0.0.1", imgPath, backupFW, false)
	if err != nil {
		t.Fatalf("ensureGen1FirmwareAtAP (flash): %v", err)
	}
	if atomic.LoadInt32(&d.gen1.otaHits) == 0 {
		t.Error("the device's /ota endpoint was never triggered")
	}
	if got := d.gen1.currentFW(); got != backupFW {
		t.Errorf("device firmware after flash = %q, want %q", got, backupFW)
	}
}

// TestEnsureGen1FirmwareAtAP_OTAFailurePropagates covers the flash-failure branch: an
// OTA that never lands the new build must surface as a loud error, not a silent pass
// that would then write a station config onto unflashed firmware. The fake refuses the
// OTA and never changes its build; a short deadline caps the post-trigger wait so the
// "did not complete" failure returns without the full update budget.
//
//nolint:paralleltest // serially bound: shares the fixed production port 8512
func TestEnsureGen1FirmwareAtAP_OTAFailurePropagates(t *testing.T) {
	// Not parallel: serves on the fixed production port 8512 (see
	// TestServeFirmwareFile_ServesImage).
	const backupFW = "20210601-000000/v1.5"
	d := newAPDevServer(t, 1)
	d.gen1.setFW("20200101-000000/v0.9") // older -> flash required
	d.gen1.otaErr = true                 // the trigger errors and the build never changes
	svc := apdevService(d, 1)
	imgPath := stageFirmwareImage(t)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	err := svc.ensureGen1FirmwareAtAP(ctx, "127.0.0.1", imgPath, backupFW, false)
	if err == nil {
		t.Fatal("expected an error when the at-AP OTA never lands the new build")
	}
}

// TestEnsureGen1FirmwareAtAP_ServeBindFailurePropagates covers the serve-bind failure
// branch: when a flash is needed but the image cannot be served (an unbindable host
// address), ensureGen1FirmwareAtAP returns that error rather than proceeding to an OTA
// the device could never fetch. The bind address is unassignable, so no real port is
// taken — safe to run in parallel.
func TestEnsureGen1FirmwareAtAP_ServeBindFailurePropagates(t *testing.T) {
	t.Parallel()
	const backupFW = "20210601-000000/v1.5"
	d := newAPDevServer(t, 1)
	d.gen1.setFW("20200101-000000/v0.9") // older -> flash required
	svc := apdevService(d, 1)
	imgPath := stageFirmwareImage(t)

	// 192.0.2.1 (TEST-NET-1) is not a local address, so the firmware listener cannot
	// bind it and serveFirmwareFile fails before any OTA is triggered.
	err := svc.ensureGen1FirmwareAtAP(context.Background(), "192.0.2.1", imgPath, backupFW, false)
	if err == nil {
		t.Fatal("expected the serve-bind failure to propagate")
	}
	if atomic.LoadInt32(&d.gen1.otaHits) != 0 {
		t.Error("an OTA was triggered despite the image never being served")
	}
}

// TestServeFirmwareFile_ListenError covers the bind failure: a malformed bind address
// cannot be listened on, so serveFirmwareFile returns an error rather than a half-open
// server.
func TestServeFirmwareFile_ListenError(t *testing.T) {
	t.Parallel()
	// An unassignable address (TEST-NET-1) cannot be bound, so Listen fails.
	_, stop, err := serveFirmwareFile(context.Background(), "192.0.2.1", "/tmp/whatever.zip")
	if err == nil {
		if stop != nil {
			stop()
		}
		t.Fatal("expected a listen error binding an unassignable address")
	}
}

// TestFetchGen1Firmware_TempCreateError covers the temp-file creation failure: with
// TMPDIR pointing at a path that cannot hold a file, os.CreateTemp fails after a
// successful download, so the fetch returns that error. Not parallel: it mutates the
// process TMPDIR via t.Setenv.
func TestFetchGen1Firmware_TempCreateError(t *testing.T) {
	srv := fakeFirmwareServer(t)
	// A TMPDIR under a regular file (not a directory) makes CreateTemp fail with ENOTDIR.
	notDir := t.TempDir() + "/afile"
	if err := os.WriteFile(notDir, []byte("x"), 0o600); err != nil {
		t.Fatalf("seed non-dir: %v", err)
	}
	t.Setenv("TMPDIR", notDir+"/nope")

	if _, err := fetchGen1Firmware(context.Background(), srv.URL+"/fw.zip"); err == nil {
		t.Fatal("expected a temp-file creation error when TMPDIR is unusable")
	}
}

// TestFetchGen1Firmware_BodyCopyError covers the write-temp failure path: a server
// that promises more bytes than it sends and then drops the connection makes io.Copy
// fail mid-stream, so the fetch removes the partial temp file and returns an error.
func TestFetchGen1Firmware_BodyCopyError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Advertise a long body, send a few bytes, then hijack and close so the client's
		// read fails before the advertised length is reached.
		w.Header().Set("Content-Length", "4096")
		w.WriteHeader(http.StatusOK)
		if _, err := io.WriteString(w, "PK\x03\x04 partial"); err != nil {
			return
		}
		hj, ok := w.(http.Hijacker)
		if !ok {
			return
		}
		conn, _, err := hj.Hijack()
		if err != nil {
			return
		}
		if cerr := conn.Close(); cerr != nil {
			t.Logf("close hijacked conn: %v", cerr)
		}
	}))
	t.Cleanup(srv.Close)

	if _, err := fetchGen1Firmware(context.Background(), srv.URL+"/fw.zip"); err == nil {
		t.Fatal("expected an error when the firmware body is truncated mid-download")
	}
}

// TestRemoveFirmwareTemp_RemovalErrorIsTraced covers the error-trace branch: a remove
// that fails for a reason other than "already gone" (here, a non-empty directory) is
// logged, not panicked or propagated — cleanup is best-effort.
func TestRemoveFirmwareTemp_RemovalErrorIsTraced(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// os.Remove on a non-empty directory fails with ENOTEMPTY (not ErrNotExist), so the
	// trace branch runs. The call must still return normally.
	if err := os.WriteFile(dir+"/keep", []byte("x"), 0o600); err != nil {
		t.Fatalf("seed dir: %v", err)
	}
	removeFirmwareTemp(dir)
}

// TestRemoveFirmwareTemp_RealFile covers the successful removal of an existing temp
// image (the path the cleanup defer takes after a real prefetch), distinct from the
// blank/missing no-ops already covered.
func TestRemoveFirmwareTemp_RealFile(t *testing.T) {
	t.Parallel()
	f, err := os.CreateTemp(t.TempDir(), "shelly-gen1-fw-*.zip")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	name := f.Name()
	if cerr := f.Close(); cerr != nil {
		t.Fatalf("close temp: %v", cerr)
	}
	removeFirmwareTemp(name)
	if _, statErr := os.Stat(name); !os.IsNotExist(statErr) {
		t.Errorf("temp image still present after removeFirmwareTemp: %v", statErr)
	}
}

// TestFetchGen1Firmware_BadURLErrors covers the request-build / dial failure branch: a
// malformed URL cannot produce a request, so the fetch errors before any download.
func TestFetchGen1Firmware_BadURLErrors(t *testing.T) {
	t.Parallel()
	if _, err := fetchGen1Firmware(context.Background(), "://not a url"); err == nil {
		t.Fatal("expected an error building a request from a malformed URL")
	}
}

// TestConfirmGen1StableAtAP_Stable covers the gate's pass: a device holding a stable
// uptime at its AP clears the bar, so the station write is allowed (nil error).
func TestConfirmGen1StableAtAP_Stable(t *testing.T) {
	t.Parallel()
	d := newAPDevServer(t, 1)
	d.gen1.uptime = 99 // well above the stability bar
	svc := apdevService(d, 1)

	if err := svc.confirmGen1StableAtAP(context.Background()); err != nil {
		t.Fatalf("confirmGen1StableAtAP (stable device): %v", err)
	}
}

// TestConfirmGen1StableAtAP_Unstable covers the gate's refusal: a device that cannot
// hold a stable uptime — the signature of a reboot loop — must NOT receive the station
// write, so the function returns a loud error. The context is cancelled just after the
// connection establishes, so the stability poll exits on its next tick instead of
// running its full recovery budget; the device never reaches the bar, so the gate
// refuses regardless.
func TestConfirmGen1StableAtAP_Unstable(t *testing.T) {
	t.Parallel()
	d := newAPDevServer(t, 1)
	d.gen1.uptime = 1 // below the bar: looks like a reboot loop
	svc := apdevService(d, 1)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	err := svc.confirmGen1StableAtAP(ctx)
	if err == nil {
		t.Fatal("expected a refusal when the device is not stable at its AP")
	}
	if !strings.Contains(err.Error(), "not stable") {
		t.Errorf("err = %q, want a 'not stable' refusal", err)
	}
}
