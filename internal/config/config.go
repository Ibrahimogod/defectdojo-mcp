// Package config loads server configuration from environment variables.
package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	DojoBaseURL       string
	ListenAddr        string
	EnableDestructive bool
	SharedToken       string
	RequestTimeout    time.Duration
	RateLimitRPS      float64
	LogLevel          string
}

const (
	envBaseURL     = "DOJO_BASE_URL"
	envListenAddr  = "DOJO_MCP_LISTEN_ADDR"
	envDestructive = "DOJO_MCP_ENABLE_DESTRUCTIVE"
	envSharedToken = "DOJO_MCP_SHARED_TOKEN"
	envReqTimeout  = "DOJO_MCP_REQUEST_TIMEOUT"
	envRateLimit   = "DOJO_MCP_RATE_LIMIT_RPS"
	envLogLevel    = "DOJO_MCP_LOG_LEVEL"

	defaultListenAddr = ":8080"
	defaultTimeout    = 30 * time.Second
	defaultRateLimit  = 5.0
	defaultLogLevel   = "info"
)

var validLogLevels = map[string]bool{
	"debug": true,
	"info":  true,
	"warn":  true,
	"error": true,
}

func Load() (Config, error) {
	cfg := Config{
		DojoBaseURL:    os.Getenv(envBaseURL),
		ListenAddr:     getOr(envListenAddr, defaultListenAddr),
		SharedToken:    os.Getenv(envSharedToken),
		RequestTimeout: defaultTimeout,
		RateLimitRPS:   defaultRateLimit,
		LogLevel:       strings.ToLower(getOr(envLogLevel, defaultLogLevel)),
	}

	if v := os.Getenv(envDestructive); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return Config{}, fmt.Errorf("%s: invalid bool %q: %w", envDestructive, v, err)
		}
		cfg.EnableDestructive = b
	}

	if v := os.Getenv(envReqTimeout); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return Config{}, fmt.Errorf("%s: invalid duration %q: %w", envReqTimeout, v, err)
		}
		cfg.RequestTimeout = d
	}

	if v := os.Getenv(envRateLimit); v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return Config{}, fmt.Errorf("%s: invalid float %q: %w", envRateLimit, v, err)
		}
		cfg.RateLimitRPS = f
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) validate() error {
	if c.DojoBaseURL == "" {
		return fmt.Errorf("%s is required", envBaseURL)
	}
	u, err := url.Parse(c.DojoBaseURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("%s: invalid URL %q", envBaseURL, c.DojoBaseURL)
	}

	if c.ListenAddr == "" {
		return fmt.Errorf("%s must not be empty", envListenAddr)
	}

	if c.RequestTimeout <= 0 {
		return fmt.Errorf("%s must be positive", envReqTimeout)
	}

	if c.RateLimitRPS <= 0 {
		return fmt.Errorf("%s must be positive", envRateLimit)
	}

	if !validLogLevels[c.LogLevel] {
		return fmt.Errorf("%s: invalid level %q (want debug|info|warn|error)", envLogLevel, c.LogLevel)
	}

	return nil
}

func getOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
