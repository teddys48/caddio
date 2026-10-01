package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppEnv          string
	Port            int
	DatabasePath    string
	CaddyAdminURL   string
	CaddyBinaryPath string
	SessionSecret   string
}

func LoadConfig() (*Config, error) {
	appEnv := getEnv("APP_ENV", "development")
	portStr := getEnv("PORT", "8080")
	dbPath := getEnv("DATABASE_PATH", "./caddy-manager.db")
	caddyURL := getEnv("CADDY_ADMIN_URL", "http://localhost:2019")
	caddyBinary := getEnv("CADDY_BINARY_PATH", "/usr/bin/caddy")
	sessionSecret := os.Getenv("SESSION_SECRET")

	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		return nil, fmt.Errorf("invalid PORT configuration '%s': must be integer between 1 and 65535", portStr)
	}

	cfg := &Config{
		AppEnv:          strings.ToLower(appEnv),
		Port:            port,
		DatabasePath:    dbPath,
		CaddyAdminURL:   caddyURL,
		CaddyBinaryPath: caddyBinary,
		SessionSecret:   sessionSecret,
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	// Validate Caddy Admin URL
	if c.CaddyAdminURL != "" {
		parsedURL, err := url.Parse(c.CaddyAdminURL)
		if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
			return fmt.Errorf("invalid CADDY_ADMIN_URL '%s': must be a valid http or https URL", c.CaddyAdminURL)
		}
	}

	// Validate Session Secret according to environment
	if c.IsProduction() {
		if strings.TrimSpace(c.SessionSecret) == "" {
			return fmt.Errorf("startup configuration error: SESSION_SECRET environment variable is required in production mode (APP_ENV=production)")
		}
		if len(c.SessionSecret) < 32 {
			return fmt.Errorf("startup configuration error: SESSION_SECRET must be at least 32 characters long in production")
		}
	} else {
		if strings.TrimSpace(c.SessionSecret) == "" {
			// Auto-generate random secret for non-production
			randomBytes := make([]byte, 32)
			_, _ = rand.Read(randomBytes)
			c.SessionSecret = hex.EncodeToString(randomBytes)
		}
	}

	return nil
}

func (c *Config) IsProduction() bool {
	return c.AppEnv == "production" || c.AppEnv == "prod"
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}
