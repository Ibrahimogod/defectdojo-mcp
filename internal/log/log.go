// Package log provides structured logging (slog) and log-line redaction
// for credentials that must never reach a log sink.
package log

import (
	"log/slog"
	"os"
	"regexp"
)

func New(level string) *slog.Logger {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	// stderr, not stdout: stdio-transport MCP uses stdout for the JSON-RPC
	// protocol stream itself, so any log line written there would corrupt it.
	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: lvl})
	return slog.New(handler)
}

// redactPattern matches an auth scheme (Token or Bearer, DefectDojo's and the
// MCP Authorization header's respective schemes) followed by its secret
// value, with or without a leading "Authorization:" prefix.
var redactPattern = regexp.MustCompile(`(?i)((?:Authorization\s*:\s*)?\b(?:Token|Bearer)\s+)[A-Za-z0-9_\-.~+/]+=*`)

// Redact masks any bearer/token credential value found in s, leaving the
// surrounding text and the scheme name intact so redacted logs stay readable.
func Redact(s string) string {
	return redactPattern.ReplaceAllString(s, "${1}REDACTED")
}
