// Package mcpserver wires the MCP protocol itself (stdio and Streamable
// HTTP transports, both backed by the same tool-handling code) plus the
// health/readiness endpoints HTTP mode exposes alongside it.
package mcpserver

import (
	"context"
	"net/http"
	"time"

	"github.com/ibrahimogod/defectdojo-mcp/internal/dojoclient"
	"github.com/ibrahimogod/defectdojo-mcp/internal/mcptools"
	"github.com/ibrahimogod/defectdojo-mcp/internal/registry"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// New builds the *mcp.Server with every curated and discovery tool
// registered. fallbackAuth is the full Authorization header value to use
// when a call carries none of its own ("" if none configured).
// enableDestructive gates DELETE through dojo_call_operation.
func New(client *dojoclient.Client, fallbackAuth string, enableDestructive bool) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "defectdojo-mcp"}, nil)
	mcptools.RegisterAll(server, &mcptools.Deps{
		Client:            client,
		Registry:          registry.All(),
		FallbackAuth:      fallbackAuth,
		EnableDestructive: enableDestructive,
	})
	return server
}

// RunStdio runs server over stdio until the client disconnects or ctx is
// cancelled.
func RunStdio(ctx context.Context, server *mcp.Server) error {
	return server.Run(ctx, &mcp.StdioTransport{})
}

// HTTPServer exposes /healthz, /readyz, and the MCP server itself at /mcp.
type HTTPServer struct {
	httpServer *http.Server
}

func NewHTTPServer(addr string, server *mcp.Server, dojoBaseURL string, readyTimeout time.Duration) *HTTPServer {
	mux := http.NewServeMux()
	registerRoutes(mux)
	mux.HandleFunc("/readyz", handleReadyz(dojoBaseURL, readyTimeout))
	mux.Handle("/mcp", mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, nil))

	return &HTTPServer{
		httpServer: &http.Server{Addr: addr, Handler: mux},
	}
}

func registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", handleHealthz)
}

func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// handleReadyz confirms DefectDojo is reachable, not that the caller is
// authenticated: any HTTP response (even 401/403/404) counts as ready,
// since readiness has no caller-specific credential to check in HTTP mode.
// Only a connection-level failure counts as not ready.
func handleReadyz(dojoBaseURL string, timeout time.Duration) http.HandlerFunc {
	client := &http.Client{Timeout: timeout}
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, dojoBaseURL+"/api/v2/", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "defectdojo unreachable: "+err.Error(), http.StatusServiceUnavailable)
			return
		}
		defer resp.Body.Close()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}
}

func (s *HTTPServer) Addr() string { return s.httpServer.Addr }

func (s *HTTPServer) ListenAndServe() error { return s.httpServer.ListenAndServe() }

func (s *HTTPServer) Shutdown(ctx context.Context) error { return s.httpServer.Shutdown(ctx) }
