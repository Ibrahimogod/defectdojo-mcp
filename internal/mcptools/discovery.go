package mcptools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/ibrahimogod/defectdojo-mcp/internal/registry"
	"github.com/ibrahimogod/defectdojo-mcp/internal/safety"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// maxCallOperationResponseBytes caps what dojo_call_operation returns
// directly; a response larger than this gets replaced with a note telling
// the caller to narrow the request instead of dumping the full payload
// into the model's context.
const maxCallOperationResponseBytes = 50_000

func registerDiscoveryTools(server *mcp.Server, deps *Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "dojo_list_operations",
		Description: "List DefectDojo API operations not covered by a curated tool, optionally filtered by tag or a substring of the operationId/summary. Use with dojo_describe_operation and dojo_call_operation to reach the full API.",
	}, listOperationsHandler(deps))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "dojo_describe_operation",
		Description: "Get the full parameter, request-body, and response schema for one DefectDojo API operation, from its pinned OpenAPI document.",
	}, describeOperationHandler(deps))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "dojo_call_operation",
		Description: "Call any DefectDojo API operation by operationId, including writes. DELETE is blocked unless the server was started with DOJO_MCP_ENABLE_DESTRUCTIVE=true. Prefer a curated tool when one exists for what you're doing.",
	}, callOperationHandler(deps))
}

type ListOperationsInput struct {
	Tag   string `json:"tag,omitempty" jsonschema:"filter by DefectDojo API tag, e.g. findings, products, engagements"`
	Query string `json:"query,omitempty" jsonschema:"filter by a case-insensitive substring of the operationId or summary"`
}

type OperationSummary struct {
	OperationID string `json:"operation_id"`
	Method      string `json:"method"`
	Path        string `json:"path"`
	Tag         string `json:"tag"`
	Summary     string `json:"summary,omitempty"`
}

type ListOperationsOutput struct {
	Count      int                `json:"count"`
	Operations []OperationSummary `json:"operations"`
}

func listOperationsHandler(deps *Deps) mcp.ToolHandlerFor[ListOperationsInput, ListOperationsOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in ListOperationsInput) (*mcp.CallToolResult, ListOperationsOutput, error) {
		tag := strings.ToLower(in.Tag)
		query := strings.ToLower(in.Query)

		var out ListOperationsOutput
		for _, op := range deps.Registry {
			if tag != "" && strings.ToLower(op.Tag) != tag {
				continue
			}
			if query != "" &&
				!strings.Contains(strings.ToLower(op.OperationID), query) &&
				!strings.Contains(strings.ToLower(op.Summary), query) {
				continue
			}
			out.Operations = append(out.Operations, OperationSummary{
				OperationID: op.OperationID,
				Method:      op.Method,
				Path:        op.Path,
				Tag:         op.Tag,
				Summary:     op.Summary,
			})
		}
		out.Count = len(out.Operations)
		return nil, out, nil
	}
}

type DescribeOperationInput struct {
	OperationID string `json:"operation_id" jsonschema:"the operationId to describe, from dojo_list_operations"`
}

type DescribeOperationOutput struct {
	OperationID       string `json:"operation_id"`
	Method            string `json:"method"`
	Path              string `json:"path"`
	Tag               string `json:"tag"`
	Summary           string `json:"summary,omitempty"`
	Parameters        any    `json:"parameters,omitempty" jsonschema:"the operation's parameters, straight from the pinned OpenAPI document"`
	RequestBodySchema any    `json:"request_body_schema,omitempty" jsonschema:"the operation's request body schema, if it takes one"`
	ResponseSchema    any    `json:"response_schema,omitempty" jsonschema:"the operation's response schema"`
}

func describeOperationHandler(deps *Deps) mcp.ToolHandlerFor[DescribeOperationInput, DescribeOperationOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in DescribeOperationInput) (*mcp.CallToolResult, DescribeOperationOutput, error) {
		op, ok := findOperation(deps, in.OperationID)
		if !ok {
			return nil, DescribeOperationOutput{}, fmt.Errorf("unknown operation_id %q; use dojo_list_operations to find valid ids", in.OperationID)
		}
		return nil, DescribeOperationOutput{
			OperationID:       op.OperationID,
			Method:            op.Method,
			Path:              op.Path,
			Tag:               op.Tag,
			Summary:           op.Summary,
			Parameters:        rawJSONToAny(op.Parameters),
			RequestBodySchema: rawJSONToAny(op.RequestBodySchema),
			ResponseSchema:    rawJSONToAny(op.ResponseSchema),
		}, nil
	}
}

type CallOperationInput struct {
	OperationID string            `json:"operation_id" jsonschema:"the operationId to call, from dojo_list_operations"`
	PathParams  map[string]string `json:"path_params,omitempty" jsonschema:"values for any {param} placeholders in the operation's path, e.g. {\"id\": \"123\"}"`
	QueryParams map[string]string `json:"query_params,omitempty" jsonschema:"query string parameters for the operation"`
}

func callOperationHandler(deps *Deps) mcp.ToolHandlerFor[CallOperationInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in CallOperationInput) (*mcp.CallToolResult, any, error) {
		op, ok := findOperation(deps, in.OperationID)
		if !ok {
			return nil, nil, fmt.Errorf("unknown operation_id %q; use dojo_list_operations to find valid ids", in.OperationID)
		}
		if err := safety.AllowMethod(op.Method, deps.EnableDestructive); err != nil {
			return nil, nil, err
		}

		auth, err := authHeader(req, deps)
		if err != nil {
			return nil, nil, err
		}

		path := substitutePathParams(op.Path, in.PathParams)
		q := url.Values{}
		for k, v := range in.QueryParams {
			q.Set(k, v)
		}

		var result any
		if err := deps.Client.Do(ctx, op.Method, auth, path, q, nil, &result); err != nil {
			return nil, nil, err
		}

		if capped, truncated := capResponseSize(result); truncated {
			return nil, capped, nil
		}
		return nil, result, nil
	}
}

// rawJSONToAny decodes a json.RawMessage into a plain Go value, so the
// SDK's schema reflection sees "arbitrary JSON" rather than treating the
// underlying []byte as an array of small integers.
func rawJSONToAny(raw json.RawMessage) any {
	if len(raw) == 0 {
		return nil
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return string(raw)
	}
	return v
}

func findOperation(deps *Deps, operationID string) (registry.OperationEntry, bool) {
	for _, op := range deps.Registry {
		if op.OperationID == operationID {
			return op, true
		}
	}
	return registry.OperationEntry{}, false
}

func substitutePathParams(path string, params map[string]string) string {
	for k, v := range params {
		path = strings.ReplaceAll(path, "{"+k+"}", v)
	}
	return path
}

func capResponseSize(v any) (any, bool) {
	b, err := json.Marshal(v)
	if err != nil || len(b) <= maxCallOperationResponseBytes {
		return v, false
	}
	return map[string]any{
		"truncated": true,
		"note": fmt.Sprintf(
			"response was %d bytes, over the %d-byte cap; narrow the request (path/query params, or a curated tool's built-in pagination) instead of requesting the full payload",
			len(b), maxCallOperationResponseBytes,
		),
	}, true
}
