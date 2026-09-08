package mcptools

import (
	"context"
	"fmt"
	"net/url"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerProductsEngagementsTestsTools(server *mcp.Server, deps *Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "dojo_list_products",
		Description: "List DefectDojo products, optionally filtered by name.",
	}, listProductsHandler(deps))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "dojo_get_product",
		Description: "Get full detail for one DefectDojo product by id.",
	}, getProductHandler(deps))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "dojo_list_engagements",
		Description: "List DefectDojo engagements, optionally filtered by product id, name, or status.",
	}, listEngagementsHandler(deps))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "dojo_get_engagement",
		Description: "Get full detail for one DefectDojo engagement by id.",
	}, getEngagementHandler(deps))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "dojo_list_tests",
		Description: "List DefectDojo tests, optionally filtered by engagement id or scan type.",
	}, listTestsHandler(deps))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "dojo_get_test",
		Description: "Get full detail for one DefectDojo test by id.",
	}, getTestHandler(deps))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "dojo_close_engagement",
		Description: "Close a DefectDojo engagement.",
	}, closeEngagementHandler(deps))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "dojo_reopen_engagement",
		Description: "Reopen a closed DefectDojo engagement.",
	}, reopenEngagementHandler(deps))
}

type listPage struct {
	Count   int              `json:"count" jsonschema:"total matching items, across all pages"`
	HasMore bool             `json:"has_more"`
	Offset  int              `json:"offset"`
	Results []map[string]any `json:"results"`
}

type IDInput struct {
	ID int `json:"id" jsonschema:"the item id"`
}

// --- Products ---

type ListProductsInput struct {
	Name   string `json:"name,omitempty" jsonschema:"filter by exact product name"`
	Tags   string `json:"tags,omitempty" jsonschema:"filter by tag (comma-separated for multiple)"`
	Limit  int    `json:"limit,omitempty" jsonschema:"max results to return, capped at 100 (default 25)"`
	Offset int    `json:"offset,omitempty" jsonschema:"pagination offset"`
}

func listProductsHandler(deps *Deps) mcp.ToolHandlerFor[ListProductsInput, listPage] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in ListProductsInput) (*mcp.CallToolResult, listPage, error) {
		auth, err := authHeader(req, deps)
		if err != nil {
			return nil, listPage{}, err
		}
		q := url.Values{}
		setStringParam(q, "name", in.Name)
		setStringParam(q, "tags", in.Tags)
		limit := normalizeLimit(in.Limit)
		q.Set("limit", fmt.Sprint(limit))
		if in.Offset > 0 {
			q.Set("offset", fmt.Sprint(in.Offset))
		}

		var page dojoPage
		if err := deps.Client.Get(ctx, auth, "/api/v2/products/", q, &page); err != nil {
			return nil, listPage{}, err
		}
		out := listPage{Count: page.Count, HasMore: page.Next != "", Offset: in.Offset}
		for _, item := range page.Results {
			out.Results = append(out.Results, summarizeProduct(item))
		}
		return nil, out, nil
	}
}

func getProductHandler(deps *Deps) mcp.ToolHandlerFor[IDInput, map[string]any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in IDInput) (*mcp.CallToolResult, map[string]any, error) {
		auth, err := authHeader(req, deps)
		if err != nil {
			return nil, nil, err
		}
		var full map[string]any
		path := fmt.Sprintf("/api/v2/products/%d/", in.ID)
		if err := deps.Client.Get(ctx, auth, path, nil, &full); err != nil {
			return nil, nil, err
		}
		return nil, full, nil
	}
}

// --- Engagements ---

type ListEngagementsInput struct {
	Product int    `json:"product,omitempty" jsonschema:"filter by product id"`
	Name    string `json:"name,omitempty" jsonschema:"filter by exact engagement name"`
	Status  string `json:"status,omitempty" jsonschema:"filter by status, e.g. In Progress, Completed"`
	Active  *bool  `json:"active,omitempty" jsonschema:"filter by active status"`
	Limit   int    `json:"limit,omitempty" jsonschema:"max results to return, capped at 100 (default 25)"`
	Offset  int    `json:"offset,omitempty" jsonschema:"pagination offset"`
}

func listEngagementsHandler(deps *Deps) mcp.ToolHandlerFor[ListEngagementsInput, listPage] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in ListEngagementsInput) (*mcp.CallToolResult, listPage, error) {
		auth, err := authHeader(req, deps)
		if err != nil {
			return nil, listPage{}, err
		}
		q := url.Values{}
		setIntParam(q, "product", in.Product)
		setStringParam(q, "name", in.Name)
		setStringParam(q, "status", in.Status)
		setBoolParam(q, "active", in.Active)
		limit := normalizeLimit(in.Limit)
		q.Set("limit", fmt.Sprint(limit))
		if in.Offset > 0 {
			q.Set("offset", fmt.Sprint(in.Offset))
		}

		var page dojoPage
		if err := deps.Client.Get(ctx, auth, "/api/v2/engagements/", q, &page); err != nil {
			return nil, listPage{}, err
		}
		out := listPage{Count: page.Count, HasMore: page.Next != "", Offset: in.Offset}
		for _, item := range page.Results {
			out.Results = append(out.Results, summarizeEngagement(item))
		}
		return nil, out, nil
	}
}

func getEngagementHandler(deps *Deps) mcp.ToolHandlerFor[IDInput, map[string]any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in IDInput) (*mcp.CallToolResult, map[string]any, error) {
		auth, err := authHeader(req, deps)
		if err != nil {
			return nil, nil, err
		}
		var full map[string]any
		path := fmt.Sprintf("/api/v2/engagements/%d/", in.ID)
		if err := deps.Client.Get(ctx, auth, path, nil, &full); err != nil {
			return nil, nil, err
		}
		return nil, full, nil
	}
}

// --- Tests ---

type ListTestsInput struct {
	Engagement int    `json:"engagement,omitempty" jsonschema:"filter by engagement id"`
	ScanType   string `json:"scan_type,omitempty" jsonschema:"filter by scan type, e.g. 'Burp Scan', 'Checkmarx Scan'"`
	Limit      int    `json:"limit,omitempty" jsonschema:"max results to return, capped at 100 (default 25)"`
	Offset     int    `json:"offset,omitempty" jsonschema:"pagination offset"`
}

func listTestsHandler(deps *Deps) mcp.ToolHandlerFor[ListTestsInput, listPage] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in ListTestsInput) (*mcp.CallToolResult, listPage, error) {
		auth, err := authHeader(req, deps)
		if err != nil {
			return nil, listPage{}, err
		}
		q := url.Values{}
		setIntParam(q, "engagement", in.Engagement)
		setStringParam(q, "scan_type", in.ScanType)
		limit := normalizeLimit(in.Limit)
		q.Set("limit", fmt.Sprint(limit))
		if in.Offset > 0 {
			q.Set("offset", fmt.Sprint(in.Offset))
		}

		var page dojoPage
		if err := deps.Client.Get(ctx, auth, "/api/v2/tests/", q, &page); err != nil {
			return nil, listPage{}, err
		}
		out := listPage{Count: page.Count, HasMore: page.Next != "", Offset: in.Offset}
		for _, item := range page.Results {
			out.Results = append(out.Results, summarizeTest(item))
		}
		return nil, out, nil
	}
}

func getTestHandler(deps *Deps) mcp.ToolHandlerFor[IDInput, map[string]any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in IDInput) (*mcp.CallToolResult, map[string]any, error) {
		auth, err := authHeader(req, deps)
		if err != nil {
			return nil, nil, err
		}
		var full map[string]any
		path := fmt.Sprintf("/api/v2/tests/%d/", in.ID)
		if err := deps.Client.Get(ctx, auth, path, nil, &full); err != nil {
			return nil, nil, err
		}
		return nil, full, nil
	}
}

// EngagementActionOutput: close/reopen return 200 with no documented body,
// so there's nothing to echo back beyond confirming it worked.
type EngagementActionOutput struct {
	OK bool `json:"ok"`
}

func closeEngagementHandler(deps *Deps) mcp.ToolHandlerFor[IDInput, EngagementActionOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in IDInput) (*mcp.CallToolResult, EngagementActionOutput, error) {
		auth, err := authHeader(req, deps)
		if err != nil {
			return nil, EngagementActionOutput{}, err
		}
		path := fmt.Sprintf("/api/v2/engagements/%d/close/", in.ID)
		if err := deps.Client.Do(ctx, "POST", auth, path, nil, nil, nil); err != nil {
			return nil, EngagementActionOutput{}, err
		}
		return nil, EngagementActionOutput{OK: true}, nil
	}
}

func reopenEngagementHandler(deps *Deps) mcp.ToolHandlerFor[IDInput, EngagementActionOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in IDInput) (*mcp.CallToolResult, EngagementActionOutput, error) {
		auth, err := authHeader(req, deps)
		if err != nil {
			return nil, EngagementActionOutput{}, err
		}
		path := fmt.Sprintf("/api/v2/engagements/%d/reopen/", in.ID)
		if err := deps.Client.Do(ctx, "POST", auth, path, nil, nil, nil); err != nil {
			return nil, EngagementActionOutput{}, err
		}
		return nil, EngagementActionOutput{OK: true}, nil
	}
}
