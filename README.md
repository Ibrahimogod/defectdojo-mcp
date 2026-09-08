# DefectDojo MCP Server

An [MCP](https://modelcontextprotocol.io) server that lets an LLM (Claude or
any other MCP client) search, triage, and resolve findings in a
[DefectDojo](https://github.com/DefectDojo/django-DefectDojo) instance: mark
false positives, accept risk, verify fixes, close findings, add notes,
manage tags. It also reaches the rest of the DefectDojo API through a small
set of generic discovery tools. Written in Go, runs as a container your MCP
client launches per session over stdio, the same way you'd run any other
containerized MCP server.

See [docs/DESIGN.md](docs/DESIGN.md) for the design rationale and tool
inventory.

## Status

Search and inspect findings, products, engagements, and tests, plus the full
triage/write path: update status, close, verify, accept risk, add notes,
manage tags and metadata, mark duplicates, close/reopen engagements. Raw
`DELETE` operations stay off by default, reachable only through the generic
dispatch tool and only when `DOJO_MCP_ENABLE_DESTRUCTIVE=true` is set. Open
[issues](https://github.com/Ibrahimogod/defectdojo-mcp/issues) track what's
next: scans, JIRA integration, reporting, and further hardening.

## Running it

Requires a DefectDojo instance and a DefectDojo API token (`Authorization:
Token <key>`, from your DefectDojo user profile). Your MCP client launches
the container for you; the shape is always the same:

```bash
docker run -i --rm \
  -e DOJO_BASE_URL=https://defectdojo.example.com \
  -e DOJO_API_TOKEN=your-defectdojo-api-token \
  ghcr.io/ibrahimogod/defectdojo-mcp
```

### Connecting an AI client

**Claude Desktop**: add this to `claude_desktop_config.json`
(`%APPDATA%\Claude\claude_desktop_config.json` on Windows,
`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS),
then restart Claude Desktop:

```json
{
  "mcpServers": {
    "defectdojo": {
      "command": "docker",
      "args": ["run", "--init", "-i", "--rm", "-e", "DOJO_BASE_URL", "-e", "DOJO_API_TOKEN", "ghcr.io/ibrahimogod/defectdojo-mcp"],
      "env": {
        "DOJO_BASE_URL": "https://defectdojo.example.com",
        "DOJO_API_TOKEN": "your-defectdojo-api-token"
      }
    }
  }
}
```

**Claude Code (CLI)**:

```bash
claude mcp add defectdojo \
  -e DOJO_BASE_URL=https://defectdojo.example.com \
  -e DOJO_API_TOKEN=your-defectdojo-api-token \
  -- docker run -i --rm -e DOJO_BASE_URL -e DOJO_API_TOKEN ghcr.io/ibrahimogod/defectdojo-mcp
```

Cursor, VS Code (Copilot), Windsurf, Zed, Cline, Continue.dev, JetBrains AI
Assistant, Gemini CLI, or anything else: see
[docs/CLIENT_SETUP.md](docs/CLIENT_SETUP.md) for exact config.

### Running it as a standing service instead

For a shared deployment reachable by more than one person, set
`DOJO_MCP_TRANSPORT=http` and see
[docs/CLIENT_SETUP.md's HTTP section](docs/CLIENT_SETUP.md#running-as-a-standing-http-service-instead)
and [docs/DESIGN.md §7](docs/DESIGN.md#7-auth-model-detail) for that mode's
auth model. `docker-compose.yml` in this repo is set up for that case:

```bash
cp .env.example .env
# edit .env: set DOJO_BASE_URL to your DefectDojo instance
docker compose up --build
```

## Building from source

The `Makefile` targets (`generate`, `build`, `test`, `lint`, `docker-build`,
`refresh-schema`) are the canonical entry points and what CI runs. `make`
isn't always available on a Windows dev box, so the same steps work directly
with `go` too:

```bash
go generate ./...
go build -o bin/defectdojo-mcp ./cmd/server
go test ./...
go vet ./...
docker build -t defectdojo-mcp .
```

`make refresh-schema` re-fetches the live schema from `$DOJO_BASE_URL` and
diffs it against the pinned copy in `openapi/`. On Windows without `make`,
run `scripts/refresh-schema.ps1` directly; it does the same thing.

## License

[MIT](LICENSE)
