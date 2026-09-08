// Package mcptools holds the curated MCP tool groups (findings,
// products/engagements/tests, discovery) plus the shared summarize helper.
package mcptools

import (
	"fmt"

	"github.com/ibrahimogod/defectdojo-mcp/internal/dojoclient"
	"github.com/ibrahimogod/defectdojo-mcp/internal/registry"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Deps holds what every tool handler needs.
type Deps struct {
	Client   *dojoclient.Client
	Registry []registry.OperationEntry
	// FallbackAuth is the full Authorization header value ("Token ...") to
	// use when a call carries no header of its own. Stdio mode always hits
	// this path, since there's no per-request caller there; HTTP mode only
	// falls back to it when the caller sent no Authorization header.
	FallbackAuth string
}

// RegisterAll wires every Phase 1 tool onto server.
func RegisterAll(server *mcp.Server, deps *Deps) {
	registerFindingsTools(server, deps)
	registerProductsEngagementsTestsTools(server, deps)
	registerDiscoveryTools(server, deps)
}

// authHeader resolves the Authorization header value to use for this call.
func authHeader(req *mcp.CallToolRequest, deps *Deps) (string, error) {
	if req.Extra != nil {
		if h := req.Extra.Header.Get("Authorization"); h != "" {
			return h, nil
		}
	}
	if deps.FallbackAuth != "" {
		return deps.FallbackAuth, nil
	}
	return "", fmt.Errorf("no DefectDojo credential available: the caller sent no Authorization header and no fallback token is configured")
}
