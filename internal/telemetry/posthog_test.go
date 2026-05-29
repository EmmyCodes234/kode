package telemetry

import (
	"encoding/hex"
	"testing"
)

func TestPostHog_AnonymousID(t *testing.T) {
	id := getAnonymousID()
	if len(id) != 64 {
		t.Fatalf("expected 64 character SHA-256 hex string, got length %d", len(id))
	}
	_, err := hex.DecodeString(id)
	if err != nil {
		t.Fatalf("anonymous ID is not valid hex: %v", err)
	}
}

func TestPostHog_ClientInit(t *testing.T) {
	dir := t.TempDir()
	client := NewPostHogClient(dir)

	if client.APIKey == "" {
		t.Fatalf("expected default API key to be set")
	}
	if client.Host != DefaultPostHogHost {
		t.Fatalf("expected default host %s, got %s", DefaultPostHogHost, client.Host)
	}
	if !client.Enabled {
		t.Fatalf("expected telemetry to be enabled by default")
	}
}
