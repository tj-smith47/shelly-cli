package network

import (
	"context"
	"testing"
	"time"
)

// With an auto-selected port the URL is only known once the login has started
// waiting, so the option has to reach the SDK.
func TestStartBrowserLogin_ReportsAuthorizeURL(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	urlCh := make(chan string, 1)
	done := make(chan error, 1)
	go func() {
		_, err := StartBrowserLogin(ctx, &BrowserLoginOptions{
			OnAuthorizeURL: func(authorizeURL string) { urlCh <- authorizeURL },
		})
		done <- err
	}()

	select {
	case authorizeURL := <-urlCh:
		if authorizeURL == "" {
			t.Error("OnAuthorizeURL received an empty URL")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("OnAuthorizeURL was not called while the login was waiting")
	}

	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Error("StartBrowserLogin() error = nil after cancel")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("StartBrowserLogin did not return after cancel")
	}
}
