package chat

import (
	"testing"
	"time"
)

func TestConfigBuilder_Success(t *testing.T) {
	cfg, err := NewConfigBuilder().
		WithAPIKey("test-api-key").
		WithEndpoint("wss://example.com").
		WithDialTimeout(3 * time.Second).
		Build()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.APIKey != "test-api-key" {
		t.Fatalf("unexpected APIKey: %s", cfg.APIKey)
	}
	if cfg.Endpoint != "wss://example.com" {
		t.Fatalf("unexpected Endpoint: %s", cfg.Endpoint)
	}
	if cfg.DialTimeout != 3*time.Second {
		t.Fatalf("unexpected DialTimeout: %v", cfg.DialTimeout)
	}
}

func TestConfigBuilder_MissingAPIKey(t *testing.T) {
	_, err := NewConfigBuilder().
		WithEndpoint("wss://example.com").
		Build()

	if err == nil {
		t.Fatal("expected error when APIKey is missing")
	}
}

func TestConfigBuilder_MissingEndpoint(t *testing.T) {
	_, err := NewConfigBuilder().
		WithAPIKey("test-api-key").
		Build()

	if err == nil {
		t.Fatal("expected error when Endpoint is missing")
	}
}

func TestConfigBuilder_DefaultDialTimeout(t *testing.T) {
	cfg, err := NewConfigBuilder().
		WithAPIKey("test-api-key").
		WithEndpoint("wss://example.com").
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.DialTimeout == 0 {
		t.Fatal("expected default DialTimeout to be set")
	}
}
