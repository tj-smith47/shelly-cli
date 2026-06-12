package client

import (
	"context"
	"net/http"
	"testing"
)

func TestWithBindInterface(t *testing.T) {
	t.Parallel()

	t.Run("round-trips a name", func(t *testing.T) {
		t.Parallel()
		ctx := WithBindInterface(context.Background(), "vmbr0")
		if got := bindInterfaceFromContext(ctx); got != "vmbr0" {
			t.Errorf("bindInterfaceFromContext = %q, want vmbr0", got)
		}
	})

	t.Run("empty name is a no-op", func(t *testing.T) {
		t.Parallel()
		base := context.Background()
		ctx := WithBindInterface(base, "")
		if ctx != base {
			t.Error("WithBindInterface(ctx, \"\") must return the context unchanged")
		}
		if got := bindInterfaceFromContext(ctx); got != "" {
			t.Errorf("bindInterfaceFromContext = %q, want empty", got)
		}
	})

	t.Run("absent value reads empty", func(t *testing.T) {
		t.Parallel()
		if got := bindInterfaceFromContext(context.Background()); got != "" {
			t.Errorf("bindInterfaceFromContext = %q, want empty", got)
		}
	})
}

func TestBoundHTTPClient(t *testing.T) {
	t.Parallel()

	t.Run("mirrors default timeout", func(t *testing.T) {
		t.Parallel()
		c := boundHTTPClient("eth0", false)
		if c.Timeout == 0 {
			t.Error("bound client must carry a request timeout, got 0")
		}
		if c.Transport == nil {
			t.Fatal("bound client must have a transport")
		}
	})

	t.Run("https skips TLS verification, http does not", func(t *testing.T) {
		t.Parallel()

		secure := boundHTTPClient("eth0", true)
		tr, ok := secure.Transport.(*http.Transport)
		if !ok {
			t.Fatalf("transport = %T, want *http.Transport", secure.Transport)
		}
		if tr.TLSClientConfig == nil || !tr.TLSClientConfig.InsecureSkipVerify {
			t.Error("https bound client must skip TLS verification (self-signed Shelly certs)")
		}

		plain := boundHTTPClient("eth0", false)
		ptr, ok := plain.Transport.(*http.Transport)
		if !ok {
			t.Fatalf("transport = %T, want *http.Transport", plain.Transport)
		}
		if ptr.TLSClientConfig != nil {
			t.Error("http bound client must not set a TLS config")
		}
	})
}
