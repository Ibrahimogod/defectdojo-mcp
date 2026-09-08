package mcptools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerRiskAcceptanceTools(server *mcp.Server, deps *Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "dojo_accept_finding_risk",
		Description: "Create a risk acceptance covering one or more DefectDojo findings, recording who accepted it and why. Owner is resolved from your own DefectDojo token.",
	}, acceptFindingRiskHandler(deps))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "dojo_list_risk_acceptances",
		Description: "List DefectDojo risk acceptances, optionally filtered by name.",
	}, listRiskAcceptancesHandler(deps))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "dojo_expire_risk_acceptance",
		Description: "Expire a DefectDojo risk acceptance now, reactivating its findings (unless the acceptance was configured not to).",
	}, expireRiskAcceptanceHandler(deps))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "dojo_reinstate_risk_acceptance",
		Description: "Reinstate an expired DefectDojo risk acceptance.",
	}, reinstateRiskAcceptanceHandler(deps))
}

type AcceptFindingRiskInput struct {
	FindingIDs     []int  `json:"finding_ids" jsonschema:"the finding ids this risk acceptance covers"`
	Name           string `json:"name" jsonschema:"a short descriptive name for this risk acceptance"`
	Justification  string `json:"justification,omitempty" jsonschema:"why the risk is being accepted"`
	ExpirationDate string `json:"expiration_date,omitempty" jsonschema:"when this acceptance expires (RFC3339, e.g. 2027-01-01T00:00:00Z); omit for no expiration"`
}

func acceptFindingRiskHandler(deps *Deps) mcp.ToolHandlerFor[AcceptFindingRiskInput, map[string]any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in AcceptFindingRiskInput) (*mcp.CallToolResult, map[string]any, error) {
		if len(in.FindingIDs) == 0 {
			return nil, nil, fmt.Errorf("finding_ids must not be empty")
		}
		auth, err := authHeader(req, deps)
		if err != nil {
			return nil, nil, err
		}

		ownerID, err := resolveOwnerID(ctx, deps, auth)
		if err != nil {
			return nil, nil, err
		}

		fields := map[string]any{
			"name":              in.Name,
			"owner":             ownerID,
			"accepted_findings": in.FindingIDs,
			"decision":          "A", // Accept
			"recommendation":    "A",
		}
		if in.Justification != "" {
			fields["decision_details"] = in.Justification
		}
		if in.ExpirationDate != "" {
			fields["expiration_date"] = in.ExpirationDate
		}

		body, err := json.Marshal(fields)
		if err != nil {
			return nil, nil, fmt.Errorf("encoding request: %w", err)
		}

		var result map[string]any
		if err := deps.Client.Do(ctx, "POST", auth, "/api/v2/risk_acceptance/", nil, body, &result); err != nil {
			return nil, nil, err
		}
		return nil, result, nil
	}
}

type ListRiskAcceptancesInput struct {
	Name   string `json:"name,omitempty" jsonschema:"filter by exact name"`
	Limit  int    `json:"limit,omitempty" jsonschema:"max results to return, capped at 100 (default 25)"`
	Offset int    `json:"offset,omitempty" jsonschema:"pagination offset"`
}

func listRiskAcceptancesHandler(deps *Deps) mcp.ToolHandlerFor[ListRiskAcceptancesInput, listPage] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in ListRiskAcceptancesInput) (*mcp.CallToolResult, listPage, error) {
		auth, err := authHeader(req, deps)
		if err != nil {
			return nil, listPage{}, err
		}
		q := url.Values{}
		setStringParam(q, "name", in.Name)
		limit := normalizeLimit(in.Limit)
		q.Set("limit", fmt.Sprint(limit))
		if in.Offset > 0 {
			q.Set("offset", fmt.Sprint(in.Offset))
		}

		var page dojoPage
		if err := deps.Client.Get(ctx, auth, "/api/v2/risk_acceptance/", q, &page); err != nil {
			return nil, listPage{}, err
		}
		out := listPage{Count: page.Count, HasMore: page.Next != "", Offset: in.Offset}
		for _, item := range page.Results {
			out.Results = append(out.Results, summarizeRiskAcceptance(item))
		}
		return nil, out, nil
	}
}

type RiskAcceptanceIDInput struct {
	ID     int    `json:"id" jsonschema:"the risk acceptance id"`
	Reason string `json:"reason,omitempty" jsonschema:"an optional reason"`
}

func expireRiskAcceptanceHandler(deps *Deps) mcp.ToolHandlerFor[RiskAcceptanceIDInput, map[string]any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in RiskAcceptanceIDInput) (*mcp.CallToolResult, map[string]any, error) {
		auth, err := authHeader(req, deps)
		if err != nil {
			return nil, nil, err
		}
		var body []byte
		if in.Reason != "" {
			b, err := json.Marshal(map[string]any{"reason": in.Reason})
			if err != nil {
				return nil, nil, fmt.Errorf("encoding request: %w", err)
			}
			body = b
		}
		var result map[string]any
		path := fmt.Sprintf("/api/v2/risk_acceptance/%d/expire/", in.ID)
		if err := deps.Client.Do(ctx, "POST", auth, path, nil, body, &result); err != nil {
			return nil, nil, err
		}
		return nil, result, nil
	}
}

type ReinstateRiskAcceptanceInput struct {
	ID             int    `json:"id" jsonschema:"the risk acceptance id"`
	Reason         string `json:"reason,omitempty" jsonschema:"an optional reason"`
	ExpirationDate string `json:"expiration_date,omitempty" jsonschema:"new expiration date (RFC3339); omit to use the DefectDojo default"`
}

func reinstateRiskAcceptanceHandler(deps *Deps) mcp.ToolHandlerFor[ReinstateRiskAcceptanceInput, map[string]any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in ReinstateRiskAcceptanceInput) (*mcp.CallToolResult, map[string]any, error) {
		auth, err := authHeader(req, deps)
		if err != nil {
			return nil, nil, err
		}
		fields := map[string]any{}
		if in.Reason != "" {
			fields["reason"] = in.Reason
		}
		if in.ExpirationDate != "" {
			fields["expiration_date"] = in.ExpirationDate
		}
		var body []byte
		if len(fields) > 0 {
			b, err := json.Marshal(fields)
			if err != nil {
				return nil, nil, fmt.Errorf("encoding request: %w", err)
			}
			body = b
		}
		var result map[string]any
		path := fmt.Sprintf("/api/v2/risk_acceptance/%d/reinstate/", in.ID)
		if err := deps.Client.Do(ctx, "POST", auth, path, nil, body, &result); err != nil {
			return nil, nil, err
		}
		return nil, result, nil
	}
}
