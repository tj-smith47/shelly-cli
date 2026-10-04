package client

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDeviceWebSocketCallersAuthenticateThroughHelper reads every non-test Go
// file under internal/ and cmd/ and fails when:
//   - transport.WithAuth (an HTTP basic header) is used outside
//     internal/client/gen1.go, the only device generation that accepts basic
//     authentication;
//   - a file opens a device websocket (client.NewDeviceWebSocket) without
//     starting it through client.StartDeviceNotifications, or builds its own
//     request frames there (rpc.NewRequestBuilder, rpc.Request), which a
//     password-protected Gen2+ device answers with 401.
func TestDeviceWebSocketCallersAuthenticateThroughHelper(t *testing.T) {
	t.Parallel()
	root := filepath.Join("..", "..")
	openers := 0
	for _, dir := range []string{"internal", "cmd"} {
		err := filepath.WalkDir(filepath.Join(root, dir), func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			src := string(data)
			rel := filepath.ToSlash(strings.TrimPrefix(path, root+string(filepath.Separator)))

			if strings.Contains(src, "transport.WithAuth(") && rel != "internal/client/gen1.go" {
				t.Errorf("%s: transport.WithAuth sends HTTP basic auth, which Gen2+ devices ignore; use client.Connect (digest) or client.CallDeviceWebSocket", rel)
			}
			if !strings.Contains(src, "client.NewDeviceWebSocket(") {
				return nil
			}
			openers++
			if !strings.Contains(src, "client.StartDeviceNotifications(") {
				t.Errorf("%s: opens a device websocket without client.StartDeviceNotifications, so a password-protected device sends it nothing", rel)
			}
			for _, raw := range []string{"rpc.NewRequestBuilder(", "rpc.Request{"} {
				if strings.Contains(src, raw) {
					t.Errorf("%s: builds its own request frame (%s) next to a device websocket; use client.CallDeviceWebSocket so a 401 is answered with digest auth", rel, raw)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", dir, err)
		}
	}
	if openers == 0 {
		t.Fatal("found no client.NewDeviceWebSocket caller; the scan is reading the wrong tree")
	}
}
