package mcptools

import (
	"context"
	"fmt"
	"net/url"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerFindingsTools(server *mcp.Server, deps *Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "dojo_search_findings",
		Description: "Search and filter DefectDojo findings. Returns a summarized list (not full finding objects) capped at 100 results per call; use dojo_get_finding for full detail on one finding.",
	}, searchFindingsHandler(deps))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "dojo_get_finding",
		Description: "Get full detail for one DefectDojo finding by id.",
	}, getFindingHandler(deps))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "dojo_list_finding_notes",
		Description: "List notes on a DefectDojo finding.",
	}, listFindingNotesHandler(deps))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "dojo_get_finding_metadata",
		Description: "Get custom metadata key/value pairs on a DefectDojo finding.",
	}, getFindingMetadataHandler(deps))
}

type SearchFindingsInput struct {
	Severity         string `json:"severity,omitempty" jsonschema:"filter by severity: Critical, High, Medium, Low, Info"`
	Active           *bool  `json:"active,omitempty" jsonschema:"filter by active status"`
	Verified         *bool  `json:"verified,omitempty" jsonschema:"filter by verified status"`
	FalsePositive    *bool  `json:"false_positive,omitempty" jsonschema:"filter by false-positive status"`
	Duplicate        *bool  `json:"duplicate,omitempty" jsonschema:"filter by duplicate status"`
	OutOfScope       *bool  `json:"out_of_scope,omitempty" jsonschema:"filter by out-of-scope status"`
	RiskAccepted     *bool  `json:"risk_accepted,omitempty" jsonschema:"filter by risk-accepted status"`
	IsMitigated      *bool  `json:"is_mitigated,omitempty" jsonschema:"filter by mitigated status"`
	ProductName      string `json:"product_name,omitempty" jsonschema:"filter by exact product name"`
	Test             int    `json:"test,omitempty" jsonschema:"filter by test id"`
	CWE              int    `json:"cwe,omitempty" jsonschema:"filter by CWE number"`
	Tags             string `json:"tags,omitempty" jsonschema:"filter by tag (comma-separated for multiple)"`
	TitleContains    string `json:"title_contains,omitempty" jsonschema:"filter by a substring of the finding title"`
	DiscoveredAfter  string `json:"discovered_after,omitempty" jsonschema:"filter to findings discovered on or after this date (YYYY-MM-DD)"`
	DiscoveredBefore string `json:"discovered_before,omitempty" jsonschema:"filter to findings discovered on or before this date (YYYY-MM-DD)"`
	Limit            int    `json:"limit,omitempty" jsonschema:"max results to return, capped at 100 (default 25)"`
	Offset           int    `json:"offset,omitempty" jsonschema:"pagination offset"`
	Order            string `json:"order,omitempty" jsonschema:"ordering field name; prefix with - for descending, e.g. -date for newest first"`
}

type SearchFindingsOutput struct {
	Count   int              `json:"count" jsonschema:"total findings matching the filters, across all pages"`
	HasMore bool             `json:"has_more" jsonschema:"whether more results exist beyond this page"`
	Offset  int              `json:"offset"`
	Results []map[string]any `json:"results"`
}

func searchFindingsHandler(deps *Deps) mcp.ToolHandlerFor[SearchFindingsInput, SearchFindingsOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in SearchFindingsInput) (*mcp.CallToolResult, SearchFindingsOutput, error) {
		auth, err := authHeader(req, deps)
		if err != nil {
			return nil, SearchFindingsOutput{}, err
		}

		q := url.Values{}
		setStringParam(q, "severity", in.Severity)
		setBoolParam(q, "active", in.Active)
		setBoolParam(q, "verified", in.Verified)
		setBoolParam(q, "false_p", in.FalsePositive)
		setBoolParam(q, "duplicate", in.Duplicate)
		setBoolParam(q, "out_of_scope", in.OutOfScope)
		setBoolParam(q, "risk_accepted", in.RiskAccepted)
		setBoolParam(q, "is_mitigated", in.IsMitigated)
		setStringParam(q, "product_name", in.ProductName)
		setIntParam(q, "test", in.Test)
		setIntParam(q, "cwe", in.CWE)
		setStringParam(q, "tags", in.Tags)
		setStringParam(q, "title", in.TitleContains)
		setStringParam(q, "discovered_after", in.DiscoveredAfter)
		setStringParam(q, "discovered_before", in.DiscoveredBefore)
		setStringParam(q, "o", in.Order)

		limit := normalizeLimit(in.Limit)
		q.Set("limit", fmt.Sprint(limit))
		if in.Offset > 0 {
			q.Set("offset", fmt.Sprint(in.Offset))
		}

		var page dojoPage
		if err := deps.Client.Get(ctx, auth, "/api/v2/findings/", q, &page); err != nil {
			return nil, SearchFindingsOutput{}, err
		}

		out := SearchFindingsOutput{
			Count:   page.Count,
			HasMore: page.Next != "",
			Offset:  in.Offset,
		}
		for _, item := range page.Results {
			out.Results = append(out.Results, summarizeFinding(item))
		}
		return nil, out, nil
	}
}

type FindingIDInput struct {
	ID int `json:"id" jsonschema:"the finding id"`
}

func getFindingHandler(deps *Deps) mcp.ToolHandlerFor[FindingIDInput, map[string]any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in FindingIDInput) (*mcp.CallToolResult, map[string]any, error) {
		auth, err := authHeader(req, deps)
		if err != nil {
			return nil, nil, err
		}
		var full map[string]any
		path := fmt.Sprintf("/api/v2/findings/%d/", in.ID)
		if err := deps.Client.Get(ctx, auth, path, nil, &full); err != nil {
			return nil, nil, err
		}
		return nil, full, nil
	}
}

type findingNotesResponse struct {
	FindingID int              `json:"finding_id"`
	Notes     []map[string]any `json:"notes"`
}

func listFindingNotesHandler(deps *Deps) mcp.ToolHandlerFor[FindingIDInput, []map[string]any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in FindingIDInput) (*mcp.CallToolResult, []map[string]any, error) {
		auth, err := authHeader(req, deps)
		if err != nil {
			return nil, nil, err
		}
		var resp findingNotesResponse
		path := fmt.Sprintf("/api/v2/findings/%d/notes/", in.ID)
		if err := deps.Client.Get(ctx, auth, path, nil, &resp); err != nil {
			return nil, nil, err
		}
		return nil, resp.Notes, nil
	}
}

func getFindingMetadataHandler(deps *Deps) mcp.ToolHandlerFor[FindingIDInput, []map[string]any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in FindingIDInput) (*mcp.CallToolResult, []map[string]any, error) {
		auth, err := authHeader(req, deps)
		if err != nil {
			return nil, nil, err
		}
		var meta []map[string]any
		path := fmt.Sprintf("/api/v2/findings/%d/metadata/", in.ID)
		if err := deps.Client.Get(ctx, auth, path, nil, &meta); err != nil {
			return nil, nil, err
		}
		return nil, meta, nil
	}
}
