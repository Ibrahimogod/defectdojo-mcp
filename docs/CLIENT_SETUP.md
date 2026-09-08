# Connecting AI clients to this server

The default and recommended way to run this server is the same way you'd
run any other containerized MCP server (grafana's, github's, etc.): your
MCP client launches a container per session over stdio, you never manage a
long-running process or a port yourself.

1. A DefectDojo API token, from your DefectDojo user profile.
2. Docker installed.
3. The image: `ghcr.io/ibrahimogod/defectdojo-mcp` (or build it yourself,
   see the main [README](../README.md)).

Search and inspect findings, products, engagements, and tests, plus the full
triage/write path: close a finding, accept risk, verify, add notes, manage
tags, and more. See [DESIGN.md §4](DESIGN.md#4-tool-inventory-roughly-30-curated-tools) for
the full tool list and [DESIGN.md §8](DESIGN.md#8-roadmap) for what's next.

Client configuration formats change often. If something below doesn't match
what you see in your client, check that client's own current MCP docs
instead of assuming this file is wrong.

Jump to: [Claude Code](#claude-code-cli) · [Claude Desktop](#claude-desktop--claudeai) ·
[Cursor](#cursor) · [VS Code](#vs-code-github-copilot-chat) ·
[Windsurf](#windsurf) · [Zed](#zed) · [Cline](#cline-vs-code-extension) ·
[Continue.dev](#continuedev) · [JetBrains AI Assistant](#jetbrains-ai-assistant-20261) ·
[Gemini CLI](#gemini-cli) · [Any other client](#any-other-mcp-client) ·
[Running as a standing HTTP service instead](#running-as-a-standing-http-service-instead)

## Claude Code (CLI)

```bash
claude mcp add defectdojo \
  -e DOJO_BASE_URL=https://defectdojo.example.com \
  -e DOJO_API_TOKEN=YOUR_DEFECTDOJO_TOKEN \
  -- docker run -i --rm -e DOJO_BASE_URL -e DOJO_API_TOKEN ghcr.io/ibrahimogod/defectdojo-mcp
```

`--scope project` writes a `.mcp.json` you can commit so a whole team gets
the same server (each person still needs their own token; don't commit a
real one). `--scope user` makes it available in every project instead of
just this one.

Verify with `claude mcp list` or `/mcp` inside a Claude Code session.

Project-scoped `.mcp.json`, with the token supplied via environment variable
expansion rather than committed literally:

```json
{
  "mcpServers": {
    "defectdojo": {
      "command": "docker",
      "args": ["run", "-i", "--rm", "-e", "DOJO_BASE_URL", "-e", "DOJO_API_TOKEN", "ghcr.io/ibrahimogod/defectdojo-mcp"],
      "env": {
        "DOJO_BASE_URL": "https://defectdojo.example.com",
        "DOJO_API_TOKEN": "${DEFECTDOJO_API_TOKEN}"
      }
    }
  }
}
```

## Claude Desktop / claude.ai

Claude Desktop's native custom-connector UI (Settings → Connectors) is for
*remote* MCP servers reached over HTTP; a container launched over stdio
isn't one of those, so this one goes in the classic
`claude_desktop_config.json` instead: the same file, and the same shape,
your other containerized MCP servers already use there:

```json
{
  "mcpServers": {
    "defectdojo": {
      "command": "docker",
      "args": ["run", "--init", "-i", "--rm", "-e", "DOJO_BASE_URL", "-e", "DOJO_API_TOKEN", "ghcr.io/ibrahimogod/defectdojo-mcp"],
      "env": {
        "DOJO_BASE_URL": "https://defectdojo.example.com",
        "DOJO_API_TOKEN": "YOUR_DEFECTDOJO_TOKEN"
      }
    }
  }
}
```

Location: `%APPDATA%\Claude\claude_desktop_config.json` on Windows,
`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS.
Restart Claude Desktop after editing it.

## Cursor

`.cursor/mcp.json` (project) or the global one via Cursor Settings → MCP:

```json
{
  "mcpServers": {
    "defectdojo": {
      "command": "docker",
      "args": ["run", "-i", "--rm", "-e", "DOJO_BASE_URL", "-e", "DOJO_API_TOKEN", "ghcr.io/ibrahimogod/defectdojo-mcp"],
      "env": {
        "DOJO_BASE_URL": "https://defectdojo.example.com",
        "DOJO_API_TOKEN": "${env:DEFECTDOJO_API_TOKEN}"
      }
    }
  }
}
```

Cursor resolves `${env:VAR}` at connect time. Don't commit a literal token,
and don't commit `.cursor/mcp.json` at all if it would have to contain one
directly.

## VS Code (GitHub Copilot Chat)

`.vscode/mcp.json` (workspace). Note the top-level key is `servers`, not
`mcpServers`, and `"type": "stdio"` has to be explicit:

```json
{
  "servers": {
    "defectdojo": {
      "type": "stdio",
      "command": "docker",
      "args": ["run", "-i", "--rm", "-e", "DOJO_BASE_URL", "-e", "DOJO_API_TOKEN", "ghcr.io/ibrahimogod/defectdojo-mcp"],
      "env": {
        "DOJO_BASE_URL": "https://defectdojo.example.com",
        "DOJO_API_TOKEN": "${input:defectdojo-token}"
      }
    }
  },
  "inputs": [
    {
      "type": "promptString",
      "id": "defectdojo-token",
      "description": "DefectDojo API token",
      "password": true
    }
  ]
}
```

The `inputs` block makes VS Code prompt for the token once (masked) instead
of it ever living in the file.

## Windsurf

`%USERPROFILE%\.codeium\windsurf\mcp_config.json` on Windows
(`~/.codeium/windsurf/mcp_config.json` on macOS/Linux):

```json
{
  "mcpServers": {
    "defectdojo": {
      "command": "docker",
      "args": ["run", "-i", "--rm", "-e", "DOJO_BASE_URL", "-e", "DOJO_API_TOKEN", "ghcr.io/ibrahimogod/defectdojo-mcp"],
      "env": {
        "DOJO_BASE_URL": "https://defectdojo.example.com",
        "DOJO_API_TOKEN": "YOUR_DEFECTDOJO_TOKEN"
      }
    }
  }
}
```

## Zed

`settings.json`. macOS: `~/.config/zed/settings.json`. Windows:
`%APPDATA%\Zed\settings.json`. Note the top-level key is `context_servers`,
not `mcpServers`, and a command-based entry needs `"source": "custom"`:

```json
{
  "context_servers": {
    "defectdojo": {
      "source": "custom",
      "command": "docker",
      "args": ["run", "-i", "--rm", "-e", "DOJO_BASE_URL", "-e", "DOJO_API_TOKEN", "ghcr.io/ibrahimogod/defectdojo-mcp"],
      "env": {
        "DOJO_BASE_URL": "https://defectdojo.example.com",
        "DOJO_API_TOKEN": "YOUR_DEFECTDOJO_TOKEN"
      }
    }
  }
}
```

## Cline (VS Code extension)

Edit via Cline's MCP Servers → Configure MCP Servers panel, which opens
`cline_mcp_settings.json` (under the extension's VS Code global storage
directory):

```json
{
  "mcpServers": {
    "defectdojo": {
      "command": "docker",
      "args": ["run", "-i", "--rm", "-e", "DOJO_BASE_URL", "-e", "DOJO_API_TOKEN", "ghcr.io/ibrahimogod/defectdojo-mcp"],
      "env": {
        "DOJO_BASE_URL": "https://defectdojo.example.com",
        "DOJO_API_TOKEN": "YOUR_DEFECTDOJO_TOKEN"
      }
    }
  }
}
```

## Continue.dev

Continue's config is **YAML**, not JSON (`config.yaml`, global or per-project
under `.continue/`):

```yaml
mcpServers:
  - name: defectdojo
    command: docker
    args:
      - run
      - -i
      - --rm
      - -e
      - DOJO_BASE_URL
      - -e
      - DOJO_API_TOKEN
      - ghcr.io/ibrahimogod/defectdojo-mcp
    env:
      DOJO_BASE_URL: https://defectdojo.example.com
      DOJO_API_TOKEN: ${{ secrets.DEFECTDOJO_API_TOKEN }}
```

Use Continue's own secrets mechanism for the token instead of a literal
value. See the [Continue MCP docs](https://docs.continue.dev/customize/mcp-tools)
for how secrets resolve in your version.

## JetBrains AI Assistant (2026.1+)

Settings → Tools → AI Assistant → Model Context Protocol (MCP) → Add, then
paste:

```json
{
  "mcpServers": {
    "defectdojo": {
      "command": "docker",
      "args": ["run", "-i", "--rm", "-e", "DOJO_BASE_URL", "-e", "DOJO_API_TOKEN", "ghcr.io/ibrahimogod/defectdojo-mcp"],
      "env": {
        "DOJO_BASE_URL": "https://defectdojo.example.com",
        "DOJO_API_TOKEN": "YOUR_DEFECTDOJO_TOKEN"
      }
    }
  }
}
```

## Gemini CLI

`~/.gemini/settings.json` (global) or `.gemini/settings.json` (project):

```json
{
  "mcpServers": {
    "defectdojo": {
      "command": "docker",
      "args": ["run", "-i", "--rm", "-e", "DOJO_BASE_URL", "-e", "DOJO_API_TOKEN", "ghcr.io/ibrahimogod/defectdojo-mcp"],
      "env": {
        "DOJO_BASE_URL": "https://defectdojo.example.com",
        "DOJO_API_TOKEN": "YOUR_DEFECTDOJO_TOKEN"
      }
    }
  }
}
```

## Any other MCP client

If a client isn't listed above, it needs to spawn:

```
docker run -i --rm -e DOJO_BASE_URL -e DOJO_API_TOKEN ghcr.io/ibrahimogod/defectdojo-mcp
```

with `DOJO_BASE_URL` and `DOJO_API_TOKEN` set in that command's environment,
over stdio. Every client that supports local/stdio MCP servers has some
version of a `command`/`args`/`env` config shape for exactly this; consult
that client's own MCP documentation for its specific field names.

## Running as a standing HTTP service instead

For a shared deployment reachable by more than one person from a single
running instance, set `DOJO_MCP_TRANSPORT=http` and run the container as a
persistent service (`docker compose up`, or `docker run -p 8080:8080 ...`,
see the main [README](../README.md)) rather than launching one per client
session. `DOJO_API_TOKEN` becomes optional in this mode: if set, it's used
as a fallback when a request arrives without its own credential; the
primary mechanism is each caller supplying their own DefectDojo token via
the MCP `Authorization: Token <key>` header on every request (see
[DESIGN.md §7](DESIGN.md#7-auth-model-detail)).

Client-side, this swaps the `command`/`args`/`env` block above for a
`url`/`headers` block pointed at `http://<host>:8080/mcp`, in whatever shape
that client uses for remote servers, and field names genuinely differ:
`url`+`headers` (Claude Code, Cursor, VS Code with `"type": "http"`,
Cline with `"type": "streamableHttp"`), `serverUrl` (Windsurf), `httpUrl`
(Gemini CLI), or plain `context_servers.<name>.url` (Zed). The header is
always the same regardless of client: `Authorization: Token
<your-defectdojo-api-key>` (note the word `Token`, not `Bearer`, which is
what most other APIs use and what most client examples online show.

Claude Desktop's native connector UI (Settings → Connectors → Add custom
connector) is built for this mode specifically, though its support for a
static header (versus OAuth) is inconsistent across versions; the
[`mcp-remote`](https://www.npmjs.com/package/mcp-remote) bridge is a
reliable fallback if the native UI doesn't work on yours:

```json
{
  "mcpServers": {
    "defectdojo": {
      "command": "npx",
      "args": [
        "mcp-remote",
        "http://localhost:8080/mcp",
        "--header",
        "Authorization: Token ${DEFECTDOJO_API_TOKEN}"
      ],
      "env": {
        "DEFECTDOJO_API_TOKEN": "YOUR_DEFECTDOJO_TOKEN"
      }
    }
  }
}
```

(That bridge needs Node.js installed for `npx` to exist; if you'd rather
avoid that dependency, the native connector UI is worth trying first.)

## Verifying the connection independent of any client

[`@modelcontextprotocol/inspector`](https://github.com/modelcontextprotocol/inspector)
can connect directly and list tools without any AI client in the loop,
useful for confirming the server itself works before debugging a client's
config. For stdio:

```bash
npx @modelcontextprotocol/inspector docker run -i --rm \
  -e DOJO_BASE_URL=https://defectdojo.example.com \
  -e DOJO_API_TOKEN=YOUR_DEFECTDOJO_TOKEN \
  ghcr.io/ibrahimogod/defectdojo-mcp
```

For HTTP mode:

```bash
npx @modelcontextprotocol/inspector http://localhost:8080/mcp \
  --header "Authorization: Token YOUR_DEFECTDOJO_TOKEN"
```
