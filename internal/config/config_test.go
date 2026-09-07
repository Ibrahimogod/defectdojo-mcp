package config

import (
	"testing"
	"time"
)

func clearEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{
		envBaseURL, envTransport, envListenAddr, envDestructive, envAPIToken,
		envReqTimeout, envRateLimit, envLogLevel,
	} {
		t.Setenv(k, "")
	}
}

func TestLoad_Defaults(t *testing.T) {
	clearEnv(t)
	t.Setenv(envBaseURL, "http://dojo.example.com")
	t.Setenv(envAPIToken, "a-token")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Transport != TransportStdio {
		t.Errorf("Transport = %q, want %q", cfg.Transport, TransportStdio)
	}
	if cfg.ListenAddr != defaultListenAddr {
		t.Errorf("ListenAddr = %q, want %q", cfg.ListenAddr, defaultListenAddr)
	}
	if cfg.RequestTimeout != defaultTimeout {
		t.Errorf("RequestTimeout = %v, want %v", cfg.RequestTimeout, defaultTimeout)
	}
	if cfg.RateLimitRPS != defaultRateLimit {
		t.Errorf("RateLimitRPS = %v, want %v", cfg.RateLimitRPS, defaultRateLimit)
	}
	if cfg.LogLevel != defaultLogLevel {
		t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, defaultLogLevel)
	}
	if cfg.EnableDestructive {
		t.Errorf("EnableDestructive = true, want false")
	}
}

func TestLoad_MissingBaseURL(t *testing.T) {
	clearEnv(t)
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want error for missing base URL")
	}
}

func TestLoad_InvalidBaseURL(t *testing.T) {
	clearEnv(t)
	t.Setenv(envBaseURL, "not-a-url")
	t.Setenv(envAPIToken, "a-token")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want error for invalid base URL")
	}
}

func TestLoad_StdioRequiresAPIToken(t *testing.T) {
	clearEnv(t)
	t.Setenv(envBaseURL, "https://dojo.example.com")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want error for stdio transport with no API token")
	}
}

func TestLoad_HTTPDoesNotRequireAPIToken(t *testing.T) {
	clearEnv(t)
	t.Setenv(envBaseURL, "https://dojo.example.com")
	t.Setenv(envTransport, "http")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil (API token optional for http transport)", err)
	}
	if cfg.Transport != TransportHTTP {
		t.Errorf("Transport = %q, want %q", cfg.Transport, TransportHTTP)
	}
}

func TestLoad_InvalidTransport(t *testing.T) {
	clearEnv(t)
	t.Setenv(envBaseURL, "https://dojo.example.com")
	t.Setenv(envTransport, "carrier-pigeon")
	t.Setenv(envAPIToken, "a-token")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want error for invalid transport")
	}
}

func TestLoad_Overrides(t *testing.T) {
	clearEnv(t)
	t.Setenv(envBaseURL, "https://dojo.example.com")
	t.Setenv(envTransport, "HTTP")
	t.Setenv(envListenAddr, ":9090")
	t.Setenv(envDestructive, "true")
	t.Setenv(envAPIToken, "a-token")
	t.Setenv(envReqTimeout, "10s")
	t.Setenv(envRateLimit, "2.5")
	t.Setenv(envLogLevel, "DEBUG")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Transport != TransportHTTP {
		t.Errorf("Transport = %q, want %q (lowercased)", cfg.Transport, TransportHTTP)
	}
	if cfg.ListenAddr != ":9090" {
		t.Errorf("ListenAddr = %q, want :9090", cfg.ListenAddr)
	}
	if !cfg.EnableDestructive {
		t.Errorf("EnableDestructive = false, want true")
	}
	if cfg.APIToken != "a-token" {
		t.Errorf("APIToken = %q, want a-token", cfg.APIToken)
	}
	if cfg.RequestTimeout != 10*time.Second {
		t.Errorf("RequestTimeout = %v, want 10s", cfg.RequestTimeout)
	}
	if cfg.RateLimitRPS != 2.5 {
		t.Errorf("RateLimitRPS = %v, want 2.5", cfg.RateLimitRPS)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, want debug (lowercased)", cfg.LogLevel)
	}
}

func TestLoad_InvalidTimeout(t *testing.T) {
	clearEnv(t)
	t.Setenv(envBaseURL, "https://dojo.example.com")
	t.Setenv(envAPIToken, "a-token")
	t.Setenv(envReqTimeout, "not-a-duration")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want error for invalid timeout")
	}
}

func TestLoad_InvalidRateLimit(t *testing.T) {
	clearEnv(t)
	t.Setenv(envBaseURL, "https://dojo.example.com")
	t.Setenv(envAPIToken, "a-token")
	t.Setenv(envRateLimit, "0")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want error for non-positive rate limit")
	}
}

func TestLoad_InvalidLogLevel(t *testing.T) {
	clearEnv(t)
	t.Setenv(envBaseURL, "https://dojo.example.com")
	t.Setenv(envAPIToken, "a-token")
	t.Setenv(envLogLevel, "verbose")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want error for invalid log level")
	}
}
