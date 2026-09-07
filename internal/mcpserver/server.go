// Package mcpserver wires the HTTP transport this server exposes: health
// endpoints today, the Streamable HTTP MCP transport from Phase 1 onward.
package mcpserver

import (
	"context"
	"log/slog"
	"net/http"
)

type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
}

func New(addr string, logger *slog.Logger) *Server {
	mux := http.NewServeMux()
	registerRoutes(mux)

	return &Server{
		httpServer: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
		logger: logger,
	}
}

func registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", handleHealthz)
}

func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (s *Server) Addr() string {
	return s.httpServer.Addr
}

func (s *Server) ListenAndServe() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
