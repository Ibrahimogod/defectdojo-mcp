package mcptools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ibrahimogod/defectdojo-mcp/internal/dojoclient"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// newTestDeps starts an httptest.Server driven by handler and returns Deps
// wired to it with a fixed fallback token, so tool handlers can be called
// directly without a live MCP session (authHeader falls back to
// deps.FallbackAuth when req.Extra is nil, which it is for &mcp.CallToolRequest{}).
func newTestDeps(t *testing.T, handler http.HandlerFunc) *Deps {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return &Deps{
		Client:       dojoclient.New(srv.URL, 5_000_000_000),
		FallbackAuth: "Token test-token",
	}
}

func decodeBody(t *testing.T, r *http.Request) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		t.Fatalf("decoding request body: %v", err)
	}
	return m
}

func TestUpdateFindingStatusHandler(t *testing.T) {
	t.Run("no fields set is an error", func(t *testing.T) {
		deps := newTestDeps(t, func(w http.ResponseWriter, r *http.Request) {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		})
		_, _, err := updateFindingStatusHandler(deps)(context.Background(), &mcp.CallToolRequest{}, UpdateFindingStatusInput{ID: 1})
		if err == nil {
			t.Fatal("expected an error when no fields are set")
		}
	})

	t.Run("only sets provided fields", func(t *testing.T) {
		var gotBody map[string]any
		var gotPath, gotMethod string
		deps := newTestDeps(t, func(w http.ResponseWriter, r *http.Request) {
			gotPath, gotMethod = r.URL.Path, r.Method
			gotBody = decodeBody(t, r)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"id":42}`))
		})

		verified := true
		_, out, err := updateFindingStatusHandler(deps)(context.Background(), &mcp.CallToolRequest{}, UpdateFindingStatusInput{
			ID:       42,
			Verified: &verified,
			Severity: "High",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotMethod != "PATCH" || gotPath != "/api/v2/findings/42/" {
			t.Fatalf("method/path = %s %s, want PATCH /api/v2/findings/42/", gotMethod, gotPath)
		}
		if len(gotBody) != 2 || gotBody["verified"] != true || gotBody["severity"] != "High" {
			t.Fatalf("body = %v, want only verified=true and severity=High", gotBody)
		}
		if out["id"] != float64(42) {
			t.Fatalf("out = %v, want id=42 echoed back", out)
		}
	})
}

func TestUntagFindingHandler(t *testing.T) {
	t.Run("empty tags is an error", func(t *testing.T) {
		deps := newTestDeps(t, func(w http.ResponseWriter, r *http.Request) {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		})
		_, _, err := untagFindingHandler(deps)(context.Background(), &mcp.CallToolRequest{}, TagFindingInput{ID: 1})
		if err == nil {
			t.Fatal("expected an error for empty tags")
		}
	})

	t.Run("204 empty body still reports ok", func(t *testing.T) {
		var gotPath, gotMethod string
		deps := newTestDeps(t, func(w http.ResponseWriter, r *http.Request) {
			gotPath, gotMethod = r.URL.Path, r.Method
			w.WriteHeader(http.StatusNoContent)
		})
		_, out, err := untagFindingHandler(deps)(context.Background(), &mcp.CallToolRequest{}, TagFindingInput{ID: 7, Tags: []string{"stale"}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotMethod != "PATCH" || gotPath != "/api/v2/findings/7/remove_tags/" {
			t.Fatalf("method/path = %s %s, want PATCH /api/v2/findings/7/remove_tags/", gotMethod, gotPath)
		}
		if !out.OK {
			t.Fatalf("out.OK = false, want true despite empty 204 response")
		}
	})
}

func TestMarkDuplicateHandler(t *testing.T) {
	var gotPath, gotMethod string
	deps := newTestDeps(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotMethod = r.URL.Path, r.Method
		w.WriteHeader(http.StatusNoContent)
	})
	_, out, err := markDuplicateHandler(deps)(context.Background(), &mcp.CallToolRequest{}, MarkDuplicateInput{ID: 5, OriginalID: 9})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != "POST" || gotPath != "/api/v2/findings/5/original/9/" {
		t.Fatalf("method/path = %s %s, want POST /api/v2/findings/5/original/9/", gotMethod, gotPath)
	}
	if !out.OK {
		t.Fatalf("out.OK = false, want true despite empty 204 response")
	}
}

func TestSetFindingMetadataHandler(t *testing.T) {
	t.Run("creates with POST when key is new", func(t *testing.T) {
		var gotMethod string
		deps := newTestDeps(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if r.Method == http.MethodGet {
				w.Write([]byte(`[]`))
				return
			}
			gotMethod = r.Method
			w.Write([]byte(`{"name":"env","value":"prod"}`))
		})
		_, _, err := setFindingMetadataHandler(deps)(context.Background(), &mcp.CallToolRequest{}, SetFindingMetadataInput{ID: 1, Name: "env", Value: "prod"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotMethod != "POST" {
			t.Fatalf("method = %s, want POST for a new metadata key", gotMethod)
		}
	})

	t.Run("updates with PUT when key exists", func(t *testing.T) {
		var gotMethod string
		deps := newTestDeps(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if r.Method == http.MethodGet {
				w.Write([]byte(`[{"name":"env","value":"staging"}]`))
				return
			}
			gotMethod = r.Method
			w.Write([]byte(`{"name":"env","value":"prod"}`))
		})
		_, _, err := setFindingMetadataHandler(deps)(context.Background(), &mcp.CallToolRequest{}, SetFindingMetadataInput{ID: 1, Name: "env", Value: "prod"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotMethod != "PUT" {
			t.Fatalf("method = %s, want PUT for an existing metadata key", gotMethod)
		}
	})
}

func TestAcceptFindingRiskHandler(t *testing.T) {
	t.Run("empty finding_ids is an error", func(t *testing.T) {
		deps := newTestDeps(t, func(w http.ResponseWriter, r *http.Request) {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		})
		_, _, err := acceptFindingRiskHandler(deps)(context.Background(), &mcp.CallToolRequest{}, AcceptFindingRiskInput{Name: "x"})
		if err == nil {
			t.Fatal("expected an error for empty finding_ids")
		}
	})

	t.Run("resolves owner then posts risk acceptance", func(t *testing.T) {
		var gotBody map[string]any
		var gotPath, gotMethod string
		deps := newTestDeps(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if r.URL.Path == "/api/v2/user_profile/" {
				w.Write([]byte(`{"user":{"id":11}}`))
				return
			}
			gotPath, gotMethod = r.URL.Path, r.Method
			gotBody = decodeBody(t, r)
			w.Write([]byte(`{"id":99}`))
		})

		_, out, err := acceptFindingRiskHandler(deps)(context.Background(), &mcp.CallToolRequest{}, AcceptFindingRiskInput{
			FindingIDs:    []int{1, 2},
			Name:          "accepted for Q1",
			Justification: "compensating control in place",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotMethod != "POST" || gotPath != "/api/v2/risk_acceptance/" {
			t.Fatalf("method/path = %s %s, want POST /api/v2/risk_acceptance/", gotMethod, gotPath)
		}
		if gotBody["owner"] != float64(11) {
			t.Fatalf("owner = %v, want 11 resolved from /user_profile/", gotBody["owner"])
		}
		if gotBody["decision"] != "A" || gotBody["decision_details"] != "compensating control in place" {
			t.Fatalf("body = %v, missing expected decision fields", gotBody)
		}
		if out["id"] != float64(99) {
			t.Fatalf("out = %v, want id=99 echoed back", out)
		}
	})
}

func TestEngagementActionHandlers(t *testing.T) {
	t.Run("close engagement handles empty 200 body", func(t *testing.T) {
		var gotPath, gotMethod string
		deps := newTestDeps(t, func(w http.ResponseWriter, r *http.Request) {
			gotPath, gotMethod = r.URL.Path, r.Method
			w.WriteHeader(http.StatusOK)
		})
		_, out, err := closeEngagementHandler(deps)(context.Background(), &mcp.CallToolRequest{}, IDInput{ID: 3})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotMethod != "POST" || gotPath != "/api/v2/engagements/3/close/" {
			t.Fatalf("method/path = %s %s, want POST /api/v2/engagements/3/close/", gotMethod, gotPath)
		}
		if !out.OK {
			t.Fatalf("out.OK = false, want true despite empty 200 response")
		}
	})

	t.Run("reopen engagement handles empty 200 body", func(t *testing.T) {
		var gotPath, gotMethod string
		deps := newTestDeps(t, func(w http.ResponseWriter, r *http.Request) {
			gotPath, gotMethod = r.URL.Path, r.Method
			w.WriteHeader(http.StatusOK)
		})
		_, out, err := reopenEngagementHandler(deps)(context.Background(), &mcp.CallToolRequest{}, IDInput{ID: 3})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotMethod != "POST" || gotPath != "/api/v2/engagements/3/reopen/" {
			t.Fatalf("method/path = %s %s, want POST /api/v2/engagements/3/reopen/", gotMethod, gotPath)
		}
		if !out.OK {
			t.Fatalf("out.OK = false, want true despite empty 200 response")
		}
	})
}

func TestAuthHeaderFallback(t *testing.T) {
	t.Run("uses fallback when request carries no header", func(t *testing.T) {
		deps := &Deps{FallbackAuth: "Token fallback-token"}
		got, err := authHeader(&mcp.CallToolRequest{}, deps)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "Token fallback-token" {
			t.Fatalf("authHeader = %q, want fallback token", got)
		}
	})

	t.Run("errors when no header and no fallback are available", func(t *testing.T) {
		deps := &Deps{}
		if _, err := authHeader(&mcp.CallToolRequest{}, deps); err == nil {
			t.Fatal("expected an error when no credential is available")
		}
	})
}
