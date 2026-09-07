package config

import (
	"testing"
	"time"
)

func clearEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{
		envBaseURL, envListenAddr, envDestructive, envSharedToken,
		envReqTimeout, envRateLimit, envLogLevel,
	} {
		t.Setenv(k, "")
	}
}

func TestLoad_Defaults(t *testing.T) {
	clearEnv(t)
	t.Setenv(envBaseURL, "http://dojo.example.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
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
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want error for invalid base URL")
	}
}

func TestLoad_Overrides(t *testing.T) {
	clearEnv(t)
	t.Setenv(envBaseURL, "https://dojo.example.com")
	t.Setenv(envListenAddr, ":9090")
	t.Setenv(envDestructive, "true")
	t.Setenv(envSharedToken, "shared-secret")
	t.Setenv(envReqTimeout, "10s")
	t.Setenv(envRateLimit, "2.5")
	t.Setenv(envLogLevel, "DEBUG")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.ListenAddr != ":9090" {
		t.Errorf("ListenAddr = %q, want :9090", cfg.ListenAddr)
	}
	if !cfg.EnableDestructive {
		t.Errorf("EnableDestructive = false, want true")
	}
	if cfg.SharedToken != "shared-secret" {
		t.Errorf("SharedToken = %q, want shared-secret", cfg.SharedToken)
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
	t.Setenv(envReqTimeout, "not-a-duration")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want error for invalid timeout")
	}
}

func TestLoad_InvalidRateLimit(t *testing.T) {
	clearEnv(t)
	t.Setenv(envBaseURL, "https://dojo.example.com")
	t.Setenv(envRateLimit, "0")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want error for non-positive rate limit")
	}
}

func TestLoad_InvalidLogLevel(t *testing.T) {
	clearEnv(t)
	t.Setenv(envBaseURL, "https://dojo.example.com")
	t.Setenv(envLogLevel, "verbose")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want error for invalid log level")
	}
}
