// Package mcptools holds the curated MCP tool groups (findings,
// products/engagements/tests, risk acceptances, discovery) plus the shared
// summarize helper.
package mcptools

import (
	"context"
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
	// EnableDestructive gates DELETE through dojo_call_operation. No
	// curated tool ever issues a DELETE regardless of this flag.
	EnableDestructive bool
}

// RegisterAll wires every tool onto server.
func RegisterAll(server *mcp.Server, deps *Deps) {
	registerFindingsTools(server, deps)
	registerFindingsWriteTools(server, deps)
	registerProductsEngagementsTestsTools(server, deps)
	registerRiskAcceptanceTools(server, deps)
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

type userProfile struct {
	User struct {
		ID int `json:"id"`
	} `json:"user"`
}

// resolveOwnerID looks up the DefectDojo user ID for whoever auth belongs
// to, for operations like risk acceptance that require a numeric owner —
// DefectDojo has no "who am I" endpoint by that name, but /user_profile/
// always returns the authenticated caller's own profile.
func resolveOwnerID(ctx context.Context, deps *Deps, auth string) (int, error) {
	var profile userProfile
	if err := deps.Client.Get(ctx, auth, "/api/v2/user_profile/", nil, &profile); err != nil {
		return 0, fmt.Errorf("resolving caller identity: %w", err)
	}
	return profile.User.ID, nil
}
