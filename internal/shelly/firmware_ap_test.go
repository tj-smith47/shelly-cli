package shelly

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

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

func TestServeFirmwareFile_ServesImage(t *testing.T) {
	t.Parallel()
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

	// No firmware update requested -> empty path, no error, no download attempted.
	if path, err := prefetchAPFirmware(context.Background(), 1, bkp,
		backup.RestoreOptions{UpdateFirmware: false}); err != nil || path != "" {
		t.Errorf("no-update: path=%q err=%v, want empty/nil", path, err)
	}

	// Update requested on a Gen2 target -> the at-AP Gen1 OTA does not apply.
	if path, err := prefetchAPFirmware(context.Background(), 2, bkp,
		backup.RestoreOptions{UpdateFirmware: true}); err != nil || path != "" {
		t.Errorf("gen2: path=%q err=%v, want empty/nil", path, err)
	}
}

func TestPrefetchAPFirmware_NoModelErrors(t *testing.T) {
	t.Parallel()
	// Update requested, Gen1, but no model to derive a URL from and no override —
	// must fail loudly rather than download from a malformed URL.
	bkp := &backup.DeviceBackup{Backup: &shellybackup.Backup{}}
	_, err := prefetchAPFirmware(context.Background(), 1, bkp,
		backup.RestoreOptions{UpdateFirmware: true})
	if err == nil || !strings.Contains(err.Error(), "cannot update firmware") {
		t.Fatalf("expected a no-model firmware-URL error, got: %v", err)
	}
}

func TestPrefetchAPFirmware_ExplicitURLDownloads(t *testing.T) {
	t.Parallel()
	srv := fakeFirmwareServer(t)
	bkp := &backup.DeviceBackup{Backup: &shellybackup.Backup{}}

	path, err := prefetchAPFirmware(context.Background(), 1, bkp,
		backup.RestoreOptions{UpdateFirmware: true, FirmwareURL: srv.URL + "/fw.zip"})
	if err != nil {
		t.Fatalf("prefetchAPFirmware with explicit URL: %v", err)
	}
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
