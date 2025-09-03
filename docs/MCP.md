# MCP Server (Filesystem)

This repo provides a pluggable MCP server that exposes read-only filesystem context to ChatGPT via JSON-RPC over stdio.

Usage:

- Build: `go build -o bin/mcp ./cmd/mcp`
- Run (default provider `fs`): `./bin/mcp --root $PWD --max-bytes 1048576 --include-hidden=false`

Client config example (ChatGPT Desktop):

{
  "mcpServers": {
    "fs": {
      "command": "/path/to/bin/mcp",
      "args": ["--root", "/home/you/projects/chatbang"],
      "env": {"FS_ROOTS": ""}
    }
  }
}

Flags:
- `--provider fs`: select provider (pluggable registry)
- `--root <path>`: allowlisted root(s), colon or comma separated
- `--max-bytes <n>`: max bytes per read
- `--include-hidden`: include dotfiles
- `--allow-binary`: allow binary reads (otherwise summarized)

Security:
- Read-only, path traversal blocked, symlinks resolved, hidden files optional.

Development:
- Providers register via `internal/mcp.Register(name, Factory)` and implement `Provider`.
