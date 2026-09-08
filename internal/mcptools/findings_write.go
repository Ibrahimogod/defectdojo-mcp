package mcptools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerFindingsWriteTools(server *mcp.Server, deps *Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "dojo_update_finding_status",
		Description: "Set any subset of a DefectDojo finding's triage status fields in one call: active, verified, false positive, duplicate, out of scope, risk accepted, mitigated, severity, mitigation notes. Only the fields you set are changed.",
	}, updateFindingStatusHandler(deps))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "dojo_close_finding",
		Description: "Close a DefectDojo finding (marks it mitigated). Optionally attach a closing note.",
	}, closeFindingHandler(deps))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "dojo_verify_finding",
		Description: "Mark a DefectDojo finding as verified. Optionally attach a note.",
	}, verifyFindingHandler(deps))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "dojo_add_finding_note",
		Description: "Add a note to a DefectDojo finding.",
	}, addFindingNoteHandler(deps))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "dojo_tag_finding",
		Description: "Add tags to a DefectDojo finding (existing tags are kept).",
	}, tagFindingHandler(deps))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "dojo_untag_finding",
		Description: "Remove tags from a DefectDojo finding.",
	}, untagFindingHandler(deps))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "dojo_set_finding_metadata",
		Description: "Set a custom metadata key/value pair on a DefectDojo finding, creating or updating it as needed.",
	}, setFindingMetadataHandler(deps))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "dojo_mark_duplicate",
		Description: "Mark a DefectDojo finding as a duplicate of another finding.",
	}, markDuplicateHandler(deps))
}

type UpdateFindingStatusInput struct {
	ID           int    `json:"id" jsonschema:"the finding id"`
	Active       *bool  `json:"active,omitempty" jsonschema:"set active status"`
	Verified     *bool  `json:"verified,omitempty" jsonschema:"set verified status"`
	FalseP       *bool  `json:"false_positive,omitempty" jsonschema:"set false-positive status"`
	Duplicate    *bool  `json:"duplicate,omitempty" jsonschema:"set duplicate status"`
	OutOfScope   *bool  `json:"out_of_scope,omitempty" jsonschema:"set out-of-scope status"`
	RiskAccepted *bool  `json:"risk_accepted,omitempty" jsonschema:"set risk-accepted status (prefer dojo_accept_finding_risk to also record who accepted it and why)"`
	IsMitigated  *bool  `json:"is_mitigated,omitempty" jsonschema:"set mitigated status"`
	Severity     string `json:"severity,omitempty" jsonschema:"set severity: Critical, High, Medium, Low, Info"`
	Mitigation   string `json:"mitigation,omitempty" jsonschema:"set the mitigation text"`
}

func updateFindingStatusHandler(deps *Deps) mcp.ToolHandlerFor[UpdateFindingStatusInput, map[string]any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in UpdateFindingStatusInput) (*mcp.CallToolResult, map[string]any, error) {
		auth, err := authHeader(req, deps)
		if err != nil {
			return nil, nil, err
		}

		fields := map[string]any{}
		if in.Active != nil {
			fields["active"] = *in.Active
		}
		if in.Verified != nil {
			fields["verified"] = *in.Verified
		}
		if in.FalseP != nil {
			fields["false_p"] = *in.FalseP
		}
		if in.Duplicate != nil {
			fields["duplicate"] = *in.Duplicate
		}
		if in.OutOfScope != nil {
			fields["out_of_scope"] = *in.OutOfScope
		}
		if in.RiskAccepted != nil {
			fields["risk_accepted"] = *in.RiskAccepted
		}
		if in.IsMitigated != nil {
			fields["is_mitigated"] = *in.IsMitigated
		}
		if in.Severity != "" {
			fields["severity"] = in.Severity
		}
		if in.Mitigation != "" {
			fields["mitigation"] = in.Mitigation
		}
		if len(fields) == 0 {
			return nil, nil, fmt.Errorf("no fields to update: set at least one of active, verified, false_positive, duplicate, out_of_scope, risk_accepted, is_mitigated, severity, mitigation")
		}

		body, err := json.Marshal(fields)
		if err != nil {
			return nil, nil, fmt.Errorf("encoding request: %w", err)
		}

		var full map[string]any
		path := fmt.Sprintf("/api/v2/findings/%d/", in.ID)
		if err := deps.Client.Do(ctx, "PATCH", auth, path, nil, body, &full); err != nil {
			return nil, nil, err
		}
		return nil, full, nil
	}
}

type CloseFindingInput struct {
	ID   int    `json:"id" jsonschema:"the finding id"`
	Note string `json:"note,omitempty" jsonschema:"an optional closing note"`
}

func closeFindingHandler(deps *Deps) mcp.ToolHandlerFor[CloseFindingInput, map[string]any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in CloseFindingInput) (*mcp.CallToolResult, map[string]any, error) {
		auth, err := authHeader(req, deps)
		if err != nil {
			return nil, nil, err
		}
		body := map[string]any{"is_mitigated": true}
		if in.Note != "" {
			body["note"] = in.Note
		}
		b, err := json.Marshal(body)
		if err != nil {
			return nil, nil, fmt.Errorf("encoding request: %w", err)
		}
		var result map[string]any
		path := fmt.Sprintf("/api/v2/findings/%d/close/", in.ID)
		if err := deps.Client.Do(ctx, "POST", auth, path, nil, b, &result); err != nil {
			return nil, nil, err
		}
		return nil, result, nil
	}
}

type VerifyFindingInput struct {
	ID   int    `json:"id" jsonschema:"the finding id"`
	Note string `json:"note,omitempty" jsonschema:"an optional note"`
}

func verifyFindingHandler(deps *Deps) mcp.ToolHandlerFor[VerifyFindingInput, map[string]any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in VerifyFindingInput) (*mcp.CallToolResult, map[string]any, error) {
		auth, err := authHeader(req, deps)
		if err != nil {
			return nil, nil, err
		}
		var body []byte
		if in.Note != "" {
			b, err := json.Marshal(map[string]any{"note": in.Note})
			if err != nil {
				return nil, nil, fmt.Errorf("encoding request: %w", err)
			}
			body = b
		}
		var result map[string]any
		path := fmt.Sprintf("/api/v2/findings/%d/verify/", in.ID)
		if err := deps.Client.Do(ctx, "POST", auth, path, nil, body, &result); err != nil {
			return nil, nil, err
		}
		return nil, result, nil
	}
}

type AddFindingNoteInput struct {
	ID      int    `json:"id" jsonschema:"the finding id"`
	Entry   string `json:"entry" jsonschema:"the note text"`
	Private bool   `json:"private,omitempty" jsonschema:"if true, only you and superusers can see this note; it's also left out of reports and issue-tracker sync"`
}

func addFindingNoteHandler(deps *Deps) mcp.ToolHandlerFor[AddFindingNoteInput, map[string]any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in AddFindingNoteInput) (*mcp.CallToolResult, map[string]any, error) {
		auth, err := authHeader(req, deps)
		if err != nil {
			return nil, nil, err
		}
		body, err := json.Marshal(map[string]any{"entry": in.Entry, "private": in.Private})
		if err != nil {
			return nil, nil, fmt.Errorf("encoding request: %w", err)
		}
		var result map[string]any
		path := fmt.Sprintf("/api/v2/findings/%d/notes/", in.ID)
		if err := deps.Client.Do(ctx, "POST", auth, path, nil, body, &result); err != nil {
			return nil, nil, err
		}
		return nil, result, nil
	}
}

type TagFindingInput struct {
	ID   int      `json:"id" jsonschema:"the finding id"`
	Tags []string `json:"tags" jsonschema:"tags to add"`
}

func tagFindingHandler(deps *Deps) mcp.ToolHandlerFor[TagFindingInput, map[string]any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in TagFindingInput) (*mcp.CallToolResult, map[string]any, error) {
		if len(in.Tags) == 0 {
			return nil, nil, fmt.Errorf("tags must not be empty")
		}
		auth, err := authHeader(req, deps)
		if err != nil {
			return nil, nil, err
		}
		body, err := json.Marshal(map[string]any{"tags": in.Tags})
		if err != nil {
			return nil, nil, fmt.Errorf("encoding request: %w", err)
		}
		var result map[string]any
		path := fmt.Sprintf("/api/v2/findings/%d/tags/", in.ID)
		if err := deps.Client.Do(ctx, "POST", auth, path, nil, body, &result); err != nil {
			return nil, nil, err
		}
		return nil, result, nil
	}
}

// UntagOutput: DefectDojo returns 204 with no body on success.
type UntagOutput struct {
	OK bool `json:"ok"`
}

func untagFindingHandler(deps *Deps) mcp.ToolHandlerFor[TagFindingInput, UntagOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in TagFindingInput) (*mcp.CallToolResult, UntagOutput, error) {
		if len(in.Tags) == 0 {
			return nil, UntagOutput{}, fmt.Errorf("tags must not be empty")
		}
		auth, err := authHeader(req, deps)
		if err != nil {
			return nil, UntagOutput{}, err
		}
		body, err := json.Marshal(map[string]any{"tags": in.Tags})
		if err != nil {
			return nil, UntagOutput{}, fmt.Errorf("encoding request: %w", err)
		}
		path := fmt.Sprintf("/api/v2/findings/%d/remove_tags/", in.ID)
		if err := deps.Client.Do(ctx, "PATCH", auth, path, nil, body, nil); err != nil {
			return nil, UntagOutput{}, err
		}
		return nil, UntagOutput{OK: true}, nil
	}
}

type SetFindingMetadataInput struct {
	ID    int    `json:"id" jsonschema:"the finding id"`
	Name  string `json:"name" jsonschema:"the metadata key"`
	Value string `json:"value" jsonschema:"the metadata value"`
}

func setFindingMetadataHandler(deps *Deps) mcp.ToolHandlerFor[SetFindingMetadataInput, map[string]any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in SetFindingMetadataInput) (*mcp.CallToolResult, map[string]any, error) {
		auth, err := authHeader(req, deps)
		if err != nil {
			return nil, nil, err
		}

		// The metadata endpoint needs POST to create a new key and PUT to
		// update an existing one; check which this is first.
		var existing []map[string]any
		listPath := fmt.Sprintf("/api/v2/findings/%d/metadata/", in.ID)
		if err := deps.Client.Get(ctx, auth, listPath, nil, &existing); err != nil {
			return nil, nil, fmt.Errorf("checking existing metadata: %w", err)
		}
		method := "POST"
		for _, m := range existing {
			if name, _ := m["name"].(string); name == in.Name {
				method = "PUT"
				break
			}
		}

		body, err := json.Marshal(map[string]any{"name": in.Name, "value": in.Value})
		if err != nil {
			return nil, nil, fmt.Errorf("encoding request: %w", err)
		}
		var result map[string]any
		if err := deps.Client.Do(ctx, method, auth, listPath, nil, body, &result); err != nil {
			return nil, nil, err
		}
		return nil, result, nil
	}
}

type MarkDuplicateInput struct {
	ID         int `json:"id" jsonschema:"the finding to mark as a duplicate"`
	OriginalID int `json:"original_id" jsonschema:"the finding it's a duplicate of"`
}

// MarkDuplicateOutput: the DefectDojo endpoint returns 204 with no body on
// success, so there's nothing to echo back beyond confirming it worked.
type MarkDuplicateOutput struct {
	OK bool `json:"ok"`
}

func markDuplicateHandler(deps *Deps) mcp.ToolHandlerFor[MarkDuplicateInput, MarkDuplicateOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in MarkDuplicateInput) (*mcp.CallToolResult, MarkDuplicateOutput, error) {
		auth, err := authHeader(req, deps)
		if err != nil {
			return nil, MarkDuplicateOutput{}, err
		}
		path := fmt.Sprintf("/api/v2/findings/%d/original/%d/", in.ID, in.OriginalID)
		if err := deps.Client.Do(ctx, "POST", auth, path, nil, nil, nil); err != nil {
			return nil, MarkDuplicateOutput{}, err
		}
		return nil, MarkDuplicateOutput{OK: true}, nil
	}
}
