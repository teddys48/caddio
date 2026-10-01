package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"caddy-manager/internal/caddy"
)

func TestCaddyVersion_FromServerHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/config/" {
			w.Header().Set("Server", "Caddy/v2.8.4")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := caddy.NewClient(server.URL)
	ver, err := client.GetVersion(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ver != "v2.8.4" {
		t.Errorf("expected version 'v2.8.4', got '%s'", ver)
	}
}

func TestCaddyVersion_FromMetricsEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/config/" {
			// Header does not contain version
			w.Header().Set("Server", "Caddy")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
			return
		}
		if r.URL.Path == "/metrics" {
			w.WriteHeader(http.StatusOK)
			metricsBody := `
# HELP caddy_build_info A metric with a constant '1' value labeled by version, goversion.
# TYPE caddy_build_info gauge
caddy_build_info{goversion="go1.22.2",version="2.9.1"} 1
`
			_, _ = w.Write([]byte(metricsBody))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := caddy.NewClient(server.URL)
	ver, err := client.GetVersion(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ver != "v2.9.1" {
		t.Errorf("expected version 'v2.9.1' (normalized with 'v'), got '%s'", ver)
	}
}

func TestCaddyVersion_FromCustomHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/config/" {
			w.Header().Set("Caddy-Version", "v2.8.0")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := caddy.NewClient(server.URL)
	ver, err := client.GetVersion(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ver != "v2.8.0" {
		t.Errorf("expected version 'v2.8.0', got '%s'", ver)
	}
}
