# DefectDojo MCP Server

An [MCP](https://modelcontextprotocol.io) server that lets an LLM (Claude or
any other MCP client) search, triage, and resolve findings in a
[DefectDojo](https://github.com/DefectDojo/django-DefectDojo) instance: mark
false positives, accept risk, verify fixes, close findings, add notes,
manage tags. It also reaches the rest of the DefectDojo API through a small
set of generic discovery tools. Written in Go, runs as a single Docker
container speaking MCP over Streamable HTTP.

See [docs/DESIGN.md](docs/DESIGN.md) for the design rationale and tool
inventory.

## Status

Early scaffold. Open [issues](https://github.com/Ibrahimogod/defectdojo-mcp/issues)
track what's next; the MCP tools themselves aren't wired up yet, so today
the server only exposes `/healthz`.

## Running it

Requires a DefectDojo instance and a DefectDojo API token (`Authorization:
Token <key>`, from your DefectDojo user profile).

```bash
cp .env.example .env
# edit .env: set DOJO_BASE_URL to your DefectDojo instance
docker compose up --build
```

Or without Compose:

```bash
docker run --rm -p 8080:8080 \
  -e DOJO_BASE_URL=https://defectdojo.example.com \
  ghcr.io/ibrahimogod/defectdojo-mcp:latest
```

Point your MCP client at `http://localhost:8080/mcp` with an `Authorization:
Token <your-defectdojo-api-key>` header. See
[docs/DESIGN.md §7](docs/DESIGN.md#7-auth-model-detail) for the auth model,
and [docs/CLIENT_SETUP.md](docs/CLIENT_SETUP.md) for exact steps for Claude
Code, Claude Desktop, Cursor, VS Code (Copilot), Windsurf, and other MCP
clients.

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

Podman works too: `podman build`/`podman run` are drop-in replacements for
the `docker` commands above.

## License

[MIT](LICENSE)
