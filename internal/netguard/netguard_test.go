package netguard

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAllowed(t *testing.T) {
	t.Parallel()
	for addr, want := range map[string]bool{
		"127.0.0.1:80": true, "[::1]:80": true, "localhost:8080": true, "127.0.0.1": true,
		"192.168.1.1:80": false, "10.0.0.1": false, "192.0.2.1:80": false, "0.0.0.0:80": false,
		"example.com:443": false, "kitchen:80": false, "": false,
	} {
		if got := Allowed(addr); got != want {
			t.Errorf("Allowed(%q) = %v, want %v", addr, got, want)
		}
	}
}

func TestGuard_RefusesBeforeDialing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	if _, err := DialContext(ctx, "tcp", "192.0.2.1:80"); !errors.Is(err, ErrBlocked) {
		t.Errorf("DialContext(non-loopback) = %v, want ErrBlocked", err)
	}
	if err := get(http.DefaultClient, "http://192.0.2.1/"); !errors.Is(err, ErrBlocked) {
		t.Errorf("http.DefaultTransport dial = %v, want ErrBlocked", err)
	}
	if _, err := net.DefaultResolver.LookupHost(ctx, "shelly-test.invalid"); err == nil {
		t.Error("hostname lookup succeeded, want it refused")
	}
	if NetDialContext() == nil {
		t.Fatal("NetDialContext() = nil under go test")
	}
	if err := get(guardedClient(), "http://10.0.0.1/"); !errors.Is(err, ErrBlocked) {
		t.Errorf("NetDialContext dial = %v, want ErrBlocked", err)
	}
}

func guardedClient() *http.Client {
	return &http.Client{Transport: &http.Transport{DialContext: NetDialContext()}}
}

func get(c *http.Client, url string) error {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, http.NoBody)
	if err != nil {
		return err
	}
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	return resp.Body.Close()
}

func TestGuard_AllowsLoopback(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	t.Cleanup(srv.Close)
	if err := get(guardedClient(), srv.URL); err != nil {
		t.Fatalf("loopback request = %v", err)
	}
	if addrs, err := net.DefaultResolver.LookupHost(context.Background(), "localhost"); err != nil || len(addrs) == 0 {
		t.Errorf("localhost lookup = %v %v, want the hosts file answer", addrs, err)
	}
}

func TestRefuseAddr_UnderTest(t *testing.T) {
	t.Parallel()
	for addr, blocked := range map[string]bool{"192.168.1.100": true, "10.0.0.5:80": true, "example.com": true,
		"127.0.0.1:8080": false, "[::1]:80": false, "localhost": false} {
		if err := RefuseAddr(addr); errors.Is(err, ErrBlocked) != blocked {
			t.Errorf("RefuseAddr(%q) = %v, want blocked %v", addr, err, blocked)
		}
	}
}

func TestRefuse_UnderTest(t *testing.T) {
	t.Parallel()
	if err := Refuse("mDNS multicast"); !errors.Is(err, ErrBlocked) {
		t.Errorf("Refuse() = %v, want ErrBlocked", err)
	}
}
