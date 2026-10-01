package tests

import (
	"testing"

	"caddy-manager/internal/config"
)

func TestConfigValidation_Development(t *testing.T) {
	cfg := &config.Config{
		AppEnv:        "development",
		Port:          8080,
		DatabasePath:  ":memory:",
		CaddyAdminURL: "http://localhost:2019",
		SessionSecret: "", // Should auto-generate
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid dev config, got error: %v", err)
	}

	if cfg.SessionSecret == "" {
		t.Errorf("expected session secret to be auto-generated in dev mode")
	}
}

func TestConfigValidation_ProductionMissingSecret(t *testing.T) {
	cfg := &config.Config{
		AppEnv:        "production",
		Port:          8080,
		DatabasePath:  "/data/caddy-manager.db",
		CaddyAdminURL: "http://caddy:2019",
		SessionSecret: "",
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatalf("expected error when SESSION_SECRET is missing in production mode, got nil")
	}
}

func TestConfigValidation_ProductionValidSecret(t *testing.T) {
	cfg := &config.Config{
		AppEnv:        "production",
		Port:          8080,
		DatabasePath:  "/data/caddy-manager.db",
		CaddyAdminURL: "http://caddy:2019",
		SessionSecret: "a_very_secure_production_secret_32bytes_min",
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid production config, got error: %v", err)
	}
}

func TestConfigValidation_InvalidCaddyURL(t *testing.T) {
	cfg := &config.Config{
		AppEnv:        "development",
		Port:          8080,
		DatabasePath:  ":memory:",
		CaddyAdminURL: "ftp://invalid-scheme",
		SessionSecret: "dev_secret",
	}

	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected error for invalid Caddy URL scheme, got nil")
	}
}
