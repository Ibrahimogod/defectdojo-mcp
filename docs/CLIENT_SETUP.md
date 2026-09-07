# Connecting AI clients to this server

The server speaks MCP over **Streamable HTTP** at `http://<host>:<port>/mcp`
(default port `8080`). Every client below needs two things:

1. That URL (`http://localhost:8080/mcp` for a local `docker compose up`, or
   wherever you've deployed the container).
2. Your own DefectDojo API token, sent as the header `Authorization: Token
   <your-defectdojo-api-key>`. The DefectDojo auth scheme uses the word
   `Token`, not `Bearer`; most client examples online say `Bearer` because
   that's what most other APIs use, so don't copy that part over. Get your
   token from your DefectDojo user profile. Per
   [DESIGN.md §7](DESIGN.md#7-auth-model-detail), this server passes that
   header straight through to DefectDojo on each request and never stores a
   credential of its own, so whatever your DefectDojo account is allowed to
   do is what the tools can do, and DefectDojo's own audit log shows your
   name rather than a shared bot account.

**Not usable yet.** The `/mcp` endpoint itself is planned but not yet
implemented (see [DESIGN.md §8](DESIGN.md#8-phased-delivery-plan); Phase 0
only ships `/healthz`). The steps below are what to run once a Phase 1+
build is deployed. Until then, connecting will fail to find any tools.

Client configuration formats change often. If something below doesn't match
what you see in your client, check that client's own current MCP docs
instead of assuming this file is wrong.

Jump to: [Claude Code](#claude-code-cli) · [Claude Desktop](#claude-desktop--claudeai) ·
[Cursor](#cursor) · [VS Code](#vs-code-github-copilot-chat) ·
[Windsurf](#windsurf) · [Zed](#zed) · [Cline](#cline-vs-code-extension) ·
[Continue.dev](#continuedev) · [JetBrains AI Assistant](#jetbrains-ai-assistant-20261) ·
[Gemini CLI](#gemini-cli) · [Any other client](#any-other-mcp-client)

## Claude Code (CLI)

```bash
claude mcp add --transport http defectdojo http://localhost:8080/mcp \
  --header "Authorization: Token YOUR_DEFECTDOJO_TOKEN"
```

`--scope project` writes a `.mcp.json` you can commit so a whole team gets
the same server (each person still needs their own token; don't commit a
real one, see the project-scoped example below). `--scope user` makes it
available in every project instead of just this one (the default scope,
`local`, is just this project for just you).

Verify with `claude mcp list` or `/mcp` inside a Claude Code session.

Project-scoped `.mcp.json`, with the token supplied via environment variable
expansion rather than committed literally:

```json
{
  "mcpServers": {
    "defectdojo": {
      "type": "http",
      "url": "http://localhost:8080/mcp",
      "headers": {
        "Authorization": "Token ${DEFECTDOJO_API_TOKEN}"
      }
    }
  }
}
```

## Claude Desktop / claude.ai

Native custom-connector support for remote MCP servers is OAuth-first. A
static-header auth option ("Request headers") exists in some versions and
plans under Settings → Connectors → Add custom connector → Advanced
settings, but support for it is inconsistent across Claude Desktop versions.
Worth trying first:

1. Settings → Connectors → Add custom connector.
2. Paste `http://localhost:8080/mcp` as the URL.
3. If an "Advanced settings" / "Request headers" option appears, add header
   `Authorization` = `Token YOUR_DEFECTDOJO_TOKEN`.

If that option isn't there on your version, use the
[`mcp-remote`](https://www.npmjs.com/package/mcp-remote) bridge instead.
Claude Desktop still supports local stdio servers through the classic
`claude_desktop_config.json`, and `mcp-remote` forwards a header-authenticated
remote HTTP server through one of those:

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

`claude_desktop_config.json` location: `%APPDATA%\Claude\claude_desktop_config.json`
on Windows, `~/Library/Application Support/Claude/claude_desktop_config.json`
on macOS. Restart Claude Desktop after editing it.

## Cursor

`.cursor/mcp.json` (project) or the global one via Cursor Settings → MCP:

```json
{
  "mcpServers": {
    "defectdojo": {
      "url": "http://localhost:8080/mcp",
      "headers": {
        "Authorization": "Token ${env:DEFECTDOJO_API_TOKEN}"
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
`mcpServers`:

```json
{
  "servers": {
    "defectdojo": {
      "type": "http",
      "url": "http://localhost:8080/mcp",
      "headers": {
        "Authorization": "Token ${input:defectdojo-token}"
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
      "serverUrl": "http://localhost:8080/mcp",
      "headers": {
        "Authorization": "Token YOUR_DEFECTDOJO_TOKEN"
      }
    }
  }
}
```

## Zed

`settings.json`. macOS: `~/.config/zed/settings.json`. Windows:
`%APPDATA%\Zed\settings.json`. Note the top-level key is `context_servers`,
not `mcpServers`:

```json
{
  "context_servers": {
    "defectdojo": {
      "url": "http://localhost:8080/mcp",
      "headers": {
        "Authorization": "Token YOUR_DEFECTDOJO_TOKEN"
      }
    }
  }
}
```

If Zed's direct-URL support doesn't pick up the header correctly on your
version, fall back to the `mcp-remote` bridge, same idea as the Claude
Desktop fallback above:

```json
{
  "context_servers": {
    "defectdojo": {
      "source": "custom",
      "command": "npx",
      "args": [
        "-y", "mcp-remote", "http://localhost:8080/mcp",
        "--header", "Authorization:Token ${DEFECTDOJO_API_TOKEN}"
      ]
    }
  }
}
```

## Cline (VS Code extension)

Edit via Cline's MCP Servers → Configure MCP Servers panel, which opens
`cline_mcp_settings.json` (under the extension's VS Code global storage
directory). Set `type` explicitly; omitting it falls back to the legacy SSE
transport:

```json
{
  "mcpServers": {
    "defectdojo": {
      "type": "streamableHttp",
      "url": "http://localhost:8080/mcp",
      "headers": {
        "Authorization": "Token YOUR_DEFECTDOJO_TOKEN"
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
    type: streamable-http
    url: http://localhost:8080/mcp
    requestOptions:
      headers:
        Authorization: "Token ${{ secrets.DEFECTDOJO_API_TOKEN }}"
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
      "url": "http://localhost:8080/mcp",
      "headers": {
        "Authorization": "Token YOUR_DEFECTDOJO_TOKEN"
      }
    }
  }
}
```

Custom-header support for remote servers is inconsistently documented across
JetBrains IDE versions. If the header isn't picked up, use the same
`mcp-remote` stdio-bridge pattern shown for Claude Desktop above (JetBrains
supports STDIO servers via a `command`/`args` entry the same way).

## Gemini CLI

`~/.gemini/settings.json` (global) or `.gemini/settings.json` (project).
Note the field is `httpUrl`, not `url`:

```json
{
  "mcpServers": {
    "defectdojo": {
      "httpUrl": "http://localhost:8080/mcp",
      "headers": {
        "Authorization": "Token YOUR_DEFECTDOJO_TOKEN"
      }
    }
  }
}
```

## Any other MCP client

If a client isn't listed above, it needs the same three things: a
Streamable-HTTP transport type, the URL `http://<host>:8080/mcp`, and a
custom header `Authorization: Token <your-defectdojo-api-key>`. Consult that
client's own MCP documentation for its specific config shape.

## Verifying the connection independent of any client

[`@modelcontextprotocol/inspector`](https://github.com/modelcontextprotocol/inspector)
can connect directly and list tools without any AI client in the loop, which
is useful for confirming the server itself is reachable before debugging a
client's config:

```bash
npx @modelcontextprotocol/inspector http://localhost:8080/mcp \
  --header "Authorization: Token YOUR_DEFECTDOJO_TOKEN"
```
