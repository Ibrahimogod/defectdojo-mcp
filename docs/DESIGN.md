# DefectDojo MCP Server: Design & Roadmap

Design reference for the project: why it's built the way it is, what the
tools look like, and what's not built yet.

## 0. What this is, in one paragraph

An MCP (Model Context Protocol) server that sits in front of a DefectDojo
instance (API v2, OpenAPI 3.0.3; see `openapi/` for the pinned schema this
project targets) and lets an LLM search, triage, and resolve security
findings: mark false positives, accept risk, verify fixes, close findings,
add notes, manage tags. It also reaches every other operation the DefectDojo
API exposes, without hand-writing 190+ one-off tool wrappers. Runs as a
container over stdio by default, the same way any other containerized MCP
server does; a Streamable HTTP mode is available too for a shared
deployment reachable by more than one person.

## 1. Why these choices

| Decision | Choice | Why |
|---|---|---|
| Language | **Go** | This workload is I/O-bound on DefectDojo's own response times, not compute-bound. Rust's runtime edge doesn't show up here, but its borrow-checker overhead does slow delivery. Go gives a static binary, a ~10-15MB distroless image, sub-second cold start, and a solid official MCP SDK (`github.com/modelcontextprotocol/go-sdk`) with much less ceremony than Rust's `rmcp`. |
| Transport | **stdio by default, Streamable HTTP as an option** | Every other MCP server in a typical setup (grafana's, github's) runs as a container the client launches per session over stdio, and matching that is simpler for the common case than requiring a standing service. HTTP stays available (`DOJO_MCP_TRANSPORT=http`) for a shared/team deployment reachable by more than one person from one running instance. Both share the same tool-handling code; only the transport adapter differs. |
| API coverage | **Curated tools + generic dispatch**, not 190+ 1:1 tool wrappers | Exposing every DefectDojo operation as its own MCP tool would go past what LLMs can reliably pick from in a tool list. Instead: roughly 30 hand-crafted tools for the actual triage workflow (the 90% use case), plus 3 generic discovery/dispatch tools that can invoke any DefectDojo operation by `operationId`. Full functional coverage, without a 380-entry tool list. See §4. |
| Auth model | **One token per server instance**, matching the transport | For stdio, there's no such thing as "per caller"; each container is one person's own instance, so `DOJO_API_TOKEN` is a required env var, same pattern as `GITHUB_PERSONAL_ACCESS_TOKEN` or `GRAFANA_SERVICE_ACCOUNT_TOKEN` on other containerized MCP servers. For HTTP mode, where one instance really is shared across people, the primary mechanism is per-request passthrough of the caller's own token via the MCP `Authorization` header, with `DOJO_API_TOKEN` as an optional fallback. Either way, DefectDojo's own RBAC and audit log (`mitigated_by`, `reporter`, notes authorship) reflect the real person, not a shared service account. See §7. |
| Destructive ops | **Deny by default** | Raw `DELETE` operations (delete a product, a finding, a user...) are reachable only through the generic dispatch tool, and only when `DOJO_MCP_ENABLE_DESTRUCTIVE=true` is set. No curated tool ever issues a DELETE. `*_delete_preview` endpoints (dry-run of what a delete would remove) are always allowed since they're read-only. |

Open to revisiting any of these if a real constraint shows up. If you're
thinking about swapping the language or transport, open an issue first so
it's discussed rather than silently redone.

## 2. Source of truth for the API

- Schema endpoint (append to your instance's base URL): `/api/v2/oa3/schema/?format=json`
- Swagger UI for browsing: `/api/v2/oa3/swagger-ui/`
- Pinned copy used for code generation: [`openapi/defectdojo-schema.v3.2.100.json`](../openapi/defectdojo-schema.v3.2.100.json)
  (DefectDojo v3.2.100, OpenAPI 3.0.3, 191 paths, 380 operations). Not fetched
  live at build time. It's pinned, with a `make refresh-schema` /
  `scripts/refresh-schema.{sh,ps1}` target that re-fetches from
  `$DOJO_BASE_URL` and diffs it against the pinned copy, so picking up a
  newer DefectDojo schema is a deliberate step, not an accident.
- Auth scheme, confirmed from `components.securitySchemes` in the schema:
  `tokenAuth`, header `Authorization: Token <key>`. DefectDojo's prefix is
  the literal word `Token`, not `Bearer`. `basicAuth` and `cookieAuth`
  (session cookie) exist too but are out of scope; token auth only.
- Pagination: DRF-style `limit`/`offset` query params, plus an `o` ordering
  param on list endpoints. Responses look like `{count, next, previous, results}`.

## 3. Repo layout

```
.
  cmd/
    server/main.go              # entrypoint: config load, wire everything, start HTTP server
  internal/
    config/                     # env-based config struct + validation
    dojoclient/                 # small hand-written REST client, no generated code
      client.go                 # Get/Do against DefectDojo, auth header in, decoded JSON out
    registry/                   # operation registry built from the same OpenAPI doc at build time
      registry.go               # operationId -> {method, path, params, request/response schema}
      registry_gen.go            # go:generate output embedding the pruned schema
    safety/                     # method/operationId allow-deny list, destructive-op gating
    mcptools/                   # one file per tool group, not per tool
      findings.go                # search/get/update/close/verify/accept-risk/notes/tags/metadata
      risk_acceptance.go
      products_engagements_tests.go
      scans.go                   # import-scan / reimport-scan, see #7
      jira.go
      discovery.go                # dojo_list_operations / describe_operation / call_operation
      summarize.go                # shared: trims finding/product/etc JSON down to what's actually
                                   # worth putting in front of an LLM. Both the generic dispatch
                                   # tool and the curated list tools call into this, so truncation
                                   # behavior lives in one place, not several
    mcpserver/                   # stdio and Streamable HTTP transport wiring, auth, health/readyz
    log/                         # structured logging (slog), token redaction
  openapi/
    defectdojo-schema.v3.2.100.json   # pinned schema
  test/
    integration/                 # spins up a fake DefectDojo (httptest) or a real one via docker-compose
  Dockerfile
  docker-compose.yml
  Makefile                       # generate, build, test, lint, docker-build, refresh-schema
  .github/workflows/
```

## 4. Tool inventory (roughly 30 curated tools)

Curated tools return summarized JSON by default: id, title, severity, the
status booleans, product/engagement/test names, SLA dates. Not the full
60+-field DefectDojo finding object, because a list of 100 full findings
would blow the model's context for no real benefit. Full detail is available
through `dojo_get_finding` or the generic dispatch tool when it's actually
needed. This truncation logic lives in `internal/mcptools/summarize.go` and
both the curated tools and `dojo_call_operation` route through it.

Findings, the core workflow:
- `dojo_search_findings`. Filter by severity, active/verified/false_p/duplicate/out_of_scope/risk_accepted/is_mitigated, product name, test, CWE, tags, title substring, discovered date range. Supports `limit`/`offset`/`order`.
- `dojo_get_finding`. Full detail for one finding, optionally with related findings/notes inline.
- `dojo_update_finding_status`. The one tool for triage. Sets any subset of `active, verified, false_p, duplicate, out_of_scope, risk_accepted, is_mitigated, severity, mitigation` in a single PATCH. This is what an LLM reaches for to "resolve" a finding.
- `dojo_close_finding` / `dojo_verify_finding`. Thin wrappers over `/findings/{id}/close/` and `/findings/{id}/verify/`.
- `dojo_accept_finding_risk`. Creates or attaches a risk acceptance (justification, optional expiration) for one or more finding IDs.
- `dojo_add_finding_note` / `dojo_list_finding_notes`
- `dojo_tag_finding` / `dojo_untag_finding`
- `dojo_get_finding_metadata` / `dojo_set_finding_metadata`
- `dojo_mark_duplicate`. Links a finding as duplicate of another (the `/findings/{id}/original/{new_fid}/` family).
- `dojo_generate_findings_report`. PDF or other report for a filtered finding set.

Risk acceptances:
- `dojo_list_risk_acceptances`, `dojo_expire_risk_acceptance`, `dojo_reinstate_risk_acceptance`.

Products, engagements, tests (context browsing, mostly read):
- `dojo_list_products`, `dojo_get_product`
- `dojo_list_engagements`, `dojo_get_engagement`, `dojo_close_engagement`, `dojo_reopen_engagement`
- `dojo_list_tests`, `dojo_get_test`

Scans (not built yet, [#7](https://github.com/Ibrahimogod/defectdojo-mcp/issues/7)):
- `dojo_import_scan` (engagement + scan_type + file), `dojo_reimport_scan` (test + file).

JIRA (not built yet, [#7](https://github.com/Ibrahimogod/defectdojo-mcp/issues/7), only relevant if a JIRA instance is configured on the DefectDojo side):
- `dojo_push_finding_to_jira`, `dojo_get_finding_jira_mapping`.

Discovery / generic dispatch, the full-coverage fallback:
- `dojo_list_operations(tag?, query?)`. Lists all DefectDojo operations (operationId, method, path, tag, one-line summary) from the embedded registry, optionally filtered.
- `dojo_describe_operation(operationId)`. Returns the full parameter/request-body/response schema for one operation, straight from the pinned OpenAPI doc.
- `dojo_call_operation(operationId, path_params?, query_params?)`. Executes through `dojoclient`, returns a size-capped response. Only GET operations are allowed for now; write support (and the `DOJO_MCP_ENABLE_DESTRUCTIVE`-gated DELETE allowance specifically) lands with the write path.

Every DefectDojo operation not in the curated list (user management,
notifications, SLA configs, endpoints, languages, technologies, network
locations, announcements, system settings, and so on) is reachable through
the three discovery/dispatch tools. That covers full coverage without a
380-entry tool list.

## 5. Non-functional requirements

- Timeouts and retries. Every upstream call carries a context timeout
  (default 30s, configurable). Retry idempotent GETs on 502/503/504 with
  capped exponential backoff, max 3 attempts. Never auto-retry non-idempotent
  verbs.
- Rate limiting. A token-bucket limiter sits in front of the DefectDojo
  client so a chatty LLM loop can't hammer the upstream instance. Returns a
  clear MCP tool error, not a raw 429, when throttled.
- Pagination safety. `dojo_search_findings` and every other list tool cap
  `limit` at a hard max (100, say) no matter what's requested, and always
  return `count`/`next`/`offset` so the model can page deliberately instead
  of trying to pull everything at once.
- Error mapping. DefectDojo 4xx/5xx bodies get translated into structured
  MCP tool errors with the upstream status and message, not a raw stack
  trace.
- Logging. `slog` structured logs. The DefectDojo token (and the MCP
  `Authorization` header) get redacted in every log line and in any panic
  recovery output. A redaction test fails CI if a token shape ever shows up
  in log output.
- Health. `/healthz` (liveness, no upstream call) and `/readyz` (a cheap
  authenticated call to DefectDojo, to confirm the upstream is reachable).
  Both matter for the compose file as much as for whatever orchestrates this
  in production.
- Testing. Unit tests for `safety` (the destructive-op gate is the single
  most important thing to get right here; explicit tests for every DELETE
  operationId being blocked by default and allowed once the env var is set),
  `summarize` (truncation correctness), and `dojoclient` (retry/timeout
  behavior against an `httptest.Server`). Integration tests hit either a
  real DefectDojo (via `docker-compose.yml`, gated behind a build tag / env
  flag so CI doesn't require it) or a schema-validated fake.
- CI. `go vet`, `golangci-lint`, `go test ./...`, `docker build`. Build fails
  if `go generate ./...` produces a diff, which keeps the generated
  client/registry honest against the pinned schema.

## 6. Docker

- Multi-stage `Dockerfile`. Build stage `golang`, `CGO_ENABLED=0 go build`,
  final stage `gcr.io/distroless/static-debian12`, non-root (`USER
  nonroot`), only the static binary in the image. Target under ~20MB.
- `docker-compose.yml`: runs the server in HTTP mode as a standing service,
  configured entirely through environment variables (see `.env.example`).
  No instance-specific defaults baked in, since anyone running this points
  it at their own DefectDojo. The default stdio mode doesn't need Compose
  at all: an MCP client launches the container itself, per session.
- Config entirely via environment variables (12-factor style):
  `DOJO_BASE_URL`, `DOJO_MCP_TRANSPORT` (`stdio` default, or `http`),
  `DOJO_MCP_LISTEN_ADDR` (http mode), `DOJO_MCP_ENABLE_DESTRUCTIVE`,
  `DOJO_API_TOKEN` (required for stdio, optional fallback for http, see
  §7), `DOJO_MCP_REQUEST_TIMEOUT`, `DOJO_MCP_RATE_LIMIT_RPS`,
  `DOJO_MCP_LOG_LEVEL`.
- Published images: `ghcr.io/ibrahimogod/defectdojo-mcp` (see
  `.github/workflows/publish.yml`).

## 7. Auth model detail

**stdio (default):** there's no per-request caller to distinguish since
each container is one person's own instance, launched by their own MCP
client. `DOJO_API_TOKEN` is a required env var; the server uses it for
every upstream call for the life of that container. Same pattern as
`GITHUB_PERSONAL_ACCESS_TOKEN`/`GRAFANA_SERVICE_ACCOUNT_TOKEN` on other
containerized MCP servers. Per-user attribution still holds in DefectDojo's
own history, since each person runs their own container with their own
token.

**HTTP (shared deployment), per-caller passthrough:**
1. The MCP client is set up with this server's URL and an `Authorization:
   Token <the user's own DefectDojo API key>` header. Each user generates
   their own key from their DefectDojo profile.
2. The MCP server's auth middleware extracts that header verbatim per
   request and hands it straight to `dojoclient` for the upstream call. The
   server itself never stores or needs a DefectDojo credential of its own.
3. Anything the tool does (closing a finding, adding a note) shows up in
   DefectDojo's own history as that person. DefectDojo's own RBAC (a
   read-only user can't accept risk, for instance) is enforced by DefectDojo
   itself, for free.

Fallback, `DOJO_API_TOKEN` set in HTTP mode: if an incoming request has no
`Authorization` header, fall back to that single token from config. This
loses per-user attribution and should only be used for a
trusted, single-tenant deployment.

## 8. Roadmap

Scaffold and the read path (search/get findings, products, engagements,
tests, discovery tools) are done. What's left is tracked as issues rather
than spelled out here, since that's where it'll actually get updated as
work happens:

- [#6](https://github.com/Ibrahimogod/defectdojo-mcp/issues/6): triage/write path (update, close, verify, accept risk, notes, tags)
- [#7](https://github.com/Ibrahimogod/defectdojo-mcp/issues/7): scans, JIRA, reporting
- [#8](https://github.com/Ibrahimogod/defectdojo-mcp/issues/8): hardening (rate limiting, retries, redaction test, integration tests)

Roughly in that order, but not a hard commitment; whatever's most useful
next wins.

## 9. Explicit non-goals (v1)

- Not a general-purpose OpenAPI-to-MCP framework for other APIs. This is a
  DefectDojo-specific server, even though the registry/dispatch pattern
  happens to be generic internally.
- No UI. No caching layer or database. The server is stateless; DefectDojo
  is the source of truth.
- No multi-tenant DefectDojo (more than one base URL per running server
  instance). One server, one DefectDojo instance.
- No write access to users/products/engagements creation-side admin
  operations beyond what's listed in §4. Those stay reachable only via the
  generic dispatch tool, gated the same as anything else.

## 10. Open questions / follow-ups

- Whether a given target DefectDojo instance actually has a JIRA integration
  configured, which affects whether the JIRA tools ([#7](https://github.com/Ibrahimogod/defectdojo-mcp/issues/7)) are worth testing
  against it.
- Docker network reachability from the host running this container to
  whatever DefectDojo instance it's pointed at. Confirm before assuming a
  container can reach it directly.
- Whether per-user DefectDojo API tokens already exist for the intended
  users, or need to be provisioned first.
